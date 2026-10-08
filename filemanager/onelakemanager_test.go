package filemanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/datalakeerror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const (
	testOneLakeWorkspace = "11111111-1111-1111-1111-111111111111"
	testOneLakeLakehouse = "22222222-2222-2222-2222-222222222222"
	testOneLakeSecret    = "test-client-secret"
)

type staticOneLakeCredential struct{}

func (staticOneLakeCredential) GetToken(context.Context, policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: "test-token", ExpiresOn: time.Now().Add(time.Hour)}, nil
}

type oneLakeFakeServer struct {
	t              *testing.T
	server         *httptest.Server
	mu             sync.Mutex
	objects        map[string][]byte
	requests       []*http.Request
	listPages      [][]oneLakeFakePath
	listMissing    bool
	downloadStatus int
	downloadCode   string
	delay          time.Duration
}

type oneLakeFakePath struct {
	name         string
	isDirectory  bool
	lastModified string
}

func newOneLakeFakeServer(t *testing.T) *oneLakeFakeServer {
	fake := &oneLakeFakeServer{t: t, objects: make(map[string][]byte)}
	fake.server = httptest.NewTLSServer(http.HandlerFunc(fake.serveHTTP))
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *oneLakeFakeServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-r.Context().Done():
			return
		}
	}

	f.mu.Lock()
	requestCopy := r.Clone(context.Background())
	f.requests = append(f.requests, requestCopy)
	f.mu.Unlock()

	if r.Method == http.MethodGet && r.URL.Query().Get("resource") == "filesystem" {
		f.serveList(w, r)
		return
	}
	if r.Method == http.MethodGet {
		f.serveDownload(w, r)
		return
	}

	objectPath := strings.TrimPrefix(r.URL.Path, "/"+testOneLakeWorkspace+"/")
	switch {
	case r.Method == http.MethodPut && r.URL.Query().Get("resource") == "file":
		// Like OneLake, creating an existing file truncates it.
		f.mu.Lock()
		f.objects[objectPath] = []byte{}
		f.mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	case r.Method == http.MethodPatch && r.URL.Query().Get("action") == "append":
		// Like OneLake, appends require an existing file (only create truncates) and a non-empty body.
		body, err := io.ReadAll(r.Body)
		require.NoError(f.t, err)
		if len(body) == 0 {
			writeOneLakeError(w, http.StatusBadRequest, "InvalidInput")
			return
		}
		f.mu.Lock()
		existing, found := f.objects[objectPath]
		if found {
			f.objects[objectPath] = append(existing, body...)
		}
		f.mu.Unlock()
		if !found {
			writeOneLakeError(w, http.StatusNotFound, "PathNotFound")
			return
		}
		w.WriteHeader(http.StatusAccepted)
	case r.Method == http.MethodPatch && r.URL.Query().Get("action") == "flush":
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodDelete:
		f.mu.Lock()
		_, found := f.objects[objectPath]
		delete(f.objects, objectPath)
		f.mu.Unlock()
		if !found {
			writeOneLakeError(w, http.StatusNotFound, "PathNotFound")
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "unexpected request", http.StatusBadRequest)
	}
}

func (f *oneLakeFakeServer) serveDownload(w http.ResponseWriter, r *http.Request) {
	if f.downloadStatus != 0 {
		writeOneLakeError(w, f.downloadStatus, f.downloadCode)
		return
	}
	objectPath := strings.TrimPrefix(r.URL.Path, "/"+testOneLakeWorkspace+"/")
	f.mu.Lock()
	data, found := f.objects[objectPath]
	f.mu.Unlock()
	if !found {
		writeOneLakeError(w, http.StatusNotFound, "PathNotFound")
		return
	}

	w.Header().Set("ETag", `"test-etag"`)
	w.Header().Set("Accept-Ranges", "bytes")
	if rangeHeader := requestRange(r); rangeHeader != "" {
		var start, end int
		if strings.HasSuffix(rangeHeader, "-") {
			_, err := fmt.Sscanf(rangeHeader, "bytes=%d-", &start)
			require.NoError(f.t, err)
			end = len(data) - 1
		} else {
			_, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end)
			require.NoError(f.t, err)
		}
		w.Header().Set("Content-Range", "bytes 0-0/1")
		w.WriteHeader(http.StatusPartialContent)
		_, err := w.Write(data[start : end+1])
		require.NoError(f.t, err)
		return
	}
	_, err := w.Write(data)
	require.NoError(f.t, err)
}

