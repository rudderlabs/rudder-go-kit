package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/testhelper"

	"github.com/fsouza/fake-gcs-server/fakestorage"
)

func TestGCSManager(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := testhelper.GetFreePort()
	require.NoError(t, err)

	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		InitialObjects: []fakestorage.Object{
			{
				ObjectAttrs: fakestorage.ObjectAttrs{
					BucketName: "test-bucket",
					Name:       "test-prefix/testFile",
				},
				Content: []byte("inside the file"),
			},
		},
		Scheme: "http",
		Host:   "127.0.0.1",
		Port:   uint16(port),
	})
	require.NoError(t, err)
	defer server.Stop()

	gcsURL := fmt.Sprintf("%s/storage/v1/", server.URL())
	t.Log("GCS URL:", gcsURL)

	conf := map[string]interface{}{
		"bucketName": "test-bucket",
		"prefix":     "test-prefix",
		"endPoint":   gcsURL,
		"disableSSL": true,
		"jsonReads":  true,
	}
	m, err := New(&Settings{
		Provider: "GCS",
		Config:   conf,
		Logger:   logger.NOP,
		Conf:     config.New(),
	})
	require.NoError(t, err)

	tempDir := t.TempDir()
	f, err := os.Create(tempDir + "/testFile")
	require.NoError(t, err)

	t.Log("pre-existing file")
	uploadedFile, err := m.Upload(ctx, f)
	require.NoError(t, err)
	require.Equal(t, "test-prefix/testFile", uploadedFile.ObjectName)

	t.Run("new file", func(t *testing.T) {
		tempDir := t.TempDir()
		f, err := os.Create(tempDir + "/testFile-new")
		require.NoError(t, err)

		_, err = m.Upload(ctx, f)
		require.NoError(t, err)
	})
}

func TestGCSManagerConcurrentDelete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	port, err := testhelper.GetFreePort()
	require.NoError(t, err)

	// Create a channel to track concurrent operations
	activeOps := make(chan int, 1000)
	maxConcurrent := 0
	currentConcurrent := 0
	var mu sync.Mutex

	// Create fake GCS server
	server, err := fakestorage.NewServerWithOptions(fakestorage.Options{
		Scheme: "http",
		Host:   "127.0.0.1",
		Port:   uint16(port),
	})
	require.NoError(t, err)
	defer server.Stop()

	// port for the custom server with delete handler
	customPort, err := testhelper.GetFreePort()
	require.NoError(t, err)

	fakeServerWithDeleteHandler := &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", customPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodDelete {
				mu.Lock()
				currentConcurrent++
				if currentConcurrent > maxConcurrent {
					maxConcurrent = currentConcurrent
				}
				mu.Unlock()

				activeOps <- 1
				defer func() {
					<-activeOps
					mu.Lock()
					currentConcurrent--
					mu.Unlock()
				}()
			}
			// Create a new request to forward to the GCS server
			forwardURL := fmt.Sprintf("%s%s", server.URL(), r.URL.Path)
			if r.URL.RawQuery != "" {
				forwardURL += "?" + r.URL.RawQuery
			}

			forwardReq, err := http.NewRequestWithContext(r.Context(), r.Method, forwardURL, r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Copy headers
			for k, v := range r.Header {
				forwardReq.Header[k] = v
			}

			resp, err := http.DefaultClient.Do(forwardReq)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer func() {
				_ = resp.Body.Close()
			}()
			for k, v := range resp.Header {
				w.Header()[k] = v
			}
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
		}),
	}

	// Start the server in a goroutine
	go func() {
		if err := fakeServerWithDeleteHandler.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("custom server error: %v", err)
		}
	}()

	defer func() {
		_ = fakeServerWithDeleteHandler.Close()
	}()

	gcsURL := fmt.Sprintf("http://127.0.0.1:%d/storage/v1/", customPort)
	t.Log("GCS URL:", gcsURL)

	// Test different concurrency levels
	concurrencyLevels := []int{1, 10}
	objectCount := 100

	bucketName := fmt.Sprintf("test-bucket-%s", rand.UniqueString(10))
	server.CreateBucketWithOpts(fakestorage.CreateBucketOpts{
		Name: bucketName,
	})

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("concurrency_%d", concurrency), func(t *testing.T) {
			mu.Lock()
			maxConcurrent = 0
			currentConcurrent = 0
			mu.Unlock()

			prefix := fmt.Sprintf("test-prefix-%d", concurrency)

			conf := map[string]interface{}{
				"bucketName":                  bucketName,
				"prefix":                      prefix,
				"endPoint":                    gcsURL,
				"disableSSL":                  true,
				"jsonReads":                   true,
				"concurrentDeleteObjRequests": concurrency,
			}

			m, err := NewGCSManager(conf, logger.NOP, func() time.Duration {
				return 5 * time.Minute
			})
			require.NoError(t, err)

			// Upload test objects
			keys := make([]string, objectCount)
			for i := 0; i < objectCount; i++ {
				objName := fmt.Sprintf("%s/obj-%d.txt", prefix, i)
				content := []byte(fmt.Sprintf("data-%d", i))

				// Upload object to fake server
				server.CreateObject(fakestorage.Object{
					ObjectAttrs: fakestorage.ObjectAttrs{
						BucketName: bucketName,
						Name:       objName,
					},
					Content: content,
				})

				keys[i] = objName
			}

			err = m.Delete(ctx, keys)
			require.NoError(t, err)

			for _, key := range keys {
				_, err := server.GetObject(bucketName, key)
				require.Error(t, err, "Object should be deleted")
			}

			require.LessOrEqual(t, maxConcurrent, concurrency,
				"Maximum concurrent operations should not exceed configured concurrency")

			if concurrency > 1 {
				require.Greater(t, maxConcurrent, 1,
					"Should have multiple concurrent operations when concurrency > 1")
			}
		})
	}
}
