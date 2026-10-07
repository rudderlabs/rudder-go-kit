package filemanager

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/datalakeerror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/file"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/filesystem"
	"github.com/google/uuid"

	obskit "github.com/rudderlabs/rudder-observability-kit/go/labels"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

const oneLakeEndpoint = "https://onelake.dfs.fabric.microsoft.com"

type oneLakeConfig struct {
	workspaceID  string
	lakehouseID  string
	tenantID     string
	clientID     string
	clientSecret string
}

type oneLakeCredentialKey struct {
	tenantID   string
	clientID   string
	secretHash [sha256.Size]byte
}

type oneLakeManagerOptions struct {
	serviceEndpoint string
	credential      azcore.TokenCredential
	clientOptions   azcore.ClientOptions
}

var oneLakeCredentialCache sync.Map

// OneLakeManager manages files in a Microsoft Fabric Lakehouse's Files area.
type OneLakeManager struct {
	*baseManager
	config     oneLakeConfig
	credential azcore.TokenCredential
	filesystem *filesystem.Client
}

var _ FileManager = (*OneLakeManager)(nil)

// NewOneLakeManager creates a file manager for Microsoft Fabric OneLake.
func NewOneLakeManager(config map[string]any, log logger.Logger, defaultTimeout func() time.Duration) (*OneLakeManager, error) {
	return newOneLakeManager(config, log, defaultTimeout, oneLakeManagerOptions{})
}

func newOneLakeManager(config map[string]any, log logger.Logger, defaultTimeout func() time.Duration, options oneLakeManagerOptions) (*OneLakeManager, error) {
	oneLakeConfig, err := parseOneLakeConfig(config)
	if err != nil {
		return nil, err
	}

	credential := options.credential
	if credential == nil {
		credential, err = getOneLakeCredential(oneLakeConfig)
		if err != nil {
			return nil, fmt.Errorf("onelake: creating credential: %w", err)
		}
	}

	endpoint := options.serviceEndpoint
	if endpoint == "" {
		endpoint = oneLakeEndpoint
	}
	filesystemURL := strings.TrimRight(endpoint, "/") + "/" + oneLakeConfig.workspaceID
	client, err := filesystem.NewClient(filesystemURL, credential, &filesystem.ClientOptions{
		ClientOptions: options.clientOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("onelake: creating filesystem client: %w", err)
	}

	return &OneLakeManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config:     oneLakeConfig,
		credential: credential,
		filesystem: client,
	}, nil
}

func parseOneLakeConfig(config map[string]any) (oneLakeConfig, error) {
	workspaceID, err := requiredOneLakeString(config, "fabricWorkspaceId")
	if err != nil {
		return oneLakeConfig{}, err
	}
	if parsed, parseErr := uuid.Parse(workspaceID); parseErr != nil || parsed.String() != workspaceID {
		return oneLakeConfig{}, errors.New("onelake: invalid fabricWorkspaceId")
	}

	lakehouseID, err := requiredOneLakeString(config, "lakehouseId")
	if err != nil {
		return oneLakeConfig{}, err
	}
	if parsed, parseErr := uuid.Parse(lakehouseID); parseErr != nil || parsed.String() != lakehouseID {
		return oneLakeConfig{}, errors.New("onelake: invalid lakehouseId")
	}

	tenantID, err := requiredOneLakeString(config, "tenantId")
	if err != nil {
		return oneLakeConfig{}, err
	}
	clientID, err := requiredOneLakeString(config, "clientId")
	if err != nil {
		return oneLakeConfig{}, err
	}
	clientSecret, err := requiredOneLakeString(config, "clientSecret")
	if err != nil {
		return oneLakeConfig{}, err
	}

	return oneLakeConfig{
		workspaceID:  workspaceID,
		lakehouseID:  lakehouseID,
		tenantID:     tenantID,
		clientID:     clientID,
		clientSecret: clientSecret,
	}, nil
}

func requiredOneLakeString(config map[string]any, key string) (string, error) {
	value, ok := config[key].(string)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("onelake: missing or blank %s", key)
	}
	return value, nil
}