func (f *oneLakeFakeServer) serveList(w http.ResponseWriter, r *http.Request) {
	if f.listMissing {
		writeOneLakeError(w, http.StatusNotFound, "PathNotFound")
		return
	}
	page := 0
	if marker := r.URL.Query().Get("continuation"); marker != "" {
		_, err := fmt.Sscanf(marker, "page-%d", &page)
		require.NoError(f.t, err)
	}
	if page >= len(f.listPages) {
		_, err := io.WriteString(w, `{"paths":[]}`)
		require.NoError(f.t, err)
		return
	}
	if page+1 < len(f.listPages) {
		w.Header().Set("x-ms-continuation", "page-"+strconv.Itoa(page+1))
	}
	type listedPath struct {
		Name         string `json:"name"`
		IsDirectory  string `json:"isDirectory"`
		LastModified string `json:"lastModified,omitempty"`
	}
	paths := make([]listedPath, 0, len(f.listPages[page]))
	for _, item := range f.listPages[page] {
		paths = append(paths, listedPath{Name: item.name, IsDirectory: strconv.FormatBool(item.isDirectory), LastModified: item.lastModified})
	}
	require.NoError(f.t, json.NewEncoder(w).Encode(map[string]any{"paths": paths}))
}

func writeOneLakeError(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("x-ms-error-code", code)
	w.WriteHeader(status)
	_, _ = io.WriteString(w, `{"error":{"code":"`+code+`","message":"test response"}}`)
}

func testOneLakeConfig() map[string]any {
	return map[string]any{
		"fabricWorkspaceId": testOneLakeWorkspace,
		"lakehouseId":       testOneLakeLakehouse,
		"tenantId":          "tenant",
		"clientId":          "client",
		"clientSecret":      testOneLakeSecret,
	}
}

func newTestOneLakeManager(t *testing.T, fake *oneLakeFakeServer) *OneLakeManager {
	manager, err := newOneLakeManager(testOneLakeConfig(), logger.NOP, func() time.Duration { return time.Minute }, oneLakeManagerOptions{
		serviceEndpoint: fake.server.URL,
		credential:      staticOneLakeCredential{},
		clientOptions: azcore.ClientOptions{
			Transport: fake.server.Client(),
		},
	})
	require.NoError(t, err)
	return manager
}

func TestNewOneLakeManager(t *testing.T) {
	t.Run("factory", func(t *testing.T) {
		manager, err := New(&Settings{Provider: "ONELAKE", Config: testOneLakeConfig(), Logger: logger.NOP, Conf: kitconfig.New()})
		require.NoError(t, err)
		require.IsType(t, &OneLakeManager{}, manager)
		require.Equal(t, "", manager.Prefix())
		require.Equal(t, defaultTimeout, manager.(*OneLakeManager).getTimeout())

		_, err = New(&Settings{Provider: "UNKNOWN", Config: map[string]any{}, Logger: logger.NOP, Conf: kitconfig.New()})
		require.ErrorIs(t, err, ErrInvalidServiceProvider)
	})

	t.Run("timeout configuration", func(t *testing.T) {
		conf := kitconfig.New()
		conf.Set("FileManager.ONELAKE.timeout", "3s")
		manager, err := New(&Settings{Provider: "ONELAKE", Config: testOneLakeConfig(), Logger: logger.NOP, Conf: conf})
		require.NoError(t, err)
		require.Equal(t, 3*time.Second, manager.(*OneLakeManager).getTimeout())
	})

	t.Run("required config", func(t *testing.T) {
		for _, key := range []string{"fabricWorkspaceId", "lakehouseId", "tenantId", "clientId", "clientSecret"} {
			t.Run(key+" missing", func(t *testing.T) {
				config := testOneLakeConfig()
				delete(config, key)
				_, err := NewOneLakeManager(config, logger.NOP, nil)
				require.ErrorContains(t, err, key)
				require.NotContains(t, err.Error(), testOneLakeSecret)
			})
			t.Run(key+" blank", func(t *testing.T) {
				config := testOneLakeConfig()
				config[key] = "  "
				_, err := NewOneLakeManager(config, logger.NOP, nil)
				require.ErrorContains(t, err, key)
			})
		}
	})

	t.Run("canonical ids", func(t *testing.T) {
		for _, key := range []string{"fabricWorkspaceId", "lakehouseId"} {
			for _, invalid := range []string{"AAAAAAAA-AAAA-AAAA-AAAA-AAAAAAAAAAAA", "{aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa}", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"} {
				config := testOneLakeConfig()
				config[key] = invalid
				_, err := NewOneLakeManager(config, logger.NOP, nil)
				require.ErrorContains(t, err, key)
			}
		}
	})

	t.Run("unknown config ignored", func(t *testing.T) {
		config := testOneLakeConfig()
		config["prefix"] = "ignored"
		config["oneLakeHost"] = "https://invalid.example"
		manager, err := NewOneLakeManager(config, logger.NOP, nil)
		require.NoError(t, err)
		require.Equal(t, "", manager.Prefix())
		require.Equal(t, oneLakeEndpoint+"/"+testOneLakeWorkspace, manager.filesystem.DFSURL())
	})

	t.Run("credential cache", func(t *testing.T) {
		oneLakeCredentialCache.Clear()
		t.Cleanup(oneLakeCredentialCache.Clear)
		first, err := NewOneLakeManager(testOneLakeConfig(), logger.NOP, nil)
		require.NoError(t, err)
		second, err := NewOneLakeManager(testOneLakeConfig(), logger.NOP, nil)
		require.NoError(t, err)
		require.Same(t, first.credential, second.credential)

		changed := testOneLakeConfig()
		changed["clientSecret"] = "different-secret"
		third, err := NewOneLakeManager(changed, logger.NOP, nil)
		require.NoError(t, err)
		require.NotSame(t, first.credential, third.credential)
	})
}

