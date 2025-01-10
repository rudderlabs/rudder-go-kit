package provider

import (
	"errors"
	"time"
)

var ErrProviderInit = errors.New("failed to initialize provider")

// ProviderParams represents parameters for feature flag provider
type ProviderParams struct {
	WorkspaceID string
	Traits      map[string]string
}

// FeatureValue represents a feature flag value
type FeatureValue struct {
	Name          string
	Enabled       bool
	Value         interface{}
	LastUpdatedAt *time.Time
	IsStale       bool
	Error         error
}

// Provider defines the interface for feature flag providers
type Provider interface {
	// GetFeatureFlags retrieves all feature flags for given parameters
	GetFeatureFlags(params ProviderParams) (map[string]*FeatureValue, error)

	// Name returns the provider name
	Name() string
}

type ProviderConfig struct {
	Type                   string
	ApiKey                 string
	TimeoutInSeconds       int
	RetryAttempts          int
	RetryWaitTimeInSeconds int
}

func NewProvider(config ProviderConfig) (Provider, error) {
	switch config.Type {
	case "flagsmith":
		return NewFlagsmithProvider(config)
	default:
		return nil, ErrProviderInit
	}
}