func getOneLakeCredential(config oneLakeConfig) (azcore.TokenCredential, error) {
	key := oneLakeCredentialKey{
		tenantID:   config.tenantID,
		clientID:   config.clientID,
		secretHash: sha256.Sum256([]byte(config.clientSecret)),
	}
	if cached, ok := oneLakeCredentialCache.Load(key); ok {
		return cached.(azcore.TokenCredential), nil
	}

	credential, err := azidentity.NewClientSecretCredential(config.tenantID, config.clientID, config.clientSecret, nil)
	if err != nil {
		return nil, err
	}
	actual, _ := oneLakeCredentialCache.LoadOrStore(key, credential)
	return actual.(azcore.TokenCredential), nil
}

func resetOneLakeCredentialCache() {
	oneLakeCredentialCache = sync.Map{}
}

// Upload uploads a file to OneLake.
func (m *OneLakeManager) Upload(ctx context.Context, input *os.File, prefixes ...string) (UploadedFile, error) {
	segments := append(append([]string{}, prefixes...), path.Base(input.Name()))
	objectName := path.Join(segments...)
	if err := validateOneLakeObjectName(objectName); err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	client := m.fileClient(objectName)
	if err := m.prepareUpload(ctx, objectName, client); err != nil {
		return UploadedFile{}, err
	}
	if err := client.UploadFile(ctx, input, nil); err != nil {
		return UploadedFile{}, fmt.Errorf("onelake: uploading %s: %w", objectName, err)
	}
	return m.uploadedFile(objectName), nil
}

// UploadReader streams data from a reader into OneLake.
func (m *OneLakeManager) UploadReader(ctx context.Context, objectName string, reader io.Reader) (UploadedFile, error) {
	if err := validateOneLakeObjectName(objectName); err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	client := m.fileClient(objectName)
	if err := m.prepareUpload(ctx, objectName, client); err != nil {
		return UploadedFile{}, err
	}
	if err := client.UploadStream(ctx, reader, nil); err != nil {
		return UploadedFile{}, fmt.Errorf("onelake: uploading %s: %w", objectName, err)
	}
	return m.uploadedFile(objectName), nil
}

func (m *OneLakeManager) prepareUpload(ctx context.Context, objectName string, client *file.Client) error {
	parent := path.Dir(objectName)
	if parent != "." {
		current := m.filesRoot()
		for segment := range strings.SplitSeq(parent, "/") {
			current = path.Join(current, segment)
			_, err := m.filesystem.NewDirectoryClient(current).Create(ctx, nil)
			if err != nil && !datalakeerror.HasCode(err, datalakeerror.PathAlreadyExists) {
				return fmt.Errorf("onelake: creating parent for %s: %w", objectName, err)
			}
		}
	}
	if _, err := client.Create(ctx, nil); err != nil {
		if !datalakeerror.HasCode(err, datalakeerror.PathAlreadyExists) {
			return fmt.Errorf("onelake: creating %s: %w", objectName, err)
		}
		if _, deleteErr := client.Delete(ctx, nil); deleteErr != nil && !datalakeerror.HasCode(deleteErr, datalakeerror.PathNotFound, datalakeerror.BlobNotFound) {
			return fmt.Errorf("onelake: replacing %s: %w", objectName, deleteErr)
		}
		if _, createErr := client.Create(ctx, nil); createErr != nil {
			return fmt.Errorf("onelake: creating %s: %w", objectName, createErr)
		}
	}
	return nil
}

// Download retrieves a OneLake file and writes it to output.
func (m *OneLakeManager) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	if err := validateOneLakeObjectName(key); err != nil {
		return err
	}

	downloadOptions := applyDownloadOptions(opts...)
	sdkOptions := &file.DownloadStreamOptions{Range: &file.HTTPRange{}}
	if downloadOptions.isRangeRequest {
		sdkOptions.Range.Offset = downloadOptions.offset
		sdkOptions.Range.Count = downloadOptions.length
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	response, err := m.fileClient(key).DownloadStream(ctx, sdkOptions)
	if err != nil {
		if datalakeerror.HasCode(err, datalakeerror.PathNotFound, datalakeerror.BlobNotFound) {
			return ErrKeyNotFound
		}
		return fmt.Errorf("onelake: downloading %s: %w", key, err)
	}

	body := response.NewRetryReader(ctx, &file.RetryReaderOptions{})
	_, copyErr := io.Copy(io.NewOffsetWriter(output, 0), body)
	closeErr := body.Close()
	if copyErr != nil {
		return fmt.Errorf("onelake: downloading %s: %w", key, copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("onelake: closing download %s: %w", key, closeErr)
	}
	return nil
}

// Delete removes files from OneLake. Missing files are ignored.
func (m *OneLakeManager) Delete(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := validateOneLakeObjectName(key); err != nil {
			return err
		}

		deleteCtx, cancel := context.WithTimeout(ctx, m.getTimeout())
		_, err := m.fileClient(key).Delete(deleteCtx, nil)
		cancel()
		if err == nil || datalakeerror.HasCode(err, datalakeerror.PathNotFound, datalakeerror.BlobNotFound) {
			continue
		}
		m.logger.Errorn("OneLake object delete failed", logger.NewStringField("objectName", key), obskit.Error(err))
		return fmt.Errorf("onelake: deleting %s: %w", key, err)
	}
	return nil
}

