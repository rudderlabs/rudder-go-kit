package rudderauth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/postgres"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rand"
)

// AccountFixture is the fixture used to create an account in the config backend
type AccountFixture struct {
	ID                   string          // the account id
	Type                 string          // the account type (e.g. reddit)
	Category             string          // the account category (default: destination)
	WorkspaceID          string          // the workspace id the account belongs to
	DestinationID        string          // the destination id associated with this account if any
	Options              json.RawMessage // the account options, if any
	HasAccountDefinition bool            // whether the account has a v1 account definition

	Secret            json.RawMessage   // the account secret, if any
	AuthClientSecrets map[string]string // the client secrets, e.g. CLIENT_ID, CLIENT_SECRET, etc.
}

func (af *AccountFixture) applyDefaults() {
	if af.Options == nil {
		af.Options = json.RawMessage(`{}`)
	}
}

func (af *AccountFixture) accountDefName() string {
	return strings.ToUpper(fmt.Sprintf("%s_%s", af.Category, af.Type))
}

// Resource represents a Rudder Auth resource
type Resource struct {
	// AdminUsername is the admin username to access the config backend
	AdminUsername string
	// AdminPassword is the admin password to access the config backend
	AdminPassword string
	// HostedSecret is the hosted service secret to access the config backend
	HostedSecret string
	// Fixture is the account fixture used to create an account in the config backend
	Fixture AccountFixture
	// Postgres is the Postgres resource used by the config backend and secrets service
	Postgres *postgres.Resource
	// SecretsURL is the URL of the secrets service
	SecretsURL string
	// AuthURL is the URL of the auth service
	AuthURL string
	// ConfigBackendURL is the URL of the config backend service
	ConfigBackendURL string
}

