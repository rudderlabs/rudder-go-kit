package filemanager_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	jsoniter "github.com/json-iterator/go"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/filemanager"
	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

var (
	AzuriteEndpoint, gcsURL, minioEndpoint, azureSASTokens string
	base64Secret                                           = base64.StdEncoding.EncodeToString([]byte(secretAccessKey))
	bucket                                                 = "filemanager-test-1"
	region                                                 = "us-east-1"
	accessKeyId                                            = "MYACCESSKEY"
	secretAccessKey                                        = "MYSECRETKEY"
	hold                                                   bool
	regexRequiredSuffix                                    = regexp.MustCompile(".json.gz$")
	fileList                                               []string
)

func TestMain(m *testing.M) {
	config.Reset()
	logger.Reset()

	os.Exit(run(m))
}

// run minio server & store data in it.
func run(m *testing.M) int {
	flag.BoolVar(&hold, "hold", false, "hold environment clean-up after test execution until Ctrl+C is provided")
	flag.Parse()

	// docker pool setup
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Errorf("could not connect to docker: %s", err))
	}

	// running minio container on docker
	minioResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "minio/minio",
		Tag:        "latest",
		Cmd:        []string{"server", "/data"},
		Env: []string{
			fmt.Sprintf("MINIO_ACCESS_KEY=%s", accessKeyId),
			fmt.Sprintf("MINIO_SECRET_KEY=%s", secretAccessKey),
			fmt.Sprintf("MINIO_SITE_REGION=%s", region),
			"MINIO_API_SELECT_PARQUET=on",
		},
	})
	if err != nil {
		panic(fmt.Errorf("could not start resource: %s", err))
	}
	defer func() {
		if err := pool.Purge(minioResource); err != nil {
			log.Printf("Could not purge resource: %s \n", err)
		}
	}()

	minioEndpoint = fmt.Sprintf("localhost:%s", minioResource.GetPort("9000/tcp"))

	// check if minio server is up & running.
	if err := pool.Retry(func() error {
		url := fmt.Sprintf("http://%s/minio/health/live", minioEndpoint)
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer func() { httputil.CloseResponse(resp) }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("status code not OK")
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}
	fmt.Println("minio is up & running properly")

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyId, secretAccessKey, ""),
		Secure: false, // no SSL
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("minioClient created successfully")

	// creating bucket inside minio where testing will happen.
	err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: "us-east-1"})
	if err != nil {
		panic(err)
	}
	fmt.Println("bucket created successfully")

	// Running Azure emulator, Azurite.
	AzuriteResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "mcr.microsoft.com/azure-storage/azurite",
		Tag:        "latest",
		Env: []string{
			fmt.Sprintf("AZURITE_ACCOUNTS=%s:%s", accessKeyId, base64Secret),
			fmt.Sprintf("DefaultEndpointsProtocol=%s", "http"),
		},
	})
	if err != nil {
		log.Fatalf("Could not start azure resource: %s", err)
	}
	defer func() {
		if err := pool.Purge(AzuriteResource); err != nil {
			log.Printf("Could not purge resource: %s \n", err)
		}
	}()
	AzuriteEndpoint = fmt.Sprintf("localhost:%s", AzuriteResource.GetPort("10000/tcp"))
	fmt.Println("Azurite endpoint", AzuriteEndpoint)
	fmt.Println("azurite resource successfully created")

	azureSASTokens, err = createAzureSASTokens()
	if err != nil {
		log.Fatalf("Could not create azure sas tokens: %s", err)
	}

	// Running GCS emulator
	GCSResource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "fsouza/fake-gcs-server",
		Tag:        "1.49.0",
		Cmd:        []string{"-scheme", "http", "-backend", "memory", "-location", "us-east-1"},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	defer func() {
		if err := pool.Purge(GCSResource); err != nil {
			log.Printf("Could not purge resource: %s \n", err)
		}
	}()

	GCSEndpoint := fmt.Sprintf("localhost:%s", GCSResource.GetPort("4443/tcp"))
	fmt.Println("GCS test server successfully created with endpoint: ", GCSEndpoint)
	gcsURL = fmt.Sprintf("http://%s/storage/v1/", GCSEndpoint)
	_ = os.Setenv("STORAGE_EMULATOR_HOST", GCSEndpoint)
	_ = os.Setenv("RSERVER_WORKLOAD_IDENTITY_TYPE", "GKE")

	for i := 0; i < 10; i++ {
		if err := func() error {
			client, err := storage.NewClient(context.TODO(), option.WithEndpoint(gcsURL))
			if err != nil {
				return fmt.Errorf("failed to create client: %w", err)
			}
			bkt := client.Bucket(bucket)
			err = bkt.Create(context.Background(), "test", &storage.BucketAttrs{Name: bucket})
			if err != nil {
				return fmt.Errorf("failed to create bucket: %w", err)
			}
			return nil
		}(); err != nil {
			if i == 9 {
				log.Fatalf("Could not connect to docker: %s", err)
			}
			time.Sleep(time.Second)
			continue
		}
		fmt.Println("bucket created successfully")
		break
	}

	// getting list of files in `testData` directory while will be used to testing filemanager.
	searchDir := "./goldenDirectory"
	err = filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		if regexRequiredSuffix.MatchString(path) {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	if len(fileList) == 0 {
		panic("file list empty, no data to test.")
	}
	fmt.Println("files list: ", fileList)

	code := m.Run()
	blockOnHold()
	return code
}

