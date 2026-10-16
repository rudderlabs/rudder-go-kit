package filemanager

import (
	"bytes"
	"context"
	"encoding/pem"
	"encoding/xml"
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

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscredentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	aliyuncredentials "github.com/aliyun/credentials-go/credentials"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const testOSSBucket = "test-bucket"

type ossFakeListObject struct {
	key          string
	lastModified time.Time
}

type ossFakeRequest struct {
	rawQuery      string
	rangeHeader   string
	securityToken string
	authorization string
}

type ossFakeServer struct {
	server *httptest.Server

	mu             sync.Mutex
	objects        map[string][]byte
	requests       []ossFakeRequest
	listPages      [][]ossFakeListObject
	deleteOmit     map[string]bool
	delay          time.Duration
	downloadStatus int
	downloadCode   string
}

func newOSSFakeServer(t *testing.T) *ossFakeServer {
	t.Helper()
	fake := &ossFakeServer{
		objects:    make(map[string][]byte),
		deleteOmit: make(map[string]bool),
	}
	fake.server = httptest.NewServer(http.HandlerFunc(fake.serveHTTP))
	t.Cleanup(fake.server.Close)
	return fake
}

func (f *ossFakeServer) serveHTTP(w http.ResponseWriter, r *http.Request) {
	if f.delay > 0 {
		select {
		case <-time.After(f.delay):
		case <-r.Context().Done():
			return
		}
	}

	f.mu.Lock()
	f.requests = append(f.requests, ossFakeRequest{
		rawQuery:      r.URL.RawQuery,
		rangeHeader:   r.Header.Get("Range"),
		securityToken: r.Header.Get("x-oss-security-token"),
		authorization: r.Header.Get("Authorization"),
	})
	f.mu.Unlock()

	if r.Method == http.MethodGet && r.URL.Query().Get("list-type") == "2" {
		f.serveList(w, r)
		return
	}
	if r.Method == http.MethodPost {
		if _, ok := r.URL.Query()["delete"]; ok {
			f.serveDelete(w, r)
			return
		}
	}

	prefix := "/" + testOSSBucket + "/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		writeOSSError(w, http.StatusBadRequest, "InvalidURI", "expected an IP-endpoint path-style request")
		return
	}
	key := strings.TrimPrefix(r.URL.Path, prefix)
	switch r.Method {
	case http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeOSSError(w, http.StatusBadRequest, "InvalidRequest", err.Error())
			return
		}
		f.mu.Lock()
		f.objects[key] = body
		f.mu.Unlock()
		w.Header().Set("ETag", `"fake-etag"`)
		w.Header().Set("x-oss-request-id", "put-request")
		w.WriteHeader(http.StatusOK)
	case http.MethodGet:
		f.serveDownload(w, r, key)
	default:
		writeOSSError(w, http.StatusBadRequest, "InvalidRequest", "unexpected request")
	}
}

func (f *ossFakeServer) serveDownload(w http.ResponseWriter, r *http.Request, key string) {
	if f.downloadStatus != 0 {
		writeOSSError(w, f.downloadStatus, f.downloadCode, "fake download error")
		return
	}
	f.mu.Lock()
	contents, found := f.objects[key]
	f.mu.Unlock()
	if !found {
		writeOSSError(w, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
		return
	}
	w.Header().Set("ETag", `"fake-etag"`)
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("x-oss-request-id", "get-request")
	if rangeHeader := r.Header.Get("Range"); rangeHeader != "" {
		var start, end int
		if strings.HasSuffix(rangeHeader, "-") {
			if _, err := fmt.Sscanf(rangeHeader, "bytes=%d-", &start); err != nil {
				writeOSSError(w, http.StatusRequestedRangeNotSatisfiable, "InvalidRange", err.Error())
				return
			}
			end = len(contents) - 1
		} else if _, err := fmt.Sscanf(rangeHeader, "bytes=%d-%d", &start, &end); err != nil {
			writeOSSError(w, http.StatusRequestedRangeNotSatisfiable, "InvalidRange", err.Error())
			return
		}
		if start < 0 || end < start || end >= len(contents) {
			writeOSSError(w, http.StatusRequestedRangeNotSatisfiable, "InvalidRange", "range outside object")
			return
		}
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(contents)))
		w.WriteHeader(http.StatusPartialContent)
		_, _ = w.Write(contents[start : end+1])
		return
	}
	_, _ = w.Write(contents)
}

