package provider

import (
	"context"
	"time"

	flagsmith "github.com/Flagsmith/flagsmith-go-client/v3"
)

// FlagsmithProvider implements the Provider interface using Flagsmith
type FlagsmithProvider struct {
	Client FlagsmithClient
}

type FlagsmithFlags interface {
	AllFlags() []flagsmith.Flag
}

type FlagsmithClient interface {
	GetIdentityFlags(ctx context.Context, identifier string, traits []*flagsmith.Trait) (FlagsmithFlags, error)
}

// FlagsmithClientAdapter wraps the flagsmith client to implement FlagsmithClient interface
type FlagsmithClientAdapter struct {
	client *flagsmith.Client
}

// GetIdentityFlags implements FlagsmithClient interface
func (a *FlagsmithClientAdapter) GetIdentityFlags(ctx context.Context, identifier string, traits []*flagsmith.Trait) (FlagsmithFlags, error) {
	flags, err := a.client.GetIdentityFlags(ctx, identifier, traits)
	if err != nil {
		return nil, err
	}
	return &flags, nil
}

// NewFlagsmithProvider creates a new Flagsmith provider instance
func NewFlagsmithProvider(config ProviderConfig) (*FlagsmithProvider, error) {

	client := flagsmith.NewClient(config.ApiKey,
		flagsmith.WithRequestTimeout(time.Duration(config.TimeoutInSeconds)*time.Second),
		flagsmith.WithRetries(config.RetryAttempts, time.Duration(config.RetryWaitTimeInSeconds)*time.Second),
	)
	if client == nil {
		return nil, ErrProviderInit
	}

	return &FlagsmithProvider{
		Client: &FlagsmithClientAdapter{client: client},
	}, nil
}

// GetFeatureFlags implements Provider.GetFeatureFlags
func (p *FlagsmithProvider) GetFeatureFlags(params ProviderParams) (map[string]*FeatureValue, error) {
	// Create traits map
	traits := make([]*flagsmith.Trait, 0)
	for k, v := range params.Traits {
		traits = append(traits, &flagsmith.Trait{
			TraitKey:   k,
			TraitValue: v,
		})
	}

	// Get flags for identity with traits
	flags, err := p.Client.GetIdentityFlags(context.Background(), params.WorkspaceID, traits)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := make(map[string]*FeatureValue)

	// Convert flags to our format
	for _, flag := range flags.AllFlags() {
		featureValue := FeatureValue{
			Name:          flag.FeatureName,
			Enabled:       flag.Enabled,
			Value:         flag.Value,
			LastUpdatedAt: &now,
			IsStale:       false,
		}

		result[flag.FeatureName] = &featureValue
	}

	return result, nil
}

// Name implements Provider.Name
func (p *FlagsmithProvider) Name() string {
	return "flagsmith"
}
