package rudo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"maps"
	"net/http"
	"net/url"

	"github.com/ory/dockertest/v3"
	"github.com/samber/lo"

	clustertypes "github.com/rudderlabs/rudder-schemas/go/cluster"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

type Resource struct {
	URL string
}

var defaultEnvVars map[string]string = map[string]string{
	"RUDO_ENABLE_STATS": "false",

	"RELEASE_NAME":                    "default",
	"ETCD_HOSTS":                      "http://localhost:2379",
	"RUDO_GATEWAY_SEPARATE_SERVICE":   "false",
	"PARTITION_COUNT":                 "64",
	"RUDO_SCHEDULER_ENABLED":          "false",
	"RUDO_WORKSPACE_PARTITION_GROUPS": "10",

	"RUDO_CLUSTERMANAGER_TYPE":                         "static",
	"RUDO_CLUSTERMANAGER_STATIC_GATEWAY_NODES_PATTERN": "gw-node-*",
	"RUDO_CLUSTERMANAGER_STATIC_SRCROUTER_NODES":       "srcrouter-node-0",
	"RUDO_CLUSTERMANAGER_STATIC_MIN_REPLICA_COUNT":     "1",
	"RUDO_CLUSTERMANAGER_STATIC_MAX_REPLICA_COUNT":     "1024",
	"RUDO_CLUSTERMANAGER_STATIC_INITIAL_REPLICA_COUNT": "2",

	"RUDO_NEWWORKSPACE_ASSIGNER_STRATEGY":          "multi-node-round-robin",
	"RUDO_NEWWORKSPACE_POLLER_STATIC_LIST":         "ws-1 ws-2",
	"RUDO_NEWWORKSPACE_POLLER_STATIC_LIST_ENABLED": "true",
	"RUDO_NEWWORKSPACE_POLLER_BASE_URL":            "http://invalid:8080",
	"WORKSPACE_NAMESPACE":                          "default",
}

func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Option) (*Resource, error) {
	conf := config{
		tag: "latest",
	}
	for _, opt := range opts {
		opt(&conf)
	}
	envs := maps.Clone(defaultEnvVars)
	maps.Copy(envs, conf.extraEnvVars)

	container, err := pool.RunWithOptions(
		&dockertest.RunOptions{
			Repository:   "422074288268.dkr.ecr.us-east-1.amazonaws.com/rudderstack/rudder-orchestrator",
			Tag:          conf.tag,
			ExposedPorts: []string{"8080/tcp"},
			PortBindings: internal.IPv4PortBindings([]string{"8080"}, internal.WithBindIP(conf.bindIP)),
			Auth:         registry.AuthConfiguration(),
			Env: lo.MapToSlice(envs, func(k, v string) string {
				return fmt.Sprintf("%s=%s", k, v)
			}),
		}, internal.DefaultHostConfig)
	d.Cleanup(func() {
		if err := pool.Purge(container); err != nil {
			d.Log("Could not purge resource:", err)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("run redis container: %w", err)
	}

	rudoURL := fmt.Sprintf("http://%s:%s", container.GetBoundIP("8080/tcp"), container.GetPort("8080/tcp"))
	err = pool.Retry(func() error {
		healthURL, _ := url.JoinPath(rudoURL, "/health")
		resp, err := http.Get(healthURL)
		defer func() { httputil.CloseResponse(resp) }()
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &Resource{URL: rudoURL}, nil
}

func (r *Resource) CreateMigration(ctx context.Context, migration []WorkspaceMigration) (*clustertypes.PartitionMigrationInfo, error) {
	body, err := jsonrs.Marshal(migration)
	if err != nil {
		return nil, fmt.Errorf("encoding request body: %w", err)
	}
	migrationURL, _ := url.JoinPath(r.URL, "/orchestrator/v1/migrations")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, migrationURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	defer func() { httputil.CloseResponse(resp) }()
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, string(respBody))
	}
	var migrationInfo clustertypes.PartitionMigrationInfo
	if err := jsonrs.Unmarshal(respBody, &migrationInfo); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &migrationInfo, nil
}

func (r *Resource) ListMigrations(ctx context.Context) ([]clustertypes.PartitionMigrationInfo, error) {
	migrationURL, _ := url.JoinPath(r.URL, "/orchestrator/v1/migrations/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, migrationURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	defer func() { httputil.CloseResponse(resp) }()
	if err != nil {
		return nil, fmt.Errorf("performing request: %w", err)
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, string(respBody))
	}
	var migrations []clustertypes.PartitionMigrationInfo
	if err := jsonrs.Unmarshal(respBody, &migrations); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return migrations, nil
}

type WorkspaceMigration struct {
	WorkspaceID string      `json:"workspaceId"`
	Migrations  []Migration `json:"migration"`
}

type Migration struct {
	Src Src `json:"src"`
	Dst Dst `json:"dst"`
}

type Src struct {
	ServerID      int   `json:"serverId"`
	PartitionIdxs []int `json:"partition_idxs,omitempty"`
}

type Dst struct {
	ServerID int `json:"serverId"`
}