func (f *ossFakeServer) serveList(w http.ResponseWriter, r *http.Request) {
	page := 0
	if token := r.URL.Query().Get("continuation-token"); token != "" {
		if _, err := fmt.Sscanf(token, "page-%d", &page); err != nil {
			writeOSSError(w, http.StatusBadRequest, "InvalidArgument", err.Error())
			return
		}
	}
	var objects []ossFakeListObject
	if page < len(f.listPages) {
		objects = f.listPages[page]
	}
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-oss-request-id", "list-request")
	_, _ = io.WriteString(w, "<ListBucketResult><Name>"+testOSSBucket+"</Name>")
	_, _ = io.WriteString(w, "<Prefix>"+xmlText(r.URL.Query().Get("prefix"))+"</Prefix>")
	_, _ = io.WriteString(w, "<MaxKeys>"+xmlText(r.URL.Query().Get("max-keys"))+"</MaxKeys>")
	_, _ = io.WriteString(w, "<KeyCount>"+strconv.Itoa(len(objects))+"</KeyCount>")
	if page+1 < len(f.listPages) {
		_, _ = io.WriteString(w, "<IsTruncated>true</IsTruncated><NextContinuationToken>page-"+strconv.Itoa(page+1)+"</NextContinuationToken>")
	} else {
		_, _ = io.WriteString(w, "<IsTruncated>false</IsTruncated>")
	}
	for _, object := range objects {
		modified := object.lastModified
		if modified.IsZero() {
			modified = time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
		}
		_, _ = io.WriteString(w, "<Contents><Key>"+xmlText(object.key)+"</Key><LastModified>"+modified.Format(time.RFC3339)+"</LastModified><ETag>&quot;etag&quot;</ETag><Size>1</Size><StorageClass>Standard</StorageClass></Contents>")
	}
	_, _ = io.WriteString(w, "</ListBucketResult>")
}

func (f *ossFakeServer) serveDelete(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Objects []struct {
			Key string `xml:"Key"`
		} `xml:"Object"`
	}
	if err := xml.NewDecoder(r.Body).Decode(&request); err != nil {
		writeOSSError(w, http.StatusBadRequest, "MalformedXML", err.Error())
		return
	}
	f.mu.Lock()
	for _, object := range request.Objects {
		delete(f.objects, object.Key)
	}
	omit := maps.Clone(f.deleteOmit)
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-oss-request-id", "delete-request")
	_, _ = io.WriteString(w, "<DeleteResult><EncodingType>url</EncodingType>")
	for _, object := range request.Objects {
		if !omit[object.Key] {
			_, _ = io.WriteString(w, "<Deleted><Key>"+xmlText(url.QueryEscape(object.Key))+"</Key></Deleted>")
		}
	}
	_, _ = io.WriteString(w, "</DeleteResult>")
}

func (f *ossFakeServer) lastRequest() ossFakeRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.requests) == 0 {
		return ossFakeRequest{}
	}
	return f.requests[len(f.requests)-1]
}

func xmlText(value string) string {
	var output strings.Builder
	_ = xml.EscapeText(&output, []byte(value))
	return output.String()
}

func writeOSSError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.Header().Set("x-oss-request-id", "error-request")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, "<Error><Code>"+xmlText(code)+"</Code><Message>"+xmlText(message)+"</Message><RequestId>error-request</RequestId></Error>")
}

func testOSSConfig(fake *ossFakeServer) map[string]any {
	return map[string]any{
		"bucketName":  testOSSBucket,
		"prefix":      "configured/prefix/",
		"region":      "cn-hangzhou",
		"endPoint":    fake.server.URL,
		"disableSSL":  true,
		"accessKeyID": "test-access-key-id",
		"accessKey":   "test-access-key-secret",
	}
}

func newTestOSSManager(t *testing.T, fake *ossFakeServer) *OSSManager {
	t.Helper()
	manager, err := newOSSManager(kitconfig.New(), testOSSConfig(fake), logger.NOP, func() time.Duration { return time.Minute }, ossManagerOptions{allowInsecureEndpoint: true})
	require.NoError(t, err)
	return manager
}

func productionOSSConfig() map[string]any {
	return map[string]any{
		"bucketName":  testOSSBucket,
		"prefix":      "configured/prefix/",
		"region":      "cn-hangzhou",
		"accessKeyID": "test-access-key-id",
		"accessKey":   "test-access-key-secret",
	}
}

type stubAliyunCredential struct {
	mu        sync.Mutex
	model     *aliyuncredentials.CredentialModel
	err       error
	delay     time.Duration
	active    int
	maxActive int
	calls     int
}

func (s *stubAliyunCredential) GetCredential() (*aliyuncredentials.CredentialModel, error) {
	s.mu.Lock()
	s.active++
	s.calls++
	if s.active > s.maxActive {
		s.maxActive = s.active
	}
	s.mu.Unlock()
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	s.mu.Lock()
	s.active--
	s.mu.Unlock()
	return s.model, s.err
}

type testOSSCredentialsProvider struct {
	credentials osscredentials.Credentials
	err         error
}

func (p *testOSSCredentialsProvider) GetCredentials(context.Context) (osscredentials.Credentials, error) {
	return p.credentials, p.err
}

