package filemanager

import (
	"os"

	"github.com/rudderlabs/rudder-go-kit/config"
)

type ProviderConfigOpts struct {
	Provider           string
	Bucket             string
	Prefix             string
	Config             *config.Config
	ExternalIDSupplier func() string
}

func GetProviderConfigFromEnv(opts ProviderConfigOpts) map[string]any {
	if opts.Config == nil {
		opts.Config = config.Default
	}
	if opts.ExternalIDSupplier == nil {
		opts.ExternalIDSupplier = func() string { return "" }
	}
	config := opts.Config
	providerConfig := make(map[string]any)
	switch opts.Provider {
	case "S3":
		providerConfig["bucketName"] = opts.Bucket
		providerConfig["prefix"] = opts.Prefix
		providerConfig["accessKeyID"] = config.GetStringVar("", "AWS_ACCESS_KEY_ID")
		providerConfig["accessKey"] = config.GetStringVar("", "AWS_SECRET_ACCESS_KEY")
		providerConfig["enableSSE"] = config.GetBoolVar(false, "AWS_ENABLE_SSE")
		providerConfig["regionHint"] = config.GetStringVar("us-east-1", "AWS_S3_REGION_HINT")
		providerConfig["iamRoleArn"] = config.GetStringVar("", "BACKUP_IAM_ROLE_ARN")
		providerConfig["region"] = config.GetStringVar("", "AWS_REGION")
		if providerConfig["iamRoleArn"] != "" {
			providerConfig["externalID"] = opts.ExternalIDSupplier()
		}
	case "GCS":
		providerConfig["bucketName"] = opts.Bucket
		providerConfig["prefix"] = opts.Prefix
		credentials, err := os.ReadFile(config.GetStringVar("", "GOOGLE_APPLICATION_CREDENTIALS"))
		if err == nil {
			providerConfig["credentials"] = string(credentials)
		}
	case "AZURE_BLOB":
		providerConfig["containerName"] = opts.Bucket
		providerConfig["prefix"] = opts.Prefix
		providerConfig["accountName"] = config.GetStringVar("", "AZURE_STORAGE_ACCOUNT")
		providerConfig["accountKey"] = config.GetStringVar("", "AZURE_STORAGE_ACCESS_KEY")
	case "MINIO":
		providerConfig["bucketName"] = opts.Bucket
		providerConfig["prefix"] = opts.Prefix
		providerConfig["endPoint"] = config.GetStringVar("localhost:9000", "MINIO_ENDPOINT")
		providerConfig["accessKeyID"] = config.GetStringVar("minioadmin", "MINIO_ACCESS_KEY_ID")
		providerConfig["secretAccessKey"] = config.GetStringVar("minioadmin", "MINIO_SECRET_ACCESS_KEY")
		providerConfig["useSSL"] = config.GetBoolVar(false, "MINIO_SSL")
	case "DIGITAL_OCEAN_SPACES":
		providerConfig["bucketName"] = opts.Bucket
		providerConfig["prefix"] = opts.Prefix
		providerConfig["endPoint"] = config.GetStringVar("", "DO_SPACES_ENDPOINT")
		providerConfig["accessKeyID"] = config.GetStringVar("", "DO_SPACES_ACCESS_KEY_ID")
		providerConfig["accessKey"] = config.GetStringVar("", "DO_SPACES_SECRET_ACCESS_KEY")
	}

	return providerConfig
}