func TestOneLakeUploadAndValidation(t *testing.T) {
	fake := newOneLakeFakeServer(t)
	manager := newTestOneLakeManager(t, fake)

	file, err := os.CreateTemp(t.TempDir(), "payload-*.parquet")
	require.NoError(t, err)
	_, err = file.WriteString("uploaded file")
	require.NoError(t, err)
	require.NoError(t, file.Close())
	file, err = os.Open(file.Name())
	require.NoError(t, err)
	defer func() { require.NoError(t, file.Close()) }()

	uploaded, err := manager.Upload(context.Background(), file, "a", "b")
	require.NoError(t, err)
	expectedName := path.Join("a", "b", path.Base(file.Name()))
	require.Equal(t, expectedName, uploaded.ObjectName)
	require.Equal(t, oneLakeEndpoint+"/"+testOneLakeWorkspace+"/"+testOneLakeLakehouse+"/Files/"+expectedName, uploaded.Location)
	require.Equal(t, []byte("uploaded file"), fake.objects[path.Join(testOneLakeLakehouse, "Files", expectedName)])

	empty, err := os.Create(path.Join(t.TempDir(), "empty.parquet"))
	require.NoError(t, err)
	defer func() { require.NoError(t, empty.Close()) }()
	uploaded, err = manager.Upload(context.Background(), empty, "a")
	require.NoError(t, err)
	require.Equal(t, "a/empty.parquet", uploaded.ObjectName)
	require.Equal(t, []byte{}, fake.objects[path.Join(testOneLakeLakehouse, "Files", "a/empty.parquet")])

	readSide, writeSide := io.Pipe()
	done := make(chan error, 1)
	go func() {
		_, writeErr := writeSide.Write([]byte("streamed payload"))
		if writeErr == nil {
			writeErr = writeSide.Close()
		}
		done <- writeErr
	}()
	uploaded, err = manager.UploadReader(context.Background(), "stream/data.parquet", readSide)
	require.NoError(t, err)
	require.NoError(t, <-done)
	require.Equal(t, "stream/data.parquet", uploaded.ObjectName)
	require.Equal(t, []byte("streamed payload"), fake.objects[path.Join(testOneLakeLakehouse, "Files", uploaded.ObjectName)])

	uploaded, err = manager.UploadReader(context.Background(), "stream/data.parquet", strings.NewReader("replacement payload"))
	require.NoError(t, err)
	require.Equal(t, "stream/data.parquet", uploaded.ObjectName)
	require.Equal(t, []byte("replacement payload"), fake.objects[path.Join(testOneLakeLakehouse, "Files", uploaded.ObjectName)])

	fake.objects[path.Join(testOneLakeLakehouse, "Files", "partial/data.parquet")] = []byte("partial")
	uploaded, err = manager.UploadReader(context.Background(), "partial/data.parquet", strings.NewReader("retry payload"))
	require.NoError(t, err)
	require.Equal(t, "partial/data.parquet", uploaded.ObjectName)
	require.Equal(t, []byte("retry payload"), fake.objects[path.Join(testOneLakeLakehouse, "Files", uploaded.ObjectName)])

	requestsBefore := len(fake.requests)
	for _, invalid := range []string{"", "/x", `a\b`, "a//b", "a/./b", "a/../b"} {
		_, err = manager.UploadReader(context.Background(), invalid, strings.NewReader("contents"))
		require.Error(t, err)
	}
	require.Len(t, fake.requests, requestsBefore)

	for prefix, expected := range map[string]string{
		"":     path.Base(file.Name()),
		"dir/": path.Join("dir", path.Base(file.Name())),
	} {
		uploaded, err = manager.Upload(context.Background(), file, prefix)
		require.NoError(t, err)
		require.Equal(t, expected, uploaded.ObjectName)
	}

	requestsBefore = len(fake.requests)
	for _, invalidPrefix := range []string{"/x", `a\b`, ".."} {
		_, err = manager.Upload(context.Background(), file, invalidPrefix)
		require.Error(t, err)
	}
	require.Len(t, fake.requests, requestsBefore)
}