func TestOSSConfigValidationAndRegion(t *testing.T) {
	base := productionOSSConfig()
	for _, test := range []struct {
		name string
		key  string
		want string
	}{
		{name: "bucket", key: "bucketName", want: "bucketName is required"},
		{name: "region", key: "region", want: "region is required"},
		{name: "access key ID", key: "accessKeyID", want: "accessKeyID is required"},
		{name: "access key secret", key: "accessKey", want: "accessKey is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := maps.Clone(base)
			delete(config, test.key)
			_, err := NewOSSManager(kitconfig.New(), config, logger.NOP, nil)
			require.ErrorContains(t, err, test.want)
		})
	}

	roleConfig := maps.Clone(base)
	roleConfig["roleBasedAuth"] = true
	delete(roleConfig, "accessKeyID")
	delete(roleConfig, "accessKey")
	for _, test := range []struct {
		name   string
		modify func(map[string]any, *kitconfig.Config)
		want   string
	}{
		{name: "role ARN", modify: func(c map[string]any, _ *kitconfig.Config) { c["externalID"] = "external" }, want: "iamRoleARN is required"},
		{name: "external ID", modify: func(c map[string]any, _ *kitconfig.Config) { c["iamRoleARN"] = "acs:ram::123:role/test" }, want: "externalID is required"},
		{name: "caller key ID", modify: func(c map[string]any, _ *kitconfig.Config) {
			c["iamRoleARN"], c["externalID"] = "acs:ram::123:role/test", "external"
		}, want: "callerAccessKeyID is required"},
		{name: "caller key secret", modify: func(c map[string]any, conf *kitconfig.Config) {
			c["iamRoleARN"], c["externalID"] = "acs:ram::123:role/test", "external"
			conf.Set("FileManager.ALIBABA_OSS.callerAccessKeyID", "caller-id")
		}, want: "callerAccessKeySecret is required"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := maps.Clone(roleConfig)
			conf := kitconfig.New()
			test.modify(config, conf)
			_, err := NewOSSManager(conf, config, logger.NOP, nil)
			require.ErrorContains(t, err, test.want)
		})
	}

	for configured, expected := range map[string]string{
		" cn-hangzhou ":                         "cn-hangzhou",
		"oss-cn-hangzhou":                       "cn-hangzhou",
		"oss-cn-hangzhou.aliyuncs.com":          "cn-hangzhou",
		"oss-cn-hangzhou-internal.aliyuncs.com": "cn-hangzhou",
	} {
		t.Run("normalize "+configured, func(t *testing.T) {
			config := maps.Clone(base)
			config["region"] = configured
			manager, err := NewOSSManager(kitconfig.New(), config, logger.NOP, nil)
			require.NoError(t, err)
			require.Equal(t, expected, manager.config.Region)
		})
	}
}

func TestOSSEndpointValidation(t *testing.T) {
	for _, test := range []struct {
		name   string
		modify func(map[string]any)
		want   string
	}{
		{name: "disable SSL", modify: func(c map[string]any) { c["disableSSL"] = true }, want: "disableSSL is test only"},
		{name: "http endpoint", modify: func(c map[string]any) { c["endPoint"] = "http://oss-cn-hangzhou.aliyuncs.com" }, want: "must use https"},
		{name: "arbitrary https endpoint", modify: func(c map[string]any) { c["endPoint"] = "https://collector.example.com" }, want: "must be an aliyuncs.com host"},
		{name: "missing scheme with path", modify: func(c map[string]any) { c["endPoint"] = "//oss-cn-hangzhou.aliyuncs.com" }, want: "must include https scheme"},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := productionOSSConfig()
			test.modify(config)
			_, err := NewOSSManager(kitconfig.New(), config, logger.NOP, nil)
			require.ErrorContains(t, err, test.want)
		})
	}

	config := productionOSSConfig()
	config["endPoint"] = "oss-cn-hangzhou.aliyuncs.com"
	manager, err := NewOSSManager(kitconfig.New(), config, logger.NOP, nil)
	require.NoError(t, err)
	require.Equal(t, "oss-cn-hangzhou.aliyuncs.com", manager.config.EndPoint)

	fake := newOSSFakeServer(t)
	manager, err = newOSSManager(kitconfig.New(), testOSSConfig(fake), logger.NOP, nil, ossManagerOptions{allowInsecureEndpoint: true})
	require.NoError(t, err)
	require.Equal(t, fake.server.URL, manager.config.EndPoint)
}

func TestOSSRolePrecedenceAndTokenPropagation(t *testing.T) {
	fake := newOSSFakeServer(t)
	config := testOSSConfig(fake)
	config["roleBasedAuth"] = true
	config["iamRoleARN"] = "acs:ram::123:role/test"
	config["externalID"] = "external"
	config["accessKeyID"] = "ignored-static-id"
	config["accessKey"] = "ignored-static-secret"
	conf := kitconfig.New()
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeyID", "caller-id")
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeySecret", "caller-secret")
	provider := &testOSSCredentialsProvider{credentials: osscredentials.Credentials{
		AccessKeyID: "role-id", AccessKeySecret: "role-secret", SecurityToken: "role-token",
	}}
	manager, err := newOSSManager(conf, config, logger.NOP, nil, ossManagerOptions{credentialsProvider: provider, allowInsecureEndpoint: true})
	require.NoError(t, err)
	require.Same(t, provider, manager.credentialsProvider)
	require.Equal(t, ossRoleSessionName, manager.config.RoleSessionName)

	_, err = manager.UploadReader(context.Background(), "role/object.txt", strings.NewReader("role payload"))
	require.NoError(t, err)
	request := fake.lastRequest()
	require.Equal(t, "role-token", request.securityToken)
	require.Contains(t, request.authorization, "role-id")
	require.NotContains(t, request.authorization, "ignored-static-id")
}

