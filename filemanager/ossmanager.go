package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	osscredentials "github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	aliyuncredentials "github.com/aliyun/credentials-go/credentials"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	kitconfig "github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
)

const (
	ossRoleSessionName                 = "rudderstack-filemanager"
	ossDefaultRoleSessionExpiration    = 3600
	ossRoleSTSReadTimeoutMaxSeconds    = 30
	ossRoleSTSConnectTimeoutMaxSeconds = 5
	ossDeleteBatchSize                 = 1000
	ossMaxListItems                    = 1000
)

// OSSConfig contains the destination settings for Alibaba Cloud OSS.
type OSSConfig struct {
	Bucket      string `mapstructure:"bucketName"`
	Prefix      string `mapstructure:"prefix"`
	Region      string `mapstructure:"region"`
	EndPoint    string `mapstructure:"endPoint"`
	UseInternal bool   `mapstructure:"useInternalEndpoint"`
	// DisableSSL is test only; never set in a customer config because it sends credentials in clear text.
	DisableSSL bool `mapstructure:"disableSSL"`

	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKey"`

	RoleBasedAuth   bool   `mapstructure:"roleBasedAuth"`
	IAMRoleARN      string `mapstructure:"iamRoleARN"`
	ExternalID      string `mapstructure:"externalID"`
	RoleSessionName string `mapstructure:"-"` // set in code only, never from destination config
}

type aliyunCredential interface {
	GetCredential() (*aliyuncredentials.CredentialModel, error)
}

// ossRoleCredentialsProvider adapts credentials-go's refreshable role credentials to the OSS SDK.
// credentials-go has mutable refresh state but no internal lock, so calls are serialized here.
type ossRoleCredentialsProvider struct {
	mu         sync.Mutex
	credential aliyunCredential
}

func (p *ossRoleCredentialsProvider) GetCredentials(ctx context.Context) (osscredentials.Credentials, error) {
	if err := ctx.Err(); err != nil {
		return osscredentials.Credentials{}, fmt.Errorf("retrieving Alibaba role credentials: %w", err)
	}

	p.mu.Lock()
	if err := ctx.Err(); err != nil {
		p.mu.Unlock()
		return osscredentials.Credentials{}, fmt.Errorf("retrieving Alibaba role credentials: %w", err)
	}
	credential, err := p.credential.GetCredential()
	p.mu.Unlock()
	if err != nil {
		// STS signs in the URL query; a transport error includes that URL in its message.
		if requestErr, ok := errors.AsType[*url.Error](err); ok {
			sanitized := *requestErr
			sanitized.URL = stripOSSRequestQuery(requestErr.URL)
			err = &sanitized
		}
		return osscredentials.Credentials{}, fmt.Errorf("retrieving Alibaba role credentials (check credentials and host clock): %w", err)
	}
	if err := ctx.Err(); err != nil {
		return osscredentials.Credentials{}, fmt.Errorf("retrieving Alibaba role credentials: %w", err)
	}
	return osscredentials.Credentials{
		AccessKeyID:     oss.ToString(credential.AccessKeyId),
		AccessKeySecret: oss.ToString(credential.AccessKeySecret),
		SecurityToken:   oss.ToString(credential.SecurityToken),
		// Expires deliberately remains nil. credentials-go owns caching and refresh.
	}, nil
}

type ossManagerOptions struct {
	credentialsProvider osscredentials.CredentialsProvider
	// allowInsecureEndpoint is for package tests only; customer config must never force plaintext OSS traffic.
	allowInsecureEndpoint bool
}

// OSSManager manages files in Alibaba Cloud Object Storage Service.
// UploadedFile.Location is a stable public identifier for recovering the object key; it is not
// guaranteed to be reachable when the client uses an internal endpoint or a non-default
// *.aliyuncs.com endpoint.
type OSSManager struct {
	*baseManager
	config *OSSConfig

	credentialsProvider osscredentials.CredentialsProvider
	client              *oss.Client
	clientOnce          sync.Once
}

var _ FileManager = (*OSSManager)(nil)

// NewOSSManager creates an Alibaba Cloud OSS file manager.
func NewOSSManager(
	kitconf *kitconfig.Config, config map[string]any,
	log logger.Logger, defaultTimeout func() time.Duration,
) (*OSSManager, error) {
	return newOSSManager(kitconf, config, log, defaultTimeout, ossManagerOptions{})
}