// FetchToken fetches a token for the account from the config backend
func (r *Resource) FetchToken() ([]byte, error) {
	url := fmt.Sprintf("%s/destination/workspaces/%s/accounts/%s/token", r.ConfigBackendURL, r.Fixture.WorkspaceID, r.Fixture.ID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json.RawMessage(`{}`)))
	if err != nil {
		return nil, fmt.Errorf("creating fetch token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.HostedSecret, "")
	res, err := http.DefaultClient.Do(req)
	defer func() { httputil.CloseResponse(res) }()
	if err != nil {
		return nil, fmt.Errorf("fetching token: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("fetching token not ok: %s: %s", res.Status, string(body))
	}
	token, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token response body: %w", err)
	}
	return token, nil
}

// ToggleStatus toggles the auth status of the destination in the config backend
func (r *Resource) ToggleStatus(status string) error {
	if r.Fixture.DestinationID == "" {
		return errors.New("cannot toggle status: Fixture.DestinationID is empty")
	}
	url := fmt.Sprintf("%s/workspaces/%s/destinations/%s/authStatus/toggle", r.ConfigBackendURL, r.Fixture.WorkspaceID, r.Fixture.DestinationID)
	body, err := jsonrs.Marshal(map[string]string{
		"authStatus": status,
	})
	if err != nil {
		return fmt.Errorf("marshaling toggle status request body: %w", err)
	}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating toggle status request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(r.HostedSecret, "")
	res, err := http.DefaultClient.Do(req)
	defer func() { httputil.CloseResponse(res) }()
	if err != nil {
		return fmt.Errorf("toggling status: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("toggling status not ok: %s: %s", res.Status, string(body))
	}
	return nil
}

type Option func(*config)

func WithSecretsTag(tag string) Option {
	return func(c *config) {
		c.secretsTag = tag
	}
}

func WithAuthTag(tag string) Option {
	return func(c *config) {
		c.authTag = tag
	}
}

func WithConfigBackendTag(tag string) Option {
	return func(c *config) {
		c.configBackendTag = tag
	}
}

func WithBindIP(ip string) Option {
	return func(c *config) {
		c.bindIP = ip
	}
}

type config struct {
	secretsRepository string
	secretsTag        string

	authRepository string
	authTag        string

	configBackendRepository string
	configBackendTag        string

	bindIP string
}

// Setup sets up a Rudder Auth resource by doing the following steps:
//
//   - it creates a docker network
//   - it starts a postgres container and seeds it with the necessary data for the provided account fixture
//   - it starts a secrets service container
//   - it starts a rudder-auth container
//   - it starts a config-backend container
func Setup(pool *dockertest.Pool, fixture AccountFixture, d resource.Cleaner, opts ...Option) (*Resource, error) {
	fixture.applyDefaults()
	suffix := rand.UniqueString(5)
	var (
		adminUsername       = strings.ToLower(rand.String(10))
		adminPassword       = strings.ToLower(rand.String(10))
		hostedServiceSecret = strings.ToLower(rand.String(10))
	)
	resource := Resource{
		AdminUsername: adminUsername,
		AdminPassword: adminPassword,
		HostedSecret:  hostedServiceSecret,
		Fixture:       fixture,
	}
	// Set Rudder Transformer
	// pulls an image first to make sure we don't have an old cached version locally,
	// then it creates a container based on it and runs it
	conf := &config{
		secretsRepository: "rudderstack/secrets",
		secretsTag:        "develop",

		authRepository: "rudderstack/rudder-auth",
		authTag:        "develop",

		configBackendRepository: "rudderstack/rudder-config-backend",
		configBackendTag:        "master",
	}

	for _, opt := range opts {
		opt(conf)
	}

	// create network
	network, err := pool.CreateNetwork("rudderauth-" + suffix)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}
	d.Cleanup(func() {
		if err := pool.RemoveNetwork(network); err != nil {
			fmt.Printf("failed to remove rudderauth network: %v", err)
		}
	})

	// start postgres
	postgresResource, err := postgres.Setup(pool, d, postgres.WithNetwork(network.Network))
	if err != nil {
		return nil, fmt.Errorf("starting postgres container: %w", err)
	}
	resource.Postgres = postgresResource

	// helper to start a container and wait for it to be healthy
	startContainer := func(name, repo, tag, port string, env []string) (*dockertest.Resource, string, error) {
		repository := registry.ImagePath(repo)
		if err := pool.Client.PullImage(docker.PullImageOptions{Repository: repository, Tag: tag}, registry.AuthConfiguration()); err != nil {
			return nil, "", fmt.Errorf("pulling image %s:%s: %w", repo, tag, err)
		}
		container, err := pool.RunWithOptions(&dockertest.RunOptions{
			Name:         name,
			Repository:   repository,
			Tag:          tag,
			PortBindings: internal.IPv4PortBindings([]string{port}, internal.WithBindIP(conf.bindIP)),
			NetworkID:    network.Network.ID,
			Env:          env,
			Auth:         registry.AuthConfiguration(),
		}, internal.DefaultHostConfig)
		if err != nil {
			return nil, "", fmt.Errorf("starting %s container: %w", repo, err)
		}
		d.Cleanup(func() {
			if err := pool.Purge(container); err != nil {
				fmt.Printf("failed to purge %s container: %v", repo, err)
			}
		})
		url := fmt.Sprintf("http://%s:%s", container.GetBoundIP(port+"/tcp"), container.GetPort(port+"/tcp"))
		if err := pool.Retry(func() (err error) {
			resp, err := http.Get(url + "/health")
			if err != nil {
				return err
			}
			defer func() { httputil.CloseResponse(resp) }()
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				return errors.New(resp.Status)
			}
			return nil
		}); err != nil {
			return nil, "", fmt.Errorf("waiting for %s to be healthy: %w", repo, err)
		}
		return container, url, nil
	}

	g := &errgroup.Group{}

	// start secrets service
	const secretsPort = "5559"
	secretsContainerName := "secrets-" + suffix
	var secretsURL string
	g.Go(func() error {
		var err error
		_, secretsURL, err = startContainer(secretsContainerName, conf.secretsRepository, conf.secretsTag, secretsPort, []string{
			"BUGSNAG_KEY=''",
			"CLUSTER_ENABLED=false",
			"DB_HOST=" + postgresResource.ContainerName,
			"DB_DATABASE=" + postgresResource.Database,
			"DB_USERNAME=" + postgresResource.User,
			"DB_PASSWORD=" + postgresResource.Password,
		})
		if err != nil {
			return err
		}
		resource.SecretsURL = secretsURL

		// create account secret
		reqBody, err := jsonrs.Marshal(map[string]any{
			"key":         "account_" + fixture.ID,
			"value":       fixture.Secret,
			"workspaceId": fixture.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("marshaling secret request body: %w", err)
		}
		resp, err := http.Post(resource.SecretsURL+"/secrets", "application/json", bytes.NewReader(reqBody))
		defer func() { httputil.CloseResponse(resp) }()
		if err != nil {
			return fmt.Errorf("creating secret: %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("creating secret not ok: %s", resp.Status)
		}
		return nil
	})

	// start rudder-auth
	const authPort = "3033"
	authContainerName := "auth-" + suffix
	var authURL string
	g.Go(func() error {
		var err error
		_, authURL, err = startContainer(authContainerName, conf.authRepository, conf.authTag, authPort, lo.MapToSlice(fixture.AuthClientSecrets, func(k, v string) string {
			return fmt.Sprintf("%s_%s=%s", strings.ToUpper(fixture.Type), strings.ToUpper(k), v)
		}))
		if err != nil {
			return err
		}
		resource.AuthURL = authURL
		return nil
	})

	// start config backend
	const configBackendPort = "5000"
	var configBackendURL string
	configBackendContainerName := "config-backend-" + suffix
	g.Go(func() error {
		var err error
		_, configBackendURL, err = startContainer(configBackendContainerName, conf.configBackendRepository, conf.configBackendTag, configBackendPort, []string{
			"ENABLE_LEGACY_ROUTES=true",
			"DB_HOST=" + postgresResource.ContainerName,
			"DB_DATABASE=" + postgresResource.Database,
			"DB_USERNAME=" + postgresResource.User,
			"DB_PASSWORD=" + postgresResource.Password,
			"ADMIN_USERNAMES=" + adminUsername,
			"ADMIN_PASSWORDS=" + adminPassword,
			"HOSTED_SERVICE_SECRETS=" + hostedServiceSecret,
			"DISABLE_CLUSTER_MODE=true",
			"SECRETS_SERVICE_URL=http://" + secretsContainerName + ":" + secretsPort,
			"CLOUD_SOURCES_AUTH_URL=http://" + authContainerName + ":" + authPort,
			"RUDDER_USER_ID=" + adminUsername,
		})
		if err != nil {
			return err
		}
		resource.ConfigBackendURL = configBackendURL
		// seed postgres with user, workspace, destination, destination definition, account & account definition
		if _, err := postgresResource.DB.Exec(`insert into users (id, name, email) values ($1, $2, $3)`,
			adminUsername,
			adminUsername,
			"admin@example.com",
		); err != nil {
			return fmt.Errorf("seeding user: %w", err)
		}
		if _, err := postgresResource.DB.Exec(`insert into workspaces (id, name, token) values ($1, $2, $3)`,
			fixture.WorkspaceID,
			"Test Workspace",
			"test-token",
		); err != nil {
			return fmt.Errorf("seeding workspace: %w", err)
		}

		var accountDefinitionName sql.NullString
		if fixture.HasAccountDefinition {
			accountDefinitionName = sql.NullString{String: fixture.accountDefName(), Valid: true}
			if _, err := postgresResource.DB.Exec(`insert into account_definitions 
				(name, type, "authenticationType", category, config, "optionsSchema", "secretSchema", "uiConfig") 
				values 
				($1, $2, 'oauth', $3, '{}', '{}', '{}', '{}')`,
				accountDefinitionName,
				fixture.Type,
				fixture.Category,
			); err != nil {
				return fmt.Errorf("seeding account definition: %w", err)
			}
		}
		if _, err := postgresResource.DB.Exec(`insert into accounts 
			(id, name, options, role, "userId", "workspaceId", "rudderCategory", "secretVersion", "accountDefinitionName") 
			values 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			fixture.ID,
			"Test Account",
			fixture.Options,
			fixture.Type,
			adminUsername,
			fixture.WorkspaceID,
			fixture.Category,
			1,
			accountDefinitionName,
		); err != nil {
			return fmt.Errorf("seeding account: %w", err)
		}

		if fixture.DestinationID != "" {
			if _, err := postgresResource.DB.Exec(`insert into destination_definitions 
			(id, name, "displayName", config, "responseRules", "configSchema", "connectionConfigSchema") 
			values 
			($1, $2, $3, $4, $5, $6, $7)`,
				"destination-1",
				"Test Destination definition",
				"Test Destination definition",
				`{"auth": {"type": "OAuth"}}`,
				`{}`,
				`{}`,
				`{}`,
			); err != nil {
				return fmt.Errorf("seeding destination definition: %w", err)
			}
			if _, err := postgresResource.DB.Exec(`insert into destinations 
			(id, name, enabled, config, "destinationDefinitionId", "workspaceId", "accountId", "createdBy", "updatedBy") 
			values 
			($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
				fixture.DestinationID,
				"Test Destination",
				true,
				`{"authStatus": "active"}`,
				"destination-1",
				fixture.WorkspaceID,
				fixture.ID,
				adminUsername,
				adminUsername,
			); err != nil {
				return fmt.Errorf("seeding destination: %w", err)
			}
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &resource, nil
}