func TestOSSRoleCredentialsProviderFakeSTS(t *testing.T) {
	var (
		handlerMu  sync.Mutex
		form       url.Values
		handlerErr error
	)
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		var parsedForm url.Values
		if err == nil {
			parsedForm, err = url.ParseQuery(string(body))
		}
		w.Header().Set("Content-Type", "application/json")
		_, writeErr := io.WriteString(w, `{"RequestId":"request-id","Credentials":{"AccessKeyId":"role-id","AccessKeySecret":"role-secret","SecurityToken":"role-token","Expiration":"`+time.Now().Add(time.Hour).UTC().Format("2006-01-02T15:04:05Z")+`"}}`)
		handlerMu.Lock()
		defer handlerMu.Unlock()
		form = parsedForm
		if err != nil {
			handlerErr = err
		} else if writeErr != nil {
			handlerErr = writeErr
		}
	}))
	defer server.Close()

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	certificateFile := path.Join(t.TempDir(), "sts-ca.pem")
	require.NoError(t, os.WriteFile(certificateFile, certificatePEM, 0o600))
	t.Setenv("SSL_CERT_FILE", certificateFile)

	fake := newOSSFakeServer(t)
	config := testOSSConfig(fake)
	config["roleBasedAuth"] = true
	config["iamRoleARN"] = "acs:ram::123:role/test"
	config["externalID"] = "workspace-id"
	delete(config, "accessKeyID")
	delete(config, "accessKey")
	conf := kitconfig.New()
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeyID", "caller-id")
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeySecret", "caller-secret")
	conf.Set("FileManager.ALIBABA_OSS.stsEndpoint", strings.TrimPrefix(server.URL, "https://"))
	manager, err := newOSSManager(conf, config, logger.NOP, nil, ossManagerOptions{allowInsecureEndpoint: true})
	require.NoError(t, err)
	credentials, err := manager.credentialsProvider.GetCredentials(context.Background())
	require.NoError(t, err)
	require.Equal(t, "role-id", credentials.AccessKeyID)
	require.Equal(t, "role-token", credentials.SecurityToken)
	handlerMu.Lock()
	require.NoError(t, handlerErr)
	observedForm := maps.Clone(form)
	handlerMu.Unlock()
	require.Equal(t, "workspace-id", observedForm.Get("ExternalId"))
	require.Equal(t, ossRoleSessionName, observedForm.Get("RoleSessionName"))
}

func TestOSSRoleCredentialsProviderConcurrencyAndNilExpires(t *testing.T) {
	id, secret, token := "role-id", "role-secret", "security-token"
	stub := &stubAliyunCredential{
		model: &aliyuncredentials.CredentialModel{AccessKeyId: &id, AccessKeySecret: &secret, SecurityToken: &token},
		delay: 2 * time.Millisecond,
	}
	provider := &ossRoleCredentialsProvider{credential: stub}
	const workers = 24
	results := make(chan osscredentials.Credentials, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			credentials, err := provider.GetCredentials(context.Background())
			results <- credentials
			errs <- err
		})
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	for credentials := range results {
		require.Equal(t, id, credentials.AccessKeyID)
		require.Equal(t, secret, credentials.AccessKeySecret)
		require.Equal(t, token, credentials.SecurityToken)
		require.Nil(t, credentials.Expires)
	}
	stub.mu.Lock()
	require.Equal(t, workers, stub.calls)
	require.Equal(t, 1, stub.maxActive)
	stub.mu.Unlock()

	nilProvider := &ossRoleCredentialsProvider{credential: &stubAliyunCredential{model: &aliyuncredentials.CredentialModel{}}}
	credentials, err := nilProvider.GetCredentials(context.Background())
	require.NoError(t, err)
	require.Equal(t, osscredentials.Credentials{}, credentials)

	failing := &ossRoleCredentialsProvider{credential: &stubAliyunCredential{err: errors.New("STS unavailable")}}
	_, err = failing.GetCredentials(context.Background())
	require.ErrorContains(t, err, "retrieving Alibaba role credentials")
	require.ErrorContains(t, err, "STS unavailable")
}

func TestOSSUploadReaderUploadAndEmptyFile(t *testing.T) {
	fake := newOSSFakeServer(t)
	manager := newTestOSSManager(t, fake)

	uploaded, err := manager.UploadReader(context.Background(), "/reader/object with space.txt", strings.NewReader("reader payload"))
	require.NoError(t, err)
	require.Equal(t, "reader/object with space.txt", uploaded.ObjectName)
	require.Equal(t, []byte("reader payload"), fake.objects[uploaded.ObjectName])
	require.Equal(t, "https://"+testOSSBucket+".oss-cn-hangzhou.aliyuncs.com/reader/object%20with%20space.txt", uploaded.Location)

	file, err := os.Create(path.Join(t.TempDir(), "payload.txt"))
	require.NoError(t, err)
	_, err = file.WriteString("file payload")
	require.NoError(t, err)
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)
	defer func() { require.NoError(t, file.Close()) }()
	uploaded, err = manager.Upload(context.Background(), file, "one", "two")
	require.NoError(t, err)
	expectedKey := "configured/prefix/one/two/payload.txt"
	require.Equal(t, expectedKey, uploaded.ObjectName)
	require.Equal(t, []byte("file payload"), fake.objects[expectedKey])

	empty, err := os.Create(path.Join(t.TempDir(), "empty.txt"))
	require.NoError(t, err)
	defer func() { require.NoError(t, empty.Close()) }()
	uploaded, err = manager.Upload(context.Background(), empty, "empty")
	require.NoError(t, err)
	require.Equal(t, "configured/prefix/empty/empty.txt", uploaded.ObjectName)
	require.Empty(t, fake.objects[uploaded.ObjectName])

	_, err = manager.UploadReader(context.Background(), "///", strings.NewReader("invalid"))
	require.ErrorContains(t, err, "object name cannot be empty")
}