func newOSSManager(
	kitconf *kitconfig.Config, config map[string]any,
	log logger.Logger, defaultTimeout func() time.Duration, options ossManagerOptions,
) (*OSSManager, error) {
	var ossConfig OSSConfig
	if err := mapstructure.Decode(config, &ossConfig); err != nil {
		return nil, fmt.Errorf("decoding Alibaba OSS config: %w", err)
	}
	ossConfig.Bucket = strings.TrimRight(strings.TrimSpace(ossConfig.Bucket), "/")
	ossConfig.Prefix = sanitizeKey(ossConfig.Prefix)
	ossConfig.Region = normalizeOSSRegion(ossConfig.Region)
	ossConfig.RoleSessionName = ossRoleSessionName

	if ossConfig.Bucket == "" {
		return nil, errors.New("bucketName is required for Alibaba OSS")
	}
	if ossConfig.Region == "" {
		return nil, errors.New("region is required for Alibaba OSS")
	}
	if kitconf == nil {
		kitconf = kitconfig.Default
	}
	if log == nil {
		log = logger.NewLogger().Child("filemanager.oss")
	}
	if originalRegion, _ := config["region"].(string); originalRegion != "" && strings.TrimSpace(originalRegion) != ossConfig.Region {
		log.Warnn("Normalizing Alibaba OSS region",
			logger.NewStringField("configuredRegion", originalRegion),
			logger.NewStringField("normalizedRegion", ossConfig.Region),
		)
	}
	if err := validateOSSEndpointConfig(&ossConfig, options.allowInsecureEndpoint); err != nil {
		return nil, err
	}

	provider := options.credentialsProvider
	if ossConfig.RoleBasedAuth {
		if ossConfig.IAMRoleARN == "" {
			return nil, errors.New("iamRoleARN is required for IAM role")
		}
		if ossConfig.ExternalID == "" {
			return nil, errors.New("externalID is required for IAM role")
		}
		callerAccessKeyID := kitconf.GetStringVar("", "FileManager.ALIBABA_OSS.callerAccessKeyID")
		if callerAccessKeyID == "" {
			return nil, errors.New("FileManager.ALIBABA_OSS.callerAccessKeyID is required for IAM role")
		}
		callerAccessKeySecret := kitconf.GetStringVar("", "FileManager.ALIBABA_OSS.callerAccessKeySecret")
		if callerAccessKeySecret == "" {
			return nil, errors.New("FileManager.ALIBABA_OSS.callerAccessKeySecret is required for IAM role")
		}
		if ossConfig.AccessKeyID != "" || ossConfig.AccessKeySecret != "" {
			log.Warnn("Ignoring static Alibaba OSS credentials because role-based authentication is enabled")
		}
		roleSessionExpiration := kitconf.GetIntVar(ossDefaultRoleSessionExpiration, 1, "FileManager.ALIBABA_OSS.roleSessionExpiration")
		if roleSessionExpiration < 900 || roleSessionExpiration > 3600 {
			return nil, fmt.Errorf("FileManager.ALIBABA_OSS.roleSessionExpiration must be between 900 and 3600 seconds (got %d)", roleSessionExpiration)
		}
		if provider == nil {
			credentialConfig := new(aliyuncredentials.Config).
				SetType("ram_role_arn").
				SetAccessKeyId(callerAccessKeyID).
				SetAccessKeySecret(callerAccessKeySecret).
				SetRoleArn(ossConfig.IAMRoleARN).
				SetRoleSessionName(ossConfig.RoleSessionName).
				SetRoleSessionExpiration(roleSessionExpiration).
				SetTimeout(ossCredentialTimeoutSeconds(defaultTimeout, ossRoleSTSReadTimeoutMaxSeconds)).
				SetConnectTimeout(ossCredentialTimeoutSeconds(defaultTimeout, ossRoleSTSConnectTimeoutMaxSeconds))
			credentialConfig.ExternalId = &ossConfig.ExternalID
			if endpoint := kitconf.GetStringVar("", "FileManager.ALIBABA_OSS.stsEndpoint"); endpoint != "" {
				credentialConfig.SetSTSEndpoint(endpoint)
			}
			credential, err := aliyuncredentials.NewCredential(credentialConfig)
			if err != nil {
				return nil, fmt.Errorf("creating Alibaba role credential: %w", err)
			}
			provider = &ossRoleCredentialsProvider{credential: credential}
		}
	} else {
		if ossConfig.AccessKeyID == "" {
			return nil, errors.New("accessKeyID is required for Alibaba OSS")
		}
		if ossConfig.AccessKeySecret == "" {
			return nil, errors.New("accessKey is required for Alibaba OSS")
		}
		if provider == nil {
			provider = osscredentials.NewStaticCredentialsProvider(ossConfig.AccessKeyID, ossConfig.AccessKeySecret)
		}
	}

	return &OSSManager{
		baseManager:         &baseManager{logger: log, defaultTimeout: defaultTimeout},
		config:              &ossConfig,
		credentialsProvider: provider,
	}, nil
}

