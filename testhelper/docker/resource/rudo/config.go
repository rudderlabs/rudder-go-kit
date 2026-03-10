package rudo

import (
	"fmt"
	"maps"
	"strings"

	"github.com/ory/dockertest/v3/docker"
)

type config struct {
	tag          string
	network      *docker.Network
	bindIP       string
	extraEnvVars map[string]string
}

type Option func(*config)

// WithTag allows specifying the tag of the rudo image to use. Defaults to "latest".
func WithTag(tag string) Option {
	return func(c *config) {
		c.tag = tag
	}
}

// WithNetwork allows specifying a Docker network for the rudo container to connect to.
func WithNetwork(network *docker.Network) Option {
	return func(c *config) {
		c.network = network
	}
}

// WithBindIP allows specifying the bind IP for the rudo container's port bindings.
func WithBindIP(bindIP string) Option {
	return func(c *config) {
		c.bindIP = bindIP
	}
}

// WithExtraEnvVars allows specifying additional environment variables for the rudo container.
func WithExtraEnvVars(extraEnvVars map[string]string) Option {
	return func(c *config) {
		if c.extraEnvVars == nil {
			c.extraEnvVars = make(map[string]string)
		}
		maps.Copy(c.extraEnvVars, extraEnvVars)
	}
}

// WithReleaseName sets the RELEASE_NAME environment variable for the rudo container, Defaults to "default".
func WithReleaseName(releaseName string) Option {
	return WithExtraEnvVars(map[string]string{
		"RELEASE_NAME": releaseName,
	})
}

// WithEtcdHosts sets the ETCD_HOSTS environment variable for the rudo container. Defaults to "http://localhost:2379".
func WithEtcdHosts(etcdHosts []string) Option {
	return WithExtraEnvVars(map[string]string{
		"ETCD_HOSTS": strings.Join(etcdHosts, ","),
	})
}

// WithGatewaySeparateService sets the RUDO_GATEWAY_SEPARATE_SERVICE environment variable for the rudo container. Defaults to "false".
func WithGatewaySeparateService(enabled bool) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_GATEWAY_SEPARATE_SERVICE": fmt.Sprintf("%v", enabled),
	})
}

// WithPartitionCount sets the PARTITION_COUNT environment variable for the rudo container. Defaults to "64".
func WithPartitionCount(count int) Option {
	return WithExtraEnvVars(map[string]string{
		"PARTITION_COUNT": fmt.Sprintf("%d", count),
	})
}

// WithSchedulerEnabled sets the RUDO_SCHEDULER_ENABLED environment variable for the rudo container. Defaults to "false".
func WithSchedulerEnabled(enabled bool) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_SCHEDULER_ENABLED": fmt.Sprintf("%v", enabled),
	})
}

// WithWorkspacePartitionGroups sets the RUDO_WORKSPACE_PARTITION_GROUPS environment variable for the rudo container. Defaults to "10".
func WithWorkspacePartitionGroups(count int) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_WORKSPACE_PARTITION_GROUPS": fmt.Sprintf("%d", count),
	})
}

// WithGatewayNodesPattern sets the RUDO_CLUSTERMANAGER_STATIC_GATEWAY_NODES_PATTERN environment variable for the rudo container. Defaults to "gw-node-*".
func WithGatewayNodesPattern(pattern string) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_CLUSTERMANAGER_STATIC_GATEWAY_NODES_PATTERN": pattern,
	})
}

// WithSrcRouterNodes sets the RUDO_CLUSTERMANAGER_STATIC_SRCROUTER_NODES environment variable for the rudo container. Defaults to "srcrouter-node-0".
func WithSrcRouterNodes(nodes []string) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_CLUSTERMANAGER_STATIC_SRCROUTER_NODES": strings.Join(nodes, ","),
	})
}

// WithInititalReplicaCount sets the RUDO_CLUSTERMANAGER_STATIC_INITIAL_REPLICA_COUNT environment variable for the rudo container. Defaults to "2".
func WithInitialReplicaCount(count int) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_CLUSTERMANAGER_STATIC_INITIAL_REPLICA_COUNT": fmt.Sprintf("%d", count),
	})
}

// WithStaticWorkspaces sets a static list of workspaces for the rudo container. Defaults to ["ws-1,"ws-2"]
func WithStaticWorkspaces(workspaces []string) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_NEWWORKSPACE_POLLER_STATIC_LIST_ENABLED": "true",
		"RUDO_NEWWORKSPACE_POLLER_STATIC_LIST":         strings.Join(workspaces, " "),
	})
}

// WithPollerBaseURL sets the base URL for the new workspace poller in the rudo container. By default static workspaces are used.
func WithPollerBaseURL(url string) Option {
	return WithExtraEnvVars(map[string]string{
		"RUDO_NEWWORKSPACE_POLLER_STATIC_LIST_ENABLED": "false",
		"RUDO_NEWWORKSPACE_POLLER_BASE_URL":            url,
	})
}

// WithWorkspaceNamespace sets the WORKSPACE_NAMESPACE environment variable for the rudo container. Defaults to "default".
func WithWorkspaceNamespace(namespace string) Option {
	return WithExtraEnvVars(map[string]string{
		"WORKSPACE_NAMESPACE": namespace,
	})
}