func TestOneLakeDownload(t *testing.T) {
	fake := newOneLakeFakeServer(t)
	manager := newTestOneLakeManager(t, fake)
	key := "dir/data.parquet"
	fake.objects[path.Join(testOneLakeLakehouse, "Files", key)] = []byte("0123456789abcdefghij")

	var output bytesWriterAt
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, "0123456789abcdefghij", output.String())

	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key, WithDownloadOffSetAndLength(5, 10)))
	require.Equal(t, "56789abcde", output.String())
	require.Equal(t, "bytes=5-14", requestRange(fake.requests[len(fake.requests)-1]))

	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key, WithDownloadOffSet(5)))
	require.Equal(t, "56789abcdefghij", output.String())
	require.Equal(t, "bytes=5-", requestRange(fake.requests[len(fake.requests)-1]))

	err := manager.Download(context.Background(), &output, "missing.parquet")
	require.ErrorIs(t, err, ErrKeyNotFound)

	fake.downloadStatus = http.StatusForbidden
	fake.downloadCode = "AuthorizationPermissionMismatch"
	err = manager.Download(context.Background(), &output, key)
	require.ErrorContains(t, err, "onelake: downloading "+key)
	var responseError *azcore.ResponseError
	require.ErrorAs(t, err, &responseError)
	require.Equal(t, "AuthorizationPermissionMismatch", responseError.ErrorCode)
	require.True(t, datalakeerror.HasCode(err, datalakeerror.AuthorizationPermissionMismatch))
}

func TestOneLakeDelete(t *testing.T) {
	fake := newOneLakeFakeServer(t)
	manager := newTestOneLakeManager(t, fake)
	key := "dir/data.parquet"
	fake.objects[path.Join(testOneLakeLakehouse, "Files", key)] = []byte("contents")
	require.NoError(t, manager.Delete(context.Background(), []string{key, "missing.parquet"}))
	require.NotContains(t, fake.objects, path.Join(testOneLakeLakehouse, "Files", key))
}

func TestOneLakeListFilesWithPrefix(t *testing.T) {
	modified := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	root := testOneLakeLakehouse + "/Files/"

	t.Run("filter and paginate", func(t *testing.T) {
		fake := newOneLakeFakeServer(t)
		fake.listPages = [][]oneLakeFakePath{
			{
				{name: root + "dir", isDirectory: true},
				{name: root + "dir/file_1", lastModified: modified.Format(http.TimeFormat)},
				{name: root + "dir/other"},
				{name: root + "dir/file_2"},
			},
			{
				{name: root + "dir/file_3"},
				{name: root + "dir/file_4"},
				{name: root + "dir/file_5"},
			},
		}
		manager := newTestOneLakeManager(t, fake)
		session := manager.ListFilesWithPrefix(context.Background(), "", "dir/file_", 3)
		first, err := session.Next()
		require.NoError(t, err)
		require.Equal(t, []string{"dir/file_1", "dir/file_2", "dir/file_3"}, fileInfoKeys(first))
		require.Equal(t, modified, first[0].LastModified)
		second, err := session.Next()
		require.NoError(t, err)
		require.Equal(t, []string{"dir/file_4", "dir/file_5"}, fileInfoKeys(second))
		third, err := session.Next()
		require.NoError(t, err)
		require.Nil(t, third)
		require.Equal(t, path.Join(testOneLakeLakehouse, "Files", "dir"), fake.requests[0].URL.Query().Get("directory"))
		require.Equal(t, "true", fake.requests[0].URL.Query().Get("recursive"))
	})

	t.Run("directory and root", func(t *testing.T) {
		for prefix, expected := range map[string]string{
			"dir/": path.Join(testOneLakeLakehouse, "Files", "dir"),
			"":     path.Join(testOneLakeLakehouse, "Files"),
		} {
			fake := newOneLakeFakeServer(t)
			fake.listPages = [][]oneLakeFakePath{{}}
			manager := newTestOneLakeManager(t, fake)
			_, err := manager.ListFilesWithPrefix(context.Background(), "", prefix, 3).Next()
			require.NoError(t, err)
			require.Equal(t, expected, fake.requests[0].URL.Query().Get("directory"))
		}
	})

	t.Run("start after", func(t *testing.T) {
		fake := newOneLakeFakeServer(t)
		fake.listPages = [][]oneLakeFakePath{{
			{name: root + "dir/file_1"},
			{name: root + "dir/file_2"},
			{name: root + "dir/file_3"},
		}}
		manager := newTestOneLakeManager(t, fake)
		files, err := manager.ListFilesWithPrefix(context.Background(), "dir/file_1", "dir/file_", 3).Next()
		require.NoError(t, err)
		require.Equal(t, []string{"dir/file_2", "dir/file_3"}, fileInfoKeys(files))
	})

	t.Run("missing directory", func(t *testing.T) {
		fake := newOneLakeFakeServer(t)
		fake.listMissing = true
		manager := newTestOneLakeManager(t, fake)
		files, err := manager.ListFilesWithPrefix(context.Background(), "", "missing/", 3).Next()
		require.NoError(t, err)
		require.Nil(t, files)
	})
}

