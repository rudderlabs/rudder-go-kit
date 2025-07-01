package filemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/rudderlabs/rudder-go-kit/logger"
)

type AzureBlobConfig struct {
	Container      string
	Prefix         string
	AccountName    string
	AccountKey     string
	SASToken       string
	EndPoint       *string
	ForcePathStyle *bool
	DisableSSL     *bool
	UseSASTokens   bool
}

// NewAzureBlobManager creates a new file manager for Azure Blob Storage
func NewAzureBlobManager(config map[string]interface{}, log logger.Logger, defaultTimeout func() time.Duration) (*AzureBlobManager, error) {
	return &AzureBlobManager{
		baseManager: &baseManager{
			logger:         log,
			defaultTimeout: defaultTimeout,
		},
		config: azureBlobConfig(config),
	}, nil
}

func (m *AzureBlobManager) ListFilesWithPrefix(ctx context.Context, startAfter, prefix string, maxItems int64) ListSession {
	return &azureBlobListSession{
		baseListSession: &baseListSession{
			ctx:        ctx,
			startAfter: startAfter,
			prefix:     prefix,
			maxItems:   maxItems,
		},
		manager: m,
	}
}

// Upload passed in file to Azure Blob Storage
func (m *AzureBlobManager) Upload(ctx context.Context, file *os.File, prefixes ...string) (UploadedFile, error) {
	objName := path.Join(m.config.Prefix, path.Join(prefixes...), path.Base(file.Name()))
	return m.UploadReader(ctx, objName, file)
}