func normalizeOSSRegion(region string) string {
	region = strings.TrimSpace(region)
	region = strings.TrimPrefix(region, "oss-")
	region = strings.TrimSuffix(region, ".aliyuncs.com")
	region = strings.TrimSuffix(region, "-internal")
	return region
}

func validateOSSEndpointConfig(config *OSSConfig, allowInsecureEndpoint bool) error {
	if config.DisableSSL && !allowInsecureEndpoint {
		return errors.New("disableSSL is test only and cannot be set for Alibaba OSS")
	}
	config.EndPoint = strings.TrimSpace(config.EndPoint)
	if config.EndPoint == "" {
		return nil
	}
	parsed, err := url.Parse(config.EndPoint)
	if err != nil {
		return fmt.Errorf("parsing Alibaba OSS endPoint: %w", err)
	}
	if parsed.Scheme == "" {
		if parsed.Host != "" {
			return errors.New("alibaba OSS endPoint must include https scheme")
		}
		// The OSS SDK adds https:// to scheme-less host names when DisableSSL is false.
		parsed, err = url.Parse("https://" + config.EndPoint)
		if err != nil {
			return fmt.Errorf("parsing Alibaba OSS endPoint: %w", err)
		}
	}
	if parsed.User != nil || (parsed.Path != "" && parsed.Path != "/") || parsed.RawQuery != "" || parsed.ForceQuery {
		return errors.New("alibaba OSS endPoint must not include userinfo, a path, or a query")
	}
	config.EndPoint = strings.TrimSuffix(config.EndPoint, "/")
	if allowInsecureEndpoint && parsed.Scheme == "http" {
		return nil
	}
	if parsed.Scheme != "https" {
		return fmt.Errorf("alibaba OSS endPoint must use https: %s", config.EndPoint)
	}
	return validateOSSAlibabaEndpointHost(parsed.Hostname())
}

func validateOSSAlibabaEndpointHost(host string) error {
	if host == "" {
		return errors.New("alibaba OSS endPoint host is required")
	}
	if host == "aliyuncs.com" || strings.HasSuffix(host, ".aliyuncs.com") {
		return nil
	}
	return fmt.Errorf("alibaba OSS endPoint must be an aliyuncs.com host: %s", host)
}

func ossCredentialTimeoutSeconds(defaultTimeoutFn func() time.Duration, maxSeconds int) int {
	timeout := defaultTimeout
	if defaultTimeoutFn != nil {
		timeout = defaultTimeoutFn()
	}
	seconds := int((timeout + time.Second - 1) / time.Second)
	if seconds < 1 {
		return 1
	}
	return min(seconds, maxSeconds)
}

func (m *OSSManager) getClient() *oss.Client {
	m.clientOnce.Do(func() {
		config := oss.LoadDefaultConfig().
			WithRegion(m.config.Region).
			WithCredentialsProvider(m.credentialsProvider)
		if m.config.EndPoint != "" {
			config.WithEndpoint(m.config.EndPoint)
		} else if m.config.UseInternal {
			config.WithUseInternalEndpoint(true)
		}
		if m.config.DisableSSL {
			config.WithDisableSSL(true)
		}
		m.client = oss.NewClient(config, func(options *oss.Options) {
			options.HttpClient = &ossRequestBodyClient{client: options.HttpClient}
		})
	})
	return m.client
}

// The OSS SDK retries seekable request bodies after an early HTTP error. net/http can
// still be reading the previous attempt when Do returns, so wait for that body's
// Close before a retry resets its reader and CRC state. A canceled request cannot
// retry and must not wait indefinitely for Close.
type ossRequestBodyClient struct {
	client oss.HTTPClient
}

func (c *ossRequestBodyClient) Do(request *http.Request) (*http.Response, error) {
	if request.Body == nil || request.Body == http.NoBody {
		return c.client.Do(request)
	}
	seekable, ok := request.Body.(interface{ IsSeekable() bool })
	if !ok || !seekable.IsSeekable() {
		return c.client.Do(request)
	}
	body := &ossRequestBody{ReadCloser: request.Body, done: make(chan struct{})}
	attempt := request.Clone(request.Context())
	attempt.Body = body
	response, err := c.client.Do(attempt)
	select {
	case <-body.done:
	case <-request.Context().Done():
	}
	return response, err
}