func TestOSSDownloadRangesAndErrors(t *testing.T) {
	fake := newOSSFakeServer(t)
	manager := newTestOSSManager(t, fake)
	key := "dir/object.txt"
	fake.objects[key] = []byte("0123456789abcdefghij")

	var output bytesWriterAt
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, "0123456789abcdefghij", output.String())
	require.Empty(t, fake.lastRequest().rangeHeader)

	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key, WithDownloadOffSetAndLength(5, 10)))
	require.Equal(t, "56789abcde", output.String())
	require.Equal(t, "bytes=5-14", fake.lastRequest().rangeHeader)

	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key, WithDownloadOffSet(10)))
	require.Equal(t, "abcdefghij", output.String())
	require.Equal(t, "bytes=10-", fake.lastRequest().rangeHeader)

	err := manager.Download(context.Background(), &output, "missing.txt")
	require.ErrorIs(t, err, ErrKeyNotFound)

	fake.downloadStatus = http.StatusForbidden
	fake.downloadCode = "AccessDenied"
	err = manager.Download(context.Background(), &output, key)
	var serviceError *oss.ServiceError
	require.ErrorAs(t, err, &serviceError)
	require.Equal(t, "AccessDenied", serviceError.Code)
}

func TestOSSListPaginationExhaustionAndMaxClamp(t *testing.T) {
	modified := time.Date(2025, 2, 3, 4, 5, 6, 0, time.UTC)
	fake := newOSSFakeServer(t)
	fake.listPages = [][]ossFakeListObject{
		{{key: "prefix/a", lastModified: modified}, {key: "prefix/b"}},
		{{key: "prefix/c"}},
	}
	manager := newTestOSSManager(t, fake)
	session := manager.ListFilesWithPrefix(context.Background(), "prefix/0", "/prefix/", 5000)
	first, err := session.Next()
	require.NoError(t, err)
	require.Equal(t, []string{"prefix/a", "prefix/b"}, fileInfoKeys(first))
	require.Equal(t, modified, first[0].LastModified)
	firstQuery, err := url.ParseQuery(fake.requests[0].rawQuery)
	require.NoError(t, err)
	require.Equal(t, "prefix/", firstQuery.Get("prefix"))
	require.Equal(t, "prefix/0", firstQuery.Get("start-after"))
	require.Equal(t, "1000", firstQuery.Get("max-keys"))
	require.Equal(t, "url", firstQuery.Get("encoding-type"))

	second, err := session.Next()
	require.NoError(t, err)
	require.Equal(t, []string{"prefix/c"}, fileInfoKeys(second))
	secondQuery, err := url.ParseQuery(fake.requests[1].rawQuery)
	require.NoError(t, err)
	require.Equal(t, "page-1", secondQuery.Get("continuation-token"))
	third, err := session.Next()
	require.NoError(t, err)
	require.Nil(t, third)
	require.Len(t, fake.requests, 2)

	zeroFake := newOSSFakeServer(t)
	zeroFake.listPages = [][]ossFakeListObject{{{key: "should-not-be-returned"}}}
	zeroManager := newTestOSSManager(t, zeroFake)
	for _, maxItems := range []int64{0, -10} {
		listed, err := zeroManager.ListFilesWithPrefix(context.Background(), "", "", maxItems).Next()
		require.NoError(t, err)
		require.Nil(t, listed)
	}
	require.Empty(t, zeroFake.requests, "non-positive maxItems should not issue a list request")
}

func TestOSSDeleteHappyNonexistentAndPartialReconciliation(t *testing.T) {
	fake := newOSSFakeServer(t)
	manager := newTestOSSManager(t, fake)
	fake.objects["a"] = []byte("a")
	fake.objects["b"] = []byte("b")
	require.NoError(t, manager.Delete(context.Background(), nil))
	require.Empty(t, fake.requests)
	require.NoError(t, manager.Delete(context.Background(), []string{"a", "missing", "b"}))
	require.NotContains(t, fake.objects, "a")
	require.NotContains(t, fake.objects, "b")
	query, err := url.ParseQuery(fake.lastRequest().rawQuery)
	require.NoError(t, err)
	_, hasDelete := query["delete"]
	require.True(t, hasDelete)
	require.Equal(t, "url", query.Get("encoding-type"))

	fake.objects["leading/delete.txt"] = []byte("leading")
	require.NoError(t, manager.Delete(context.Background(), []string{"/leading/delete.txt"}))
	require.NotContains(t, fake.objects, "leading/delete.txt")

	fake.deleteOmit["second"] = true
	err = manager.Delete(context.Background(), []string{"first", "second", "third"})
	require.ErrorContains(t, err, "omitted 1 requested keys: second")
}

func TestOSSDeleteURLDecodedResponseKeys(t *testing.T) {
	keys := []string{
		"directory/space key.txt",
		"directory/literal+plus.txt",
		"directory/percent%key.txt",
		"directory/日本語-данные.txt",
	}
	fake := newOSSFakeServer(t)
	manager := newTestOSSManager(t, fake)
	for _, key := range keys {
		fake.objects[key] = []byte("x")
	}
	require.NoError(t, manager.Delete(context.Background(), keys))
	for _, key := range keys {
		require.NotContains(t, fake.objects, key)
	}
}