func createAzureSASTokens() (string, error) {
	credential, err := azblob.NewSharedKeyCredential(accessKeyId, base64Secret)
	if err != nil {
		return "", err
	}

	sasQueryParams, err := azblob.AccountSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPSandHTTP,
		ExpiryTime:    time.Now().UTC().Add(1 * time.Hour),
		Permissions:   azblob.AccountSASPermissions{Read: true, List: true, Write: true, Delete: true}.String(),
		Services:      azblob.AccountSASServices{Blob: true}.String(),
		ResourceTypes: azblob.AccountSASResourceTypes{Container: true, Object: true}.String(),
	}.NewSASQueryParameters(credential)
	if err != nil {
		return "", err
	}

	return sasQueryParams.Encode(), nil
}

func TestFileManager(t *testing.T) {
	tests := []struct {
		name          string
		skip          string
		destName      string
		config        map[string]interface{}
		otherPrefixes []string
	}{
		{
			name:          "testing s3manager functionality",
			destName:      "S3",
			otherPrefixes: []string{"other-prefix-1", "other-prefix-2"},
			config: map[string]interface{}{
				"bucketName":       bucket,
				"accessKeyID":      accessKeyId,
				"accessKey":        secretAccessKey,
				"enableSSE":        false,
				"prefix":           "some-prefix",
				"endPoint":         minioEndpoint,
				"s3ForcePathStyle": true,
				"disableSSL":       true,
				"region":           region,
			},
		},
		{
			name:          "testing minio functionality",
			destName:      "MINIO",
			otherPrefixes: []string{"other-prefix-1", "other-prefix-2"},
			config: map[string]interface{}{
				"bucketName":       bucket,
				"accessKeyID":      accessKeyId,
				"secretAccessKey":  secretAccessKey,
				"enableSSE":        false,
				"prefix":           "some-prefix",
				"endPoint":         minioEndpoint,
				"s3ForcePathStyle": true,
				"disableSSL":       true,
				"region":           region,
			},
		},
		{
			name:          "testing digital ocean functionality",
			destName:      "DIGITAL_OCEAN_SPACES",
			otherPrefixes: []string{"other-prefix-1", "other-prefix-2"},
			config: map[string]interface{}{
				"bucketName":     bucket,
				"accessKeyID":    accessKeyId,
				"accessKey":      secretAccessKey,
				"prefix":         "some-prefix",
				"endPoint":       minioEndpoint,
				"forcePathStyle": true,
				"disableSSL":     true,
				"region":         region,
				"enableSSE":      false,
			},
		},
		{
			name:          "testing Azure blob storage filemanager functionality with account keys configured",
			destName:      "AZURE_BLOB",
			otherPrefixes: []string{"other-prefix-1", "other-prefix-2"},
			config: map[string]interface{}{
				"containerName":  bucket,
				"prefix":         "some-prefix",
				"accountName":    accessKeyId,
				"accountKey":     base64Secret,
				"endPoint":       AzuriteEndpoint,
				"forcePathStyle": true,
				"disableSSL":     true,
			},
		},
		{
			name:          "testing Azure blob storage filemanager functionality with sas tokens configured",
			destName:      "AZURE_BLOB",
			otherPrefixes: []string{"other-prefix-1", "other-prefix-2"},
			config: map[string]interface{}{
				"containerName":  bucket,
				"prefix":         "some-prefix",
				"accountName":    accessKeyId,
				"useSASTokens":   true,
				"sasToken":       azureSASTokens,
				"endPoint":       AzuriteEndpoint,
				"forcePathStyle": true,
				"disableSSL":     true,
			},
		},
		{
			name:          "testing GCS filemanager functionality",
			destName:      "GCS",
			otherPrefixes: []string{"other-prefix-1", "other-prefix-2"},
			config: map[string]interface{}{
				"bucketName": bucket,
				"prefix":     "some-prefix",
				"endPoint":   gcsURL,
				"disableSSL": true,
				"jsonReads":  true,
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			fm, err := filemanager.New(&filemanager.Settings{
				Provider: tt.destName,
				Config:   tt.config,
				Logger:   logger.NOP,
				Conf:     config.New(),
			})
			if err != nil {
				t.Fatal(err)
			}
			prefix := tt.config["prefix"].(string)

			var prefixes []string
			if prefix != "" {
				prefixes = append(prefixes, prefix)
			}
			prefixes = append(prefixes, tt.otherPrefixes...)

			// upload all files
			uploadOutputs := make([]filemanager.UploadedFile, 0)
			for _, file := range fileList {
				filePtr, err := os.Open(file)
				require.NoError(t, err, "error while opening testData file to upload")
				uploadOutput, err := fm.Upload(context.TODO(), filePtr, tt.otherPrefixes...)
				require.NoError(t, err, "error while uploading file")
				paths := append([]string{}, prefixes...)
				paths = append(paths, path.Base(file))
				require.Equal(t, path.Join(paths...),
					uploadOutput.ObjectName)
				uploadOutputs = append(uploadOutputs, uploadOutput)
				require.NoError(t, filePtr.Close())
			}

			// list files using ListFilesWithPrefix
			originalFileObject := make([]*filemanager.FileInfo, 0)
			originalFileNames := make(map[string]int)
			fileListNames := make(map[string]int)

			session := fm.ListFilesWithPrefix(context.TODO(), path.Join(prefixes...), "", 1)
			for i := 0; i < len(fileList); i++ {
				files, err := session.Next()
				require.NoError(t, err, "expected no error while listing files")
				require.Equal(t, 1, len(files), "number of files should be 1")
				originalFileObject = append(originalFileObject, files[0])
				originalFileNames[files[0].Key]++
				paths := append([]string{}, prefixes...)
				paths = append(paths, path.Base(fileList[i]))
				fileListNames[path.Join(paths...)]++
			}
			require.Equal(t, len(originalFileObject), len(fileList), "actual number of files different than expected")
			for fileListName, count := range fileListNames {
				require.Equal(t, count, originalFileNames[fileListName], "files different than expected when listed")
			}

			tempFm, err := filemanager.New(&filemanager.Settings{
				Provider: tt.destName,
				Config:   tt.config,
				Logger:   logger.NOP,
				Conf:     config.New(),
			})
			if err != nil {
				t.Fatal(err)
			}

			iteratorMap := make(map[string]int)
			iteratorCount := 0
			iter := filemanager.IterateFilesWithPrefix(context.TODO(), path.Join(prefixes...), "", int64(len(fileList)), tempFm)
			for iter.Next() {
				iteratorFile := iter.Get().Key
				iteratorMap[iteratorFile]++
				iteratorCount++
			}
			require.NoError(t, iter.Err(), "no error expected while iterating files")
			require.Equal(t, len(fileList), iteratorCount, "actual number of files different than expected")
			for fileListName, count := range fileListNames {
				require.Equal(t, count, iteratorMap[fileListName], "files different than expected when iterated")
			}

			// based on the obtained location, get object name by calling GetObjectNameFromLocation
			objectName, err := fm.GetObjectNameFromLocation(uploadOutputs[0].Location)
			require.NoError(t, err, "no error expected")
			require.Equal(t, uploadOutputs[0].ObjectName, objectName, "actual object name different than expected")

			// also get download key from file location by calling GetDownloadKeyFromFileLocation
			expectedKey := uploadOutputs[0].ObjectName
			key := fm.GetDownloadKeyFromFileLocation(uploadOutputs[0].Location)
			require.Equal(t, expectedKey, key, "actual object key different than expected")

			// get prefix based on config
			splitString := strings.Split(uploadOutputs[0].ObjectName, "/")
			var expectedPrefix string
			if len(splitString) > 1 {
				expectedPrefix = splitString[0]
			}
			actualPrefix := fm.Prefix()
			require.Equal(t, expectedPrefix, actualPrefix, "actual prefix different than expected")

			// download one of the files & assert if it matches the original one present locally.
			filePtr, err := os.Open(fileList[0])
			if err != nil {
				fmt.Printf("error: %s while opening file: %s ", err, fileList[0])
			}
			originalFile, err := io.ReadAll(filePtr)
			if err != nil {
				fmt.Printf("error: %s, while reading file: %s", err, fileList[0])
			}
			require.NoError(t, filePtr.Close())

			DownloadedFileName := path.Join(t.TempDir(), "TmpDownloadedFile")

			// fail to download the file with cancelled context
			filePtr, err = os.OpenFile(DownloadedFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
			if err != nil {
				fmt.Println("error while Creating file to download data: ", err)
			}
			ctx, cancel := context.WithCancel(context.TODO())
			cancel()
			err = fm.Download(ctx, filePtr, key)
			require.Error(t, err, "expected error while downloading file")
			require.NoError(t, filePtr.Close())

			filePtr, err = os.OpenFile(DownloadedFileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
			if err != nil {
				fmt.Println("error while Creating file to download data: ", err)
			}
			err = fm.Download(context.TODO(), filePtr, key)

			require.NoError(t, err, "expected no error")
			require.NoError(t, filePtr.Close())
			filePtr, err = os.OpenFile(DownloadedFileName, os.O_RDWR, 0o644)
			if err != nil {
				fmt.Println("error while Creating file to download data: ", err)
			}
			downloadedFile, err := io.ReadAll(filePtr)
			if err != nil {
				fmt.Println("error while reading downloaded file: ", err)
			}
			require.NoError(t, filePtr.Close())

			ans := strings.Compare(string(originalFile), string(downloadedFile))
			require.Equal(t, 0, ans, "downloaded file different than actual file")

			// fail to delete the file with cancelled context
			ctx, cancel = context.WithCancel(context.TODO())
			cancel()
			err = fm.Delete(ctx, []string{key})
			require.Error(t, err, "expected error while deleting file")

			// delete that file
			err = fm.Delete(context.TODO(), []string{key})
			require.NoError(t, err, "expected no error while deleting object")
			// list files again & assert if that file is still present.
			fmNew, err := filemanager.New(&filemanager.Settings{
				Provider: tt.destName,
				Config:   tt.config,
				Logger:   logger.NOP,
				Conf:     config.New(),
			})
			if err != nil {
				panic(err)
			}
			newFileObject, err := fmNew.ListFilesWithPrefix(context.TODO(), "", "", 1000).Next()
			if err != nil {
				fmt.Println("error while getting new file object: ", err)
			}
			require.Equal(t, len(originalFileObject)-1, len(newFileObject), "expected original file list length to be greater than new list by 1, but is different")
		})

		t.Run(tt.name, func(t *testing.T) {
			if tt.skip != "" {
				t.Skip(tt.skip)
			}
			fm, err := filemanager.New(&filemanager.Settings{
				Provider: tt.destName,
				Config:   tt.config,
				Logger:   logger.NOP,
				Conf:     config.New(),
			})
			if err != nil {
				t.Fatal(err)
			}

			// fail to upload file
			file := fileList[0]
			filePtr, err := os.Open(file)
			require.NoError(t, err, "error while opening testData file to upload")
			ctx, cancel := context.WithTimeout(context.TODO(), time.Second*5)
			cancel()
			_, err = fm.Upload(ctx, filePtr)
			require.Error(t, err, "expected error while uploading file")
			require.NoError(t, filePtr.Close())

			// MINIO doesn't support list files with context cancellation
			if tt.destName != "MINIO" {
				// fail to fetch file list
				ctx1, cancel := context.WithTimeout(context.TODO(), time.Second*5)
				cancel()
				_, err = fm.ListFilesWithPrefix(ctx1, "", "", 1000).Next()
				require.Error(t, err, "expected error while listing files")

				iter := filemanager.IterateFilesWithPrefix(ctx1, "", "", 1000, fm)
				next := iter.Next()
				require.Equal(t, false, next, "next should be false when context is cancelled")
				err = iter.Err()
				require.Error(t, err, "expected error while iterating files")
			}
		})

	}
}

func TestGCSManager_unsupported_credentials(t *testing.T) {
	var conf map[string]any
	err := jsoniter.Unmarshal(
		[]byte(`{
			"project": "my-project",
			"location": "US",
			"bucketName": "my-bucket",
			"prefix": "rudder",
			"namespace": "namespace",
			"credentials":"{\"installed\":{\"client_id\":\"1234.apps.googleusercontent.com\",\"project_id\":\"project_id\",\"auth_uri\":\"https://accounts.google.com/o/oauth2/auth\",\"token_uri\":\"https://oauth2.googleapis.com/token\",\"auth_provider_x509_cert_url\":\"https://www.googleapis.com/oauth2/v1/certs\",\"client_secret\":\"client_secret\",\"redirect_uris\":[\"urn:ietf:wg:oauth:2.0:oob\",\"http://localhost\"]}}",
			"syncFrequency": "1440",
			"syncStartAt": "09:00"
		}`),
		&conf,
	)
	assert.NoError(t, err)
	manager, err := filemanager.NewGCSManager(conf, logger.NOP, func() time.Duration { return time.Minute })
	assert.NoError(t, err)
	_, err = manager.ListFilesWithPrefix(context.TODO(), "", "/tests", 100).Next()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "client_credentials.json file is not supported")
}

func blockOnHold() {
	if !hold {
		return
	}

	log.Println("Test on hold, before cleanup")
	log.Println("Press Ctrl+C to exit")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
	close(c)
}

func TestFileManager_S3(t *testing.T) {
	// Prepare a small file for upload
	tempDir := t.TempDir()
	testFilePath := filepath.Join(tempDir, "testfile.txt")
	testFileContent := []byte("integration test content")
	require.NoError(t, os.WriteFile(testFilePath, testFileContent, 0o644))

	envAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	envSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	envBucket := os.Getenv("AWS_BUCKET_NAME")
	isV2ManagerEnabled := []bool{false, true}
	for _, enabled := range isV2ManagerEnabled {
		authMethods := []struct {
			name   string
			config map[string]any
		}{
			{
				name: "AccessKey/Secret",
				config: map[string]any{
					"bucketName":       envBucket,
					"accessKeyID":      envAccessKey,
					"secretAccessKey":  envSecretKey,
					"region":           region,
					"s3ForcePathStyle": true,
					"disableSSL":       true,
					"prefix":           "",
				},
			},
			{
				name: "AccessKey/Secret With Prefix",
				config: map[string]any{
					"bucketName":       envBucket,
					"accessKeyID":      envAccessKey,
					"secretAccessKey":  envSecretKey,
					"region":           region,
					"s3ForcePathStyle": true,
					"disableSSL":       true,
					"prefix":           "test-prefix",
				},
			},
		}

		for _, auth := range authMethods {
			t.Run("running with: "+auth.name+" and v2 manager enabled: "+strconv.FormatBool(enabled), func(t *testing.T) {
				conf := config.New()
				conf.Set("FileManager.S3ManagerV2", enabled)
				fm, err := filemanager.New(&filemanager.Settings{
					Provider: "S3",
					Config:   auth.config,
					Logger:   logger.NOP,
					Conf:     conf,
				})
				require.NoError(t, err)

				// 1. Upload a file
				filePtr, err := os.Open(testFilePath)
				require.NoError(t, err)
				uploadOutput, err := fm.Upload(context.TODO(), filePtr)
				require.NoError(t, err)
				require.NoError(t, filePtr.Close())
				// check if the file name is exactly the same as the one we uploaded
				require.Equal(t, uploadOutput.ObjectName, path.Join(auth.config["prefix"].(string), "testfile.txt"), "uploaded file name should be exactly the same as the one we uploaded")

				// 2. List files and check our file is present
				session := fm.ListFilesWithPrefix(context.TODO(), "", "", 100)
				files, err := session.Next()
				require.NoError(t, err)
				var found bool
				for _, f := range files {
					if f.Key == uploadOutput.ObjectName {
						found = true
						break
					}
				}
				require.True(t, found, "uploaded file should be listed")

				// 3. Download the file and verify contents
				downloadPath := filepath.Join(tempDir, "downloaded.txt")
				downloadPtr, err := os.OpenFile(downloadPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
				require.NoError(t, err)
				err = fm.Download(context.TODO(), downloadPtr, uploadOutput.ObjectName)
				require.NoError(t, err)
				require.NoError(t, downloadPtr.Close())

				downloadedContent, err := os.ReadFile(downloadPath)
				require.NoError(t, err)
				require.Equal(t, testFileContent, downloadedContent, "downloaded file content should match uploaded")

				// 4. Test GetObjectNameFromLocation
				objectName, err := fm.GetObjectNameFromLocation(uploadOutput.Location)
				require.NoError(t, err)
				require.Equal(t, uploadOutput.ObjectName, objectName, "object name from location should match")

				// 5. Test GetDownloadKeyFromFileLocation
				downloadKey := fm.GetDownloadKeyFromFileLocation(uploadOutput.Location)
				require.Equal(t, uploadOutput.ObjectName, downloadKey, "download key from location should match")

				// 6. Test UploadReader
				readerContent := []byte("test content from reader")
				readerObjName := "test-reader-upload.txt"
				uploadReaderOutput, err := fm.UploadReader(context.TODO(), readerObjName, bytes.NewReader(readerContent))
				require.NoError(t, err)

				// 7. Download file uploaded via UploadReader and verify
				downloadReaderPath := filepath.Join(tempDir, "downloaded-reader.txt")
				downloadReaderPtr, err := os.OpenFile(downloadReaderPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0o644)
				require.NoError(t, err)
				err = fm.Download(context.TODO(), downloadReaderPtr, uploadReaderOutput.ObjectName)
				require.NoError(t, err)
				require.NoError(t, downloadReaderPtr.Close())

				downloadedReaderContent, err := os.ReadFile(downloadReaderPath)
				require.NoError(t, err)
				require.Equal(t, readerContent, downloadedReaderContent, "downloaded content should match uploaded reader content")

				// 8. Test Delete
				err = fm.Delete(context.TODO(), []string{uploadReaderOutput.ObjectName})
				require.NoError(t, err)

				// 9. Verify deletion
				deletedSession := fm.ListFilesWithPrefix(context.TODO(), "", "", 100)
				deletedFiles, err := deletedSession.Next()
				require.NoError(t, err)
				for _, f := range deletedFiles {
					require.NotEqual(t, uploadReaderOutput.ObjectName, f.Key, "deleted file should not be present")
				}
			})
		}
	}
}

func TestS3Manager_SelectObjects(t *testing.T) {
	minioConfig := map[string]interface{}{
		"bucketName":       bucket,
		"accessKeyID":      accessKeyId,
		"accessKey":        secretAccessKey,
		"enableSSE":        false,
		"prefix":           "some-prefix",
		"endPoint":         minioEndpoint,
		"s3ForcePathStyle": true,
		"disableSSL":       true,
		"region":           region,
	}

	fm, err := filemanager.New(&filemanager.Settings{
		Provider: "S3",
		Config:   minioConfig,
		Logger:   logger.NOP,
		Conf:     config.New(),
	})
	if err != nil {
		t.Fatalf("failed to create S3 filemanager: %v", err)
	}

	parquetPath := "goldenDirectory/test.parquet"
	filePtr, err := os.Open(parquetPath)
	if err != nil {
		t.Fatalf("failed to open parquet file: %v", err)
	}
	defer filePtr.Close()

	uploadOutput, err := fm.Upload(context.TODO(), filePtr)
	if err != nil {
		t.Fatalf("failed to upload parquet file: %v", err)
	}
	// Always clean up the uploaded file at the end
	defer func() {
		if err := fm.Delete(context.TODO(), []string{uploadOutput.ObjectName}); err != nil {
			t.Errorf("failed to delete parquet file from bucket during cleanup: %v", err)
		}
	}()

	s3fm, ok := fm.(filemanager.S3Manager)
	if !ok {
		t.Fatalf("filemanager is not an S3Manager, cannot call SelectObjects")
	}

	// Helper to run select and assert output
	runSelectAndAssert := func(t *testing.T, outputFormat string) {
		selectConfig := filemanager.SelectConfig{
			SQLExpression: "SELECT * FROM S3Object",
			Key:           uploadOutput.ObjectName,
			InputFormat:   "parquet",
			OutputFormat:  outputFormat,
		}
		selectResult, err := s3fm.SelectObjects(context.TODO(), selectConfig)
		if err != nil {
			t.Fatalf("failed to call SelectObjects (%s): %v", outputFormat, err)
		}
		var receivedData bool
		for data := range selectResult {
			receivedData = true

			if data.Error != nil {
				t.Errorf("error received from SelectObjects (%s): %v", outputFormat, data.Error)
			}

			if len(data.Data) == 0 {
				t.Errorf("received empty data from SelectObjects (%s)", outputFormat)
			}
			lines := bytes.Split(data.Data, []byte("\n"))
			for i, line := range lines {
				if len(bytes.TrimSpace(line)) == 0 {
					continue
				}
				if outputFormat == "json" {
					var js map[string]interface{}
					if err := jsoniter.Unmarshal(line, &js); err != nil {
						t.Errorf("received line is not valid JSON: %v\ndata: %s", err, string(line))
					}
				} else if outputFormat == "csv" && i == 0 {
					row := string(line)
					columns := strings.Split(row, ",")
					if len(columns) < 4 {
						t.Errorf("CSV row does not have expected number of columns: %s", row)
					}
				}
			}
		}
		if !receivedData {
			t.Errorf("did not receive any data from SelectObjects (%s)", outputFormat)
		}
	}

	t.Run("SelectObjects with JSON output", func(t *testing.T) {
		runSelectAndAssert(t, "json")
	})

	t.Run("SelectObjects with CSV output", func(t *testing.T) {
		runSelectAndAssert(t, "csv")
	})
}