type ossRequestBody struct {
	io.ReadCloser
	mu     sync.Mutex
	reads  sync.WaitGroup
	closed bool
	done   chan struct{}
}

func (b *ossRequestBody) Read(p []byte) (int, error) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return 0, io.EOF
	}
	b.reads.Add(1)
	b.mu.Unlock()
	defer b.reads.Done()
	return b.ReadCloser.Read(p)
}

func (b *ossRequestBody) Close() error {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil
	}
	b.closed = true
	b.mu.Unlock()
	err := b.ReadCloser.Close()
	b.reads.Wait()
	close(b.done)
	return err
}

func (m *OSSManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

func (m *OSSManager) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	objName = sanitizeKey(objName)
	if objName == "" {
		return UploadedFile{}, errors.New("object name cannot be empty")
	}
	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	request := &oss.PutObjectRequest{Bucket: new(m.config.Bucket), Key: new(objName), Body: rdr}
	var err error
	if _, ok := rdr.(*os.File); ok {
		_, err = oss.NewUploader(m.getClient()).UploadFrom(ctx, request, rdr)
	} else {
		_, err = m.getClient().PutObject(ctx, request)
	}
	if err != nil {
		return UploadedFile{}, mapOSSError(err)
	}
	return UploadedFile{Location: objectURL(m.config.Bucket, m.config.Region, objName), ObjectName: objName}, nil
}

// objectURL returns a public OSS identifier from which the key can be recovered. It is not a
// reachability guarantee for managers using internal or non-default *.aliyuncs.com endpoints.
func objectURL(bucket, region, objectName string) string {
	return "https://" + bucket + ".oss-" + region + ".aliyuncs.com/" + escapeOSSObjectName(objectName)
}

func escapeOSSObjectName(objectName string) string {
	segments := strings.Split(objectName, "/")
	for i := range segments {
		segments[i] = url.PathEscape(segments[i])
	}
	return strings.Join(segments, "/")
}

func (m *OSSManager) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	key = sanitizeKey(key)
	request := &oss.GetObjectRequest{Bucket: new(m.config.Bucket), Key: new(key)}
	downloadOpts := applyDownloadOptions(opts...)
	if downloadOpts.isRangeRequest {
		if downloadOpts.length > 0 {
			request.Range = new(fmt.Sprintf("bytes=%d-%d", downloadOpts.offset, downloadOpts.offset+downloadOpts.length-1))
		} else {
			request.Range = new(fmt.Sprintf("bytes=%d-", downloadOpts.offset))
		}
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()
	result, err := m.getClient().GetObject(ctx, request)
	if err != nil {
		return mapOSSError(err)
	}
	_, copyErr := io.Copy(&writerAtAdapter{w: output}, result.Body)
	closeErr := result.Body.Close()
	if copyErr != nil {
		return fmt.Errorf("copying Alibaba OSS object %s: %w", key, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing Alibaba OSS object %s: %w", key, closeErr)
	}
	return nil
}

func (m *OSSManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	maxListItems := min(max(maxItems, 0), ossMaxListItems)
	return &ossListSession{
		baseListSession: &baseListSession{
			ctx: ctx, startAfter: startAfter, prefix: sanitizeKey(prefix), maxItems: maxListItems,
		},
		manager: m, isTruncated: maxListItems > 0,
	}
}

type ossListSession struct {
	*baseListSession
	manager *OSSManager

	continuationToken *string
	isTruncated       bool
}

func (l *ossListSession) Next() ([]*FileInfo, error) {
	if !l.isTruncated {
		return nil, nil
	}
	request := &oss.ListObjectsV2Request{
		Bucket: new(l.manager.config.Bucket), Prefix: new(l.prefix), MaxKeys: int32(l.maxItems),
		ContinuationToken: l.continuationToken,
	}
	if l.startAfter != "" && l.continuationToken == nil {
		request.StartAfter = new(l.startAfter)
	}
	ctx, cancel := context.WithTimeout(l.ctx, l.manager.getTimeout())
	defer cancel()
	result, err := l.manager.getClient().ListObjectsV2(ctx, request)
	if err != nil {
		return nil, mapOSSError(err)
	}
	l.isTruncated = result.IsTruncated
	l.continuationToken = result.NextContinuationToken
	files := make([]*FileInfo, 0, len(result.Contents))
	for _, object := range result.Contents {
		files = append(files, &FileInfo{Key: oss.ToString(object.Key), LastModified: oss.ToTime(object.LastModified)})
	}
	return files, nil
}

// Delete removes OSS objects. On versioned buckets OSS writes delete markers; whether verbose
// responses include every delete-markered key, as reconciliation assumes, has not been verified.
func (m *OSSManager) Delete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	for _, chunk := range lo.Chunk(keys, ossDeleteBatchSize) {
		objects := make([]oss.ObjectIdentifier, 0, len(chunk))
		requested := make(map[string]struct{}, len(chunk))
		for _, key := range chunk {
			key = sanitizeKey(key)
			objects = append(objects, oss.ObjectIdentifier{Key: new(key)})
			requested[key] = struct{}{}
		}
		deleteCtx, cancel := context.WithTimeout(ctx, m.getTimeout())
		result, err := m.getClient().DeleteMultipleObjects(deleteCtx, &oss.DeleteMultipleObjectsRequest{
			Bucket: new(m.config.Bucket),
			Delete: &oss.Delete{
				Objects: objects,
				// Quiet must remain false: the live service returns no Deleted entries in quiet mode,
				// leaving no data for partial-failure reconciliation (unlike SDK canned fixtures).
				Quiet: false,
			},
		})
		cancel()
		if err != nil {
			mapped := mapOSSError(err)
			m.logOSSError("Alibaba OSS object delete failed", mapped)
			return mapped
		}
		for _, deleted := range result.DeletedObjects {
			delete(requested, oss.ToString(deleted.Key))
		}
		if len(requested) > 0 {
			missing := make([]string, 0, len(requested))
			for key := range requested {
				missing = append(missing, key)
			}
			sort.Strings(missing)
			shown := missing[:min(len(missing), 10)]
			return fmt.Errorf("alibaba OSS delete response omitted %d requested keys: %s", len(missing), strings.Join(shown, ", "))
		}
	}
	return nil
}