// ListFilesWithPrefix starts a recursive OneLake listing session.
func (m *OneLakeManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	directory, err := m.listDirectory(prefix)
	if err == nil && startAfter != "" {
		err = validateOneLakeObjectName(startAfter)
	}
	if maxItems < 0 {
		err = errors.New("onelake: maxItems cannot be negative")
	}

	session := &oneLakeListSession{
		baseListSession: &baseListSession{
			ctx:        ctx,
			startAfter: startAfter,
			prefix:     prefix,
			maxItems:   maxItems,
		},
		manager:   m,
		directory: directory,
		err:       err,
	}
	if err == nil && maxItems > 0 {
		pageSize := int32(min(maxItems, 5000))
		session.pager = m.filesystem.NewListPathsPager(true, &filesystem.ListPathsOptions{
			Prefix:     &directory,
			MaxResults: &pageSize,
		})
	}
	return session
}

// Prefix returns the OneLake object prefix. OneLake files have no configurable prefix.
func (m *OneLakeManager) Prefix() string {
	return ""
}

// GetObjectNameFromLocation extracts a relative object name from a OneLake DFS URL.
func (m *OneLakeManager) GetObjectNameFromLocation(location string) (string, error) {
	objectName, err := m.objectNameFromLocation(location)
	if err != nil {
		m.logger.Errorn("OneLake location rejected", obskit.Error(err))
		return "", err
	}
	return objectName, nil
}

func (m *OneLakeManager) objectNameFromLocation(location string) (string, error) {
	if !strings.Contains(location, "://") {
		if err := validateOneLakeObjectName(location); err != nil {
			return "", err
		}
		return location, nil
	}

	parsed, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("onelake: parsing location: %w", err)
	}
	if parsed.Scheme != "https" {
		return "", errors.New("onelake: location must use https")
	}
	if parsed.User != nil || parsed.Host != "onelake.dfs.fabric.microsoft.com" {
		return "", errors.New("onelake: location has an invalid host")
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", errors.New("onelake: location contains unsupported URL components")
	}
	for segment := range strings.SplitSeq(strings.TrimPrefix(parsed.EscapedPath(), "/"), "/") {
		decoded, decodeErr := url.PathUnescape(segment)
		if decodeErr != nil || strings.Contains(decoded, "/") {
			return "", errors.New("onelake: location contains an invalid escaped path segment")
		}
	}

	root := "/" + m.config.workspaceID + "/" + m.config.lakehouseID + "/Files/"
	if !strings.HasPrefix(parsed.Path, root) {
		return "", errors.New("onelake: location is outside the configured Lakehouse Files area")
	}
	objectName := strings.TrimPrefix(parsed.Path, root)
	if err := validateOneLakeObjectName(objectName); err != nil {
		return "", err
	}
	return objectName, nil
}

// GetDownloadKeyFromFileLocation returns the relative OneLake object name, or an empty string for an invalid location.
func (m *OneLakeManager) GetDownloadKeyFromFileLocation(location string) string {
	objectName, err := m.GetObjectNameFromLocation(location)
	if err != nil {
		return ""
	}
	return objectName
}

func (m *OneLakeManager) fileClient(objectName string) *file.Client {
	return m.filesystem.NewFileClient(path.Join(m.filesRoot(), objectName))
}