func TestOSSLocationParserPublicInternalAndCustom(t *testing.T) {
	fake := newOSSFakeServer(t)
	manager := newTestOSSManager(t, fake)
	key := "dir/a + b%/日本語.txt"
	encoded := escapeOSSObjectName(key)
	locations := []string{
		objectURL(testOSSBucket, "cn-hangzhou", key),
		"https://" + testOSSBucket + ".oss-cn-hangzhou-internal.aliyuncs.com/" + encoded,
		fake.server.URL + "/" + testOSSBucket + "/" + encoded + "?signature=secret#fragment",
	}
	for _, location := range locations {
		actual, err := manager.GetObjectNameFromLocation(location)
		require.NoError(t, err)
		require.Equal(t, key, actual)
		require.Equal(t, key, manager.GetDownloadKeyFromFileLocation(location))
	}

	_, err := manager.GetObjectNameFromLocation("https://example.com/%zz")
	require.ErrorContains(t, err, "parsing Alibaba OSS location")
}

func TestOSSMapErrorNoSuchKeyAndRedaction(t *testing.T) {
	require.Nil(t, mapOSSError(nil))
	plain := errors.New("plain error")
	require.Same(t, plain, mapOSSError(plain))

	missing := &oss.ServiceError{Code: "NoSuchKey", StatusCode: http.StatusNotFound}
	require.ErrorIs(t, mapOSSError(missing), ErrKeyNotFound)

	serviceError := &oss.ServiceError{
		Code:          "AccessDenied",
		Message:       "denied",
		StatusCode:    http.StatusForbidden,
		Snapshot:      []byte("response-secret"),
		Headers:       http.Header{"Authorization": []string{"header-secret"}},
		RequestTarget: "GET https://example.com/" + testOSSBucket + "/key?security-token=query-secret&x=1",
	}
	mapped := mapOSSError(serviceError)
	var sanitized *oss.ServiceError
	require.ErrorAs(t, mapped, &sanitized)
	require.NotSame(t, serviceError, sanitized)
	require.Nil(t, sanitized.Snapshot)
	require.Nil(t, sanitized.Headers)
	require.Equal(t, "GET https://example.com/"+testOSSBucket+"/key", sanitized.RequestTarget)
	require.NotContains(t, mapped.Error(), "response-secret")
	require.NotContains(t, mapped.Error(), "header-secret")
	require.NotContains(t, mapped.Error(), "query-secret")
	require.Equal(t, []byte("response-secret"), serviceError.Snapshot, "mapping must not mutate the SDK error")

	require.Equal(t, "POST /bucket/key", stripOSSRequestQuery("POST /bucket/key?token=secret"))
	require.Equal(t, "/bucket/key", stripOSSRequestQuery("/bucket/key?token=secret"))
}