func (m *OSSManager) logOSSError(message string, err error) {
	if serviceErr, ok := errors.AsType[*oss.ServiceError](err); ok {
		m.logger.Errorn(message, obskit.Error(errors.New(err.Error())),
			logger.NewStringField("errorCode", serviceErr.Code),
			logger.NewStringField("errorMessage", serviceErr.Message),
		)
		return
	}
	m.logger.Errorn(message, obskit.Error(errors.New(err.Error())))
}

func (m *OSSManager) Prefix() string { return m.config.Prefix }

func (m *OSSManager) GetObjectNameFromLocation(location string) (string, error) {
	parsed, err := url.Parse(location)
	if err != nil {
		// url.Error includes the full location, which may carry signed query credentials.
		if parseErr, ok := errors.AsType[*url.Error](err); ok {
			err = parseErr.Err
		}
		return "", fmt.Errorf("parsing Alibaba OSS location: %w", err)
	}
	key := strings.TrimLeft(parsed.EscapedPath(), "/")
	if !strings.HasPrefix(parsed.Hostname(), m.config.Bucket+".") {
		key = strings.TrimPrefix(key, url.PathEscape(m.config.Bucket)+"/")
	}
	decoded, err := url.PathUnescape(key)
	if err != nil {
		return "", fmt.Errorf("decoding Alibaba OSS location path: %w", err)
	}
	return decoded, nil
}

func (m *OSSManager) GetDownloadKeyFromFileLocation(location string) string {
	key, err := m.GetObjectNameFromLocation(location)
	if err != nil {
		m.logger.Errorn("error while parsing Alibaba OSS location", obskit.Error(err))
		return ""
	}
	return key
}

// mapOSSError strips response snapshots, headers, and URL queries so returned errors are
// structurally safe to log even if a future authentication method places credentials in a query.
func mapOSSError(err error) error {
	if err == nil {
		return nil
	}
	if serviceErr, ok := errors.AsType[*oss.ServiceError](err); ok {
		if serviceErr.Code == "NoSuchKey" {
			return ErrKeyNotFound
		}
		sanitized := *serviceErr
		sanitized.Snapshot = nil
		sanitized.Headers = nil
		sanitized.RequestTarget = stripOSSRequestQuery(sanitized.RequestTarget)
		return fmt.Errorf("oss: %w", &sanitized)
	}
	return err
}

func stripOSSRequestQuery(target string) string {
	method := ""
	rawURL := target
	if prefix, remainder, found := strings.Cut(target, " "); found {
		method = prefix + " "
		rawURL = remainder
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		if before, _, found := strings.Cut(rawURL, "?"); found {
			return method + before
		}
		return target
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	return method + parsed.String()
}