func (m *AzureBlobManager) UploadReader(ctx context.Context, objName string, rdr io.Reader) (UploadedFile, error) {
	containerURL, err := m.getContainerURL()
	if err != nil {
		return UploadedFile{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	if m.createContainer() {
		_, err = containerURL.Create(ctx, azblob.Metadata{}, azblob.PublicAccessNone)
		err = m.suppressMinorErrors(err)
		if err != nil {
			return UploadedFile{}, err
		}
	}

	// Here's how to upload a blob.
	blobURL := containerURL.NewBlockBlobURL(objName)
	if file, ok := rdr.(*os.File); ok {
		_, err = azblob.UploadFileToBlockBlob(ctx, file, blobURL, azblob.UploadToBlockBlobOptions{
			BlockSize:   4 * 1024 * 1024,
			Parallelism: 16,
		})
	} else {
		_, err = azblob.UploadStreamToBlockBlob(ctx, rdr, blobURL, azblob.UploadStreamToBlockBlobOptions{})
	}
	if err != nil {
		return UploadedFile{}, err
	}

	return UploadedFile{Location: m.blobLocation(&blobURL), ObjectName: objName}, nil
}

// Download retrieves an object with the given key and writes it to the provided writer.
// Pass *os.File as output to write the downloaded file on disk.
func (m *AzureBlobManager) Download(ctx context.Context, output io.WriterAt, key string, opts ...DownloadOption) error {
	downloadOpts := applyDownloadOptions(opts...)
	containerURL, err := m.getContainerURL()
	if err != nil {
		return err
	}

	blobURL := containerURL.NewBlockBlobURL(key)

	ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
	defer cancel()

	offset := int64(0)
	count := int64(azblob.CountToEnd)

	if downloadOpts.isRangeRequest {
		offset = downloadOpts.offset
		if downloadOpts.length > 0 {
			count = downloadOpts.length
		}
	}

	// Here's how to download the blob
	downloadResponse, err := blobURL.Download(ctx, offset, count, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return err
	}

	// NOTE: automatically retries are performed if the connection fails
	bodyStream := downloadResponse.Body(azblob.RetryReaderOptions{MaxRetryRequests: 20})
	_, err = io.Copy(&writerAtAdapter{w: output}, bodyStream)
	_ = bodyStream.Close()
	return err
}

func (m *AzureBlobManager) Delete(ctx context.Context, keys []string) (err error) {
	containerURL, err := m.getContainerURL()
	if err != nil {
		return err
	}

	for _, key := range keys {
		blobURL := containerURL.NewBlockBlobURL(key)

		_ctx, cancel := context.WithTimeout(ctx, m.getTimeout())
		_, err := blobURL.Delete(_ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
		if err != nil {
			cancel()
			return err
		}
		cancel()
	}
	return
}

func (m *AzureBlobManager) Prefix() string {
	return m.config.Prefix
}

func (m *AzureBlobManager) GetObjectNameFromLocation(location string) (string, error) {
	strToken := strings.Split(location, fmt.Sprintf("%s/", m.config.Container))
	return strToken[len(strToken)-1], nil
}

func (m *AzureBlobManager) GetDownloadKeyFromFileLocation(location string) string {
	str := strings.Split(location, fmt.Sprintf("%s/", m.config.Container))
	return str[len(str)-1]
}

func (m *AzureBlobManager) suppressMinorErrors(err error) error {
	if err != nil {
		var storageError azblob.StorageError
		if errors.As(err, &storageError) { // This error is a Service-specific
			switch storageError.ServiceCode() { // Compare serviceCode to ServiceCodeXxx constants
			case azblob.ServiceCodeContainerAlreadyExists:
				m.logger.Debugn("Received 409. Container already exists")
				return nil
			}
		}
	}
	return err
}

func (m *AzureBlobManager) getBaseURL() *url.URL {
	protocol := "https"
	if m.config.DisableSSL != nil && *m.config.DisableSSL {
		protocol = "http"
	}

	endpoint := "blob.core.windows.net"
	if m.config.EndPoint != nil && *m.config.EndPoint != "" {
		endpoint = *m.config.EndPoint
	}

	baseURL := url.URL{
		Scheme: protocol,
		Host:   fmt.Sprintf("%s.%s", m.config.AccountName, endpoint),
	}

	if m.config.UseSASTokens {
		baseURL.RawQuery = m.config.SASToken
	}

	if m.config.ForcePathStyle != nil && *m.config.ForcePathStyle {
		baseURL.Host = endpoint
		baseURL.Path = fmt.Sprintf("/%s/", m.config.AccountName)
	}

	return &baseURL
}

func (m *AzureBlobManager) getContainerURL() (azblob.ContainerURL, error) {
	if m.config.Container == "" {
		return azblob.ContainerURL{}, errors.New("no container configured")
	}

	credential, err := m.getCredentials()
	if err != nil {
		return azblob.ContainerURL{}, err
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint.
	baseURL := m.getBaseURL()
	serviceURL := azblob.NewServiceURL(*baseURL, p)
	containerURL := serviceURL.NewContainerURL(m.config.Container)

	return containerURL, nil
}

func (m *AzureBlobManager) getCredentials() (azblob.Credential, error) {
	if m.config.UseSASTokens {
		return azblob.NewAnonymousCredential(), nil
	}

	accountName, accountKey := m.config.AccountName, m.config.AccountKey
	if accountName == "" || accountKey == "" {
		return nil, errors.New("either accountName or accountKey is empty")
	}

	// Create a default request pipeline using your storage account name and account key.
	return azblob.NewSharedKeyCredential(accountName, accountKey)
}

func (m *AzureBlobManager) createContainer() bool {
	return !m.config.UseSASTokens
}

func (m *AzureBlobManager) blobLocation(blobURL *azblob.BlockBlobURL) string {
	if !m.config.UseSASTokens {
		return blobURL.String()
	}

	// Reset SAS Query parameters
	blobURLParts := azblob.NewBlobURLParts(blobURL.URL())
	blobURLParts.SAS = azblob.SASQueryParameters{}
	newBlobURL := blobURLParts.URL()
	return newBlobURL.String()
}

type AzureBlobManager struct {
	*baseManager
	config *AzureBlobConfig
}

func azureBlobConfig(config map[string]interface{}) *AzureBlobConfig {
	var containerName, accountName, accountKey, sasToken, prefix string
	var endPoint *string
	var forcePathStyle, disableSSL *bool
	var useSASTokens bool
	if config["containerName"] != nil {
		tmp, ok := config["containerName"].(string)
		if ok {
			containerName = tmp
		}
	}
	if config["prefix"] != nil {
		tmp, ok := config["prefix"].(string)
		if ok {
			prefix = tmp
		}
	}
	if config["accountName"] != nil {
		tmp, ok := config["accountName"].(string)
		if ok {
			accountName = tmp
		}
	}
	if config["useSASTokens"] != nil {
		tmp, ok := config["useSASTokens"].(bool)
		if ok {
			useSASTokens = tmp
		}
	}
	if config["sasToken"] != nil {
		tmp, ok := config["sasToken"].(string)
		if ok {
			sasToken = strings.TrimPrefix(tmp, "?")
		}
	}
	if config["accountKey"] != nil {
		tmp, ok := config["accountKey"].(string)
		if ok {
			accountKey = tmp
		}
	}
	if config["endPoint"] != nil {
		tmp, ok := config["endPoint"].(string)
		if ok {
			endPoint = &tmp
		}
	}
	if config["forcePathStyle"] != nil {
		tmp, ok := config["forcePathStyle"].(bool)
		if ok {
			forcePathStyle = &tmp
		}
	}
	if config["disableSSL"] != nil {
		tmp, ok := config["disableSSL"].(bool)
		if ok {
			disableSSL = &tmp
		}
	}
	return &AzureBlobConfig{
		Container:      containerName,
		Prefix:         prefix,
		AccountName:    accountName,
		AccountKey:     accountKey,
		UseSASTokens:   useSASTokens,
		SASToken:       sasToken,
		EndPoint:       endPoint,
		ForcePathStyle: forcePathStyle,
		DisableSSL:     disableSSL,
	}
}

type azureBlobListSession struct {
	*baseListSession
	manager *AzureBlobManager

	Marker azblob.Marker
}

func (l *azureBlobListSession) Next() (fileObjects []*FileInfo, err error) {
	manager := l.manager
	maxItems := l.maxItems

	containerURL, err := manager.getContainerURL()
	if err != nil {
		return []*FileInfo{}, err
	}

	blobListingDetails := azblob.BlobListingDetails{
		Metadata: true,
	}
	segmentOptions := azblob.ListBlobsSegmentOptions{
		Details:    blobListingDetails,
		Prefix:     l.prefix,
		MaxResults: int32(l.maxItems),
	}

	ctx, cancel := context.WithTimeout(l.ctx, manager.getTimeout())
	defer cancel()

	// List the blobs in the container
	var response *azblob.ListBlobsFlatSegmentResponse

	// Checking if maxItems > 0 to avoid function calls which expect only maxItems to be returned and not more in the code
	for maxItems > 0 && l.Marker.NotDone() {
		response, err = containerURL.ListBlobsFlatSegment(ctx, l.Marker, segmentOptions)
		if err != nil {
			return
		}
		l.Marker = response.NextMarker

		fileObjects = make([]*FileInfo, 0)
		for idx := range response.Segment.BlobItems {
			if strings.Compare(response.Segment.BlobItems[idx].Name, l.startAfter) > 0 {
				fileObjects = append(fileObjects, &FileInfo{response.Segment.BlobItems[idx].Name, response.Segment.BlobItems[idx].Properties.LastModified})
				maxItems--
			}
		}
	}
	return
}