func (m *OneLakeManager) filesRoot() string {
	return path.Join(m.config.lakehouseID, "Files")
}

func (m *OneLakeManager) uploadedFile(objectName string) UploadedFile {
	location := oneLakeEndpoint + "/" + m.config.workspaceID + "/" + m.config.lakehouseID + "/Files/" + escapeOneLakeObjectName(objectName)
	return UploadedFile{Location: location, ObjectName: objectName}
}

func escapeOneLakeObjectName(objectName string) string {
	segments := strings.Split(objectName, "/")
	for i, segment := range segments {
		segments[i] = url.PathEscape(segment)
	}
	return strings.Join(segments, "/")
}

func (m *OneLakeManager) listDirectory(prefix string) (string, error) {
	if err := validateOneLakePrefix(prefix); err != nil {
		return "", err
	}
	directory := prefix
	if !strings.HasSuffix(prefix, "/") {
		directory = path.Dir(prefix)
	}
	if directory == "." || directory == "" {
		return m.filesRoot(), nil
	}
	return path.Join(m.filesRoot(), strings.TrimSuffix(directory, "/")), nil
}

func validateOneLakePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	prefix = strings.TrimSuffix(prefix, "/")
	return validateOneLakeObjectName(prefix)
}

func validateOneLakeObjectName(objectName string) error {
	if objectName == "" {
		return errors.New("onelake: object name cannot be empty")
	}
	if strings.HasPrefix(objectName, "/") {
		return errors.New("onelake: object name cannot start with a slash")
	}
	if strings.Contains(objectName, `\`) {
		return errors.New("onelake: object name cannot contain a backslash")
	}
	for segment := range strings.SplitSeq(objectName, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return errors.New("onelake: object name contains an invalid path segment")
		}
	}
	return nil
}

type oneLakeListSession struct {
	*baseListSession
	manager   *OneLakeManager
	directory string
	pager     *runtime.Pager[filesystem.ListPathsSegmentResponse]
	pending   []*FileInfo
	exhausted bool
	err       error
}

func (s *oneLakeListSession) Next() ([]*FileInfo, error) {
	if s.err != nil {
		err := s.err
		s.err = nil
		s.exhausted = true
		return nil, err
	}
	if s.exhausted || s.maxItems == 0 {
		return nil, nil
	}

	ctx, cancel := context.WithTimeout(s.ctx, s.manager.getTimeout())
	defer cancel()

	results := make([]*FileInfo, 0, s.maxItems)
	for len(results) < int(s.maxItems) {
		if len(s.pending) > 0 {
			remaining := min(int(s.maxItems)-len(results), len(s.pending))
			results = append(results, s.pending[:remaining]...)
			s.pending = s.pending[remaining:]
			if len(results) == int(s.maxItems) {
				return results, nil
			}
		}

		if !s.pager.More() {
			s.exhausted = true
			break
		}
		page, err := s.pager.NextPage(ctx)
		if err != nil {
			s.pending = append(results, s.pending...)
			if datalakeerror.HasCode(err, datalakeerror.PathNotFound) {
				s.exhausted = true
				return nil, nil
			}
			s.manager.logger.Errorn("OneLake listing page failed", logger.NewStringField("directory", s.directory), obskit.Error(err))
			return nil, fmt.Errorf("onelake: listing %s: %w", s.directory, err)
		}

		for _, item := range page.Paths {
			if item == nil || item.Name == nil || (item.IsDirectory != nil && *item.IsDirectory) {
				continue
			}
			objectName := strings.TrimPrefix(*item.Name, s.manager.filesRoot()+"/")
			if objectName == *item.Name || !strings.HasPrefix(objectName, s.prefix) || strings.Compare(objectName, s.startAfter) <= 0 {
				continue
			}
			lastModified, err := oneLakeLastModified(item.LastModified)
			if err != nil {
				return nil, fmt.Errorf("onelake: listing %s: %w", s.directory, err)
			}
			s.pending = append(s.pending, &FileInfo{Key: objectName, LastModified: lastModified})
		}
	}
	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}

func oneLakeLastModified(value *string) (time.Time, error) {
	if value == nil || *value == "" {
		return time.Time{}, nil
	}
	lastModified, err := http.ParseTime(*value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parsing last modified: %w", err)
	}
	return lastModified, nil
}