func TestOneLakeLocations(t *testing.T) {
	fake := newOneLakeFakeServer(t)
	manager := newTestOneLakeManager(t, fake)
	objectName := "dir/data.parquet"
	location := oneLakeEndpoint + "/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Files/" + objectName

	actual, err := manager.GetObjectNameFromLocation(location)
	require.NoError(t, err)
	require.Equal(t, objectName, actual)
	actual, err = manager.GetObjectNameFromLocation(objectName)
	require.NoError(t, err)
	require.Equal(t, objectName, actual)
	require.Equal(t, objectName, manager.GetDownloadKeyFromFileLocation(location))

	for _, objectName := range []string{"dir/a?b#c.parquet", "dir/a b.parquet", "dir/a%b.parquet", "dir/a%2Fb.parquet"} {
		uploaded, err := manager.UploadReader(context.Background(), objectName, strings.NewReader("x"))
		require.NoError(t, err)
		parsed, err := url.Parse(uploaded.Location)
		require.NoError(t, err)
		require.Equal(t, "", parsed.RawQuery)
		require.Equal(t, "", parsed.Fragment)
		roundTripped, err := manager.GetObjectNameFromLocation(uploaded.Location)
		require.NoError(t, err)
		require.Equal(t, objectName, roundTripped)
	}

	invalid := []string{
		"http://onelake.dfs.fabric.microsoft.com/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Files/x",
		"https://example.com/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Files/x",
		oneLakeEndpoint + "/33333333-3333-3333-3333-333333333333/" + testOneLakeLakehouse + "/Files/x",
		oneLakeEndpoint + "/" + testOneLakeWorkspace + "/33333333-3333-3333-3333-333333333333/Files/x",
		oneLakeEndpoint + "/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Tables/x",
		oneLakeEndpoint + "/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Files/../Tables/x",
		oneLakeEndpoint + "/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Files/x?sig=secret",
		oneLakeEndpoint + "/" + testOneLakeWorkspace + "/" + testOneLakeLakehouse + "/Files/a%2Fb",
	}
	for _, value := range invalid {
		_, err := manager.GetObjectNameFromLocation(value)
		require.Error(t, err)
		require.Equal(t, "", manager.GetDownloadKeyFromFileLocation(value))
	}
}