func TestOSSTimeout(t *testing.T) {
	fake := newOSSFakeServer(t)
	fake.delay = 100 * time.Millisecond
	manager := newTestOSSManager(t, fake)
	manager.SetTimeout(time.Millisecond)
	_, err := manager.UploadReader(context.Background(), "slow.txt", strings.NewReader("payload"))
	require.Error(t, err)
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestOSSRoleCredentialsProviderHonorsContextDeadline(t *testing.T) {
	id, secret := "role-id", "role-secret"
	provider := &ossRoleCredentialsProvider{credential: &stubAliyunCredential{
		model: &aliyuncredentials.CredentialModel{AccessKeyId: &id, AccessKeySecret: &secret},
		delay: 100 * time.Millisecond,
	}}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	_, err := provider.GetCredentials(ctx)
	require.ErrorIs(t, err, context.DeadlineExceeded)

	fake := newOSSFakeServer(t)
	config := testOSSConfig(fake)
	config["roleBasedAuth"] = true
	config["iamRoleARN"] = "acs:ram::123:role/test"
	config["externalID"] = "external"
	conf := kitconfig.New()
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeyID", "caller-id")
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeySecret", "caller-secret")
	manager, err := newOSSManager(conf, config, logger.NOP, nil, ossManagerOptions{credentialsProvider: provider, allowInsecureEndpoint: true})
	require.NoError(t, err)
	manager.SetTimeout(time.Millisecond)
	_, err = manager.UploadReader(context.Background(), "role-timeout.txt", strings.NewReader("payload"))
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestOSSRoleCredentialsProviderDelayedFakeSTSHonorsManagerTimeout(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"RequestId":"request-id","Credentials":{"AccessKeyId":"role-id","AccessKeySecret":"role-secret","SecurityToken":"role-token","Expiration":"`+time.Now().Add(time.Hour).UTC().Format("2006-01-02T15:04:05Z")+`"}}`)
	}))
	defer server.Close()

	certificatePEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	certificateFile := path.Join(t.TempDir(), "sts-ca.pem")
	require.NoError(t, os.WriteFile(certificateFile, certificatePEM, 0o600))
	t.Setenv("SSL_CERT_FILE", certificateFile)

	fake := newOSSFakeServer(t)
	config := testOSSConfig(fake)
	config["roleBasedAuth"] = true
	config["iamRoleARN"] = "acs:ram::123:role/test"
	config["externalID"] = "external"
	delete(config, "accessKeyID")
	delete(config, "accessKey")
	conf := kitconfig.New()
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeyID", "caller-id")
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeySecret", "caller-secret")
	conf.Set("FileManager.ALIBABA_OSS.stsEndpoint", strings.TrimPrefix(server.URL, "https://"))
	manager, err := newOSSManager(conf, config, logger.NOP, nil, ossManagerOptions{allowInsecureEndpoint: true})
	require.NoError(t, err)
	manager.SetTimeout(time.Millisecond)
	_, err = manager.UploadReader(context.Background(), "role-fake-sts-timeout.txt", strings.NewReader("payload"))
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestOSSFactoryAndEnvironmentConfig(t *testing.T) {
	conf := kitconfig.New()
	conf.Set("FileManager.ALIBABA_OSS.timeout", "3s")
	factoryConfig := productionOSSConfig()
	factoryConfig["endPoint"] = "https://oss-cn-hangzhou.aliyuncs.com"
	manager, err := New(&Settings{Provider: "ALIBABA_OSS", Config: factoryConfig, Logger: logger.NOP, Conf: conf})
	require.NoError(t, err)
	require.IsType(t, &OSSManager{}, manager)
	require.Equal(t, 3*time.Second, manager.(*OSSManager).getTimeout())
	require.Equal(t, "configured/prefix/", manager.Prefix())

	t.Setenv("ALIBABA_OSS_REGION", "cn-shanghai")
	t.Setenv("ALIBABA_OSS_ENDPOINT", "https://custom.example")
	t.Setenv("ALIBABA_OSS_ACCESS_KEY_ID", "env-id")
	t.Setenv("ALIBABA_OSS_ACCESS_KEY_SECRET", "env-secret")
	environment := GetProviderConfigFromEnv(ProviderConfigOpts{
		Provider: "ALIBABA_OSS", Bucket: "env-bucket", Prefix: "env-prefix", Config: kitconfig.New(),
	})
	require.Equal(t, "env-bucket", environment["bucketName"])
	require.Equal(t, "env-prefix", environment["prefix"])
	require.Equal(t, "cn-shanghai", environment["region"])
	require.Equal(t, "https://custom.example", environment["endPoint"])
	require.Equal(t, "env-id", environment["accessKeyID"])
	require.Equal(t, "env-secret", environment["accessKey"])
}

func TestOSSLive(t *testing.T) {
	manager := newOSSLiveStaticManager(t)
	prefix := "rudder-go-kit-test/" + uuid.NewString()
	key := prefix + "/roundtrip.txt"
	verboseExisting := prefix + "/verbose-existing.txt"
	encodedKeys := []string{
		prefix + "/space key.txt",
		prefix + "/literal+plus.txt",
		prefix + "/percent%key.txt",
		prefix + "/日本語.txt",
	}
	cleanupKeys := append([]string{key, verboseExisting}, encodedKeys...)
	t.Cleanup(func() { _ = manager.Delete(context.Background(), cleanupKeys) })

	contents := []byte("0123456789abcdefghij")
	uploaded, err := manager.UploadReader(context.Background(), key, bytes.NewReader(contents))
	require.NoError(t, err)
	require.Equal(t, key, uploaded.ObjectName)
	listed, err := manager.ListFilesWithPrefix(context.Background(), "", prefix+"/round", 10).Next()
	require.NoError(t, err)
	require.Equal(t, []string{key}, fileInfoKeys(listed))

	var output bytesWriterAt
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, contents, output.Bytes())
	output = bytesWriterAt{}
	require.NoError(t, manager.Download(context.Background(), &output, key, WithDownloadOffSetAndLength(5, 10)))
	require.Equal(t, contents[5:15], output.Bytes())
	objectName, err := manager.GetObjectNameFromLocation(uploaded.Location)
	require.NoError(t, err)
	require.Equal(t, key, objectName)

	_, err = manager.UploadReader(context.Background(), key, strings.NewReader("overwritten"))
	require.NoError(t, err)
	_, err = manager.UploadReader(context.Background(), prefix+"/empty.txt", strings.NewReader(""))
	require.NoError(t, err)
	cleanupKeys = append(cleanupKeys, prefix+"/empty.txt")

	_, err = manager.UploadReader(context.Background(), verboseExisting, strings.NewReader("verbose"))
	require.NoError(t, err)
	require.NoError(t, manager.Delete(context.Background(), []string{verboseExisting, prefix + "/verbose-never-existed"}),
		"verbose delete must return both existing and nonexistent keys for reconciliation")
	for _, encodedKey := range encodedKeys {
		_, err = manager.UploadReader(context.Background(), encodedKey, strings.NewReader("encoded"))
		require.NoError(t, err)
	}
	require.NoError(t, manager.Delete(context.Background(), encodedKeys), "live OSS key encoding must round trip through verbose delete")

	require.NoError(t, manager.Delete(context.Background(), []string{key}))
	require.ErrorIs(t, manager.Download(context.Background(), &output, key), ErrKeyNotFound)
}

func TestOSSLiveErrors(t *testing.T) {
	config := ossLiveStaticConfig(t)
	config["accessKey"] = "deliberately-wrong-secret"
	manager, err := NewOSSManager(kitconfig.New(), config, logger.NOP, func() time.Duration { return 30 * time.Second })
	require.NoError(t, err)
	_, err = manager.UploadReader(context.Background(), "rudder-go-kit-test/"+uuid.NewString()+"/auth.txt", strings.NewReader("payload"))
	require.Error(t, err)
	require.NotContains(t, err.Error(), "deliberately-wrong-secret")
}

func TestOSSLiveRole(t *testing.T) {
	manager := newOSSLiveRoleManager(t)
	key := "rudder-go-kit-test/" + uuid.NewString() + "/role.txt"
	t.Cleanup(func() { _ = manager.Delete(context.Background(), []string{key}) })
	_, err := manager.UploadReader(context.Background(), key, strings.NewReader("role payload"))
	require.NoError(t, err)
	listed, err := manager.ListFilesWithPrefix(context.Background(), "", key, 10).Next()
	require.NoError(t, err)
	require.Equal(t, []string{key}, fileInfoKeys(listed))
	var output bytesWriterAt
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, "role payload", output.String())
	require.NoError(t, manager.Delete(context.Background(), []string{key}))

	wrongConfig := map[string]any{
		"bucketName":    requiredOSSEnv(t, "OSS_TEST_BUCKET"),
		"region":        requiredOSSEnv(t, "OSS_TEST_REGION"),
		"roleBasedAuth": true,
		"iamRoleARN":    requiredOSSEnv(t, "OSS_TEST_ROLE_ARN"),
		"externalID":    "deliberately-wrong-external-id",
	}
	wrongManager, err := NewOSSManager(ossLiveRoleKitConfig(t), wrongConfig, logger.NOP, func() time.Duration { return 30 * time.Second })
	require.NoError(t, err)
	_, err = wrongManager.UploadReader(context.Background(), key, strings.NewReader("must fail"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "NoPermission")
}

func TestOSSLiveRoleRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Alibaba OSS role refresh test is intentionally long")
	}
	if os.Getenv("OSS_TEST_ROLE_REFRESH") == "" {
		t.Skip("Alibaba OSS role refresh test requires OSS_TEST_ROLE_REFRESH")
	}
	const wait = 20 * time.Minute
	manager := newOSSLiveRoleManager(t)
	key := "rudder-go-kit-test/" + uuid.NewString() + "/refresh.txt"
	t.Cleanup(func() { _ = manager.Delete(context.Background(), []string{key}) })
	_, err := manager.UploadReader(context.Background(), key, strings.NewReader("refresh payload"))
	require.NoError(t, err)
	t.Logf("waiting %s so credentials-go must refresh the role session", wait)
	time.Sleep(wait)
	var output bytesWriterAt
	require.NoError(t, manager.Download(context.Background(), &output, key))
	require.Equal(t, "refresh payload", output.String())
}

func newOSSLiveStaticManager(t *testing.T) *OSSManager {
	t.Helper()
	manager, err := NewOSSManager(kitconfig.New(), ossLiveStaticConfig(t), logger.NOP, func() time.Duration { return 2 * time.Minute })
	require.NoError(t, err)
	return manager
}

func ossLiveStaticConfig(t *testing.T) map[string]any {
	t.Helper()
	config := map[string]any{
		"bucketName":  requiredOSSEnv(t, "OSS_TEST_BUCKET"),
		"region":      requiredOSSEnv(t, "OSS_TEST_REGION"),
		"accessKeyID": requiredOSSEnv(t, "OSS_TEST_ACCESS_KEY_ID"),
		"accessKey":   requiredOSSEnv(t, "OSS_TEST_ACCESS_KEY_SECRET"),
	}
	if endpoint := os.Getenv("OSS_TEST_ENDPOINT"); endpoint != "" {
		config["endPoint"] = endpoint
	}
	return config
}

func newOSSLiveRoleManager(t *testing.T) *OSSManager {
	t.Helper()
	config := map[string]any{
		"bucketName":    requiredOSSEnv(t, "OSS_TEST_BUCKET"),
		"region":        requiredOSSEnv(t, "OSS_TEST_REGION"),
		"roleBasedAuth": true,
		"iamRoleARN":    requiredOSSEnv(t, "OSS_TEST_ROLE_ARN"),
		"externalID":    requiredOSSEnv(t, "OSS_TEST_EXTERNAL_ID"),
	}
	if endpoint := os.Getenv("OSS_TEST_ENDPOINT"); endpoint != "" {
		config["endPoint"] = endpoint
	}
	manager, err := NewOSSManager(ossLiveRoleKitConfig(t), config, logger.NOP, func() time.Duration { return 2 * time.Minute })
	require.NoError(t, err)
	return manager
}

func ossLiveRoleKitConfig(t *testing.T) *kitconfig.Config {
	t.Helper()
	conf := kitconfig.New()
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeyID", requiredOSSEnv(t, "OSS_TEST_CALLER_ACCESS_KEY_ID"))
	conf.Set("FileManager.ALIBABA_OSS.callerAccessKeySecret", requiredOSSEnv(t, "OSS_TEST_CALLER_ACCESS_KEY_SECRET"))
	if stsEndpoint := os.Getenv("OSS_TEST_STS_ENDPOINT"); stsEndpoint != "" {
		conf.Set("FileManager.ALIBABA_OSS.stsEndpoint", stsEndpoint)
	}
	return conf
}

func requiredOSSEnv(t *testing.T, name string) string {
	t.Helper()
	value := os.Getenv(name)
	if value == "" {
		t.Skip("Alibaba OSS live test requires " + name)
	}
	return value
}