func TestOneLakeTimeout(t *testing.T) {
	fake := newOneLakeFakeServer(t)
	fake.delay = 100 * time.Millisecond
	manager := newTestOneLakeManager(t, fake)
	manager.SetTimeout(time.Millisecond)
	_, err := manager.UploadReader(context.Background(), "file.parquet", strings.NewReader("contents"))
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestOneLakeLive(t *testing.T) {
	config := oneLakeLiveConfig(t)
	manager, err := NewOneLakeManager(config, logger.NOP, func() time.Duration { return 2 * time.Minute })
	require.NoError(t, err)
	prefix := "rudder-go-kit-test/" + uuid.NewString()
	key := prefix + "/roundtrip.txt"
	t.Cleanup(func() { _ = manager.Delete(context.Background(), []string{key}) })

	contents := []byte("0123456789abcdefghij")
	uploaded, err := manager.UploadReader(context.Background(), key, bytes.NewReader(contents))
	require.NoError(t, err)
	require.Equal(t, key, uploaded.ObjectName)

	files, err := manager.ListFilesWithPrefix(context.Background(), "", prefix+"/round", 10).Next()
	require.NoError(t, err)
	require.Equal(t, []string{key}, fileInfoKeys(files))

	var output bytesWriterAt
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, contents, output.Bytes())
	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key, WithDownloadOffSetAndLength(5, 10)))
	require.Equal(t, contents[5:15], output.Bytes())

	uploaded, err = manager.UploadReader(context.Background(), key, strings.NewReader("replaced"))
	require.NoError(t, err)
	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, "replaced", output.String())

	empty, err := os.Create(path.Join(t.TempDir(), "empty.txt"))
	require.NoError(t, err)
	defer func() { require.NoError(t, empty.Close()) }()
	uploaded, err = manager.Upload(context.Background(), empty, prefix)
	require.NoError(t, err)
	t.Cleanup(func() { _ = manager.Delete(context.Background(), []string{uploaded.ObjectName}) })
	name, err := manager.GetObjectNameFromLocation(uploaded.Location)
	require.NoError(t, err)
	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, name))
	require.Zero(t, output.Len())

	require.NoError(t, manager.Delete(context.Background(), []string{key, uploaded.ObjectName}))
	require.ErrorIs(t, manager.Download(context.Background(), &output, key), ErrKeyNotFound)
}

func TestOneLakeLiveErrors(t *testing.T) {
	config := oneLakeLiveConfig(t)

	t.Run("wrong secret", func(t *testing.T) {
		config := cloneOneLakeConfig(config)
		config["clientSecret"] = testOneLakeSecret
		manager, err := NewOneLakeManager(config, logger.NOP, func() time.Duration { return 30 * time.Second })
		require.NoError(t, err)
		_, err = manager.UploadReader(context.Background(), "rudder-go-kit-test/"+uuid.NewString()+"/auth.txt", strings.NewReader("contents"))
		require.Error(t, err)
		require.NotContains(t, err.Error(), testOneLakeSecret)
	})

	t.Run("missing lakehouse", func(t *testing.T) {
		config := cloneOneLakeConfig(config)
		config["lakehouseId"] = uuid.NewString()
		manager, err := NewOneLakeManager(config, logger.NOP, func() time.Duration { return 30 * time.Second })
		require.NoError(t, err)
		_, err = manager.UploadReader(context.Background(), "rudder-go-kit-test/"+uuid.NewString()+"/missing.txt", strings.NewReader("contents"))
		require.Error(t, err)
		var responseError *azcore.ResponseError
		require.ErrorAs(t, err, &responseError)
		require.NotEmpty(t, responseError.ErrorCode)
	})
}

func oneLakeLiveConfig(t *testing.T) map[string]any {
	t.Helper()
	config := map[string]any{
		"fabricWorkspaceId": os.Getenv("ONELAKE_TEST_WORKSPACE_ID"),
		"lakehouseId":       os.Getenv("ONELAKE_TEST_LAKEHOUSE_ID"),
		"tenantId":          os.Getenv("ONELAKE_TEST_TENANT_ID"),
		"clientId":          os.Getenv("ONELAKE_TEST_CLIENT_ID"),
		"clientSecret":      os.Getenv("ONELAKE_TEST_CLIENT_SECRET"),
	}
	for key, value := range config {
		if value == "" {
			t.Skip("OneLake live test requires " + key)
		}
	}
	return config
}

func cloneOneLakeConfig(config map[string]any) map[string]any {
	clone := make(map[string]any, len(config))
	maps.Copy(clone, config)
	return clone
}

type bytesWriterAt struct {
	bytes.Buffer
}

func (w *bytesWriterAt) WriteAt(p []byte, offset int64) (int, error) {
	if offset != int64(w.Len()) {
		return 0, errors.New("non-sequential write")
	}
	return w.Write(p)
}

func requestRange(request *http.Request) string {
	if value := request.Header.Get("Range"); value != "" {
		return value
	}
	return request.Header.Get("x-ms-range")
}

func fileInfoKeys(files []*FileInfo) []string {
	keys := make([]string, 0, len(files))
	for _, item := range files {
		keys = append(keys, item.Key)
	}
	return keys
}
