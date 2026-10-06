package clickhouse

import (
	"bytes"
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/distribution/reference"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"

	"github.com/rudderlabs/rudder-go-kit/httputil"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/internal"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/registry"
)

type Node struct {
	DB            *sql.DB
	Host          string
	HTTPPort      string
	NativePort    string
	User          string
	Password      string
	Database      string
	Hostname      string
	ContainerName string
	ContainerID   string
	Macros        Macros
	TLSConfig     *tls.Config
	// PlainHTTPPort is the published plain HTTP port 8123. In TLS mode it is empty without WithPlainHTTPPort.
	PlainHTTPPort string

	secure bool
}

type Resource struct {
	*Node
	Nodes       []*Node
	ClusterName string
	NetworkID   string
	CAPEM       []byte
}

func (node *Node) NativeDSN() string {
	return node.dsn("clickhouse", node.NativePort)
}

func (node *Node) HTTPDSN() string {
	scheme := "http"
	if node.secure {
		scheme = "https"
	}
	return node.dsn(scheme, node.HTTPPort)
}

func (node *Node) dsn(scheme, port string) string {
	uri := url.URL{Scheme: scheme, Host: net.JoinHostPort(node.Host, port), User: url.UserPassword(node.User, node.Password), Path: "/" + node.Database}
	if node.secure {
		uri.RawQuery = url.Values{"secure": {"true"}}.Encode()
	}
	return uri.String()
}

func (node *Node) OpenDB(protocol ch.Protocol) (*sql.DB, error) {
	var dsn string
	switch protocol {
	case ch.Native:
		dsn = node.NativeDSN()
	case ch.HTTP:
		dsn = node.HTTPDSN()
	default:
		return nil, fmt.Errorf("unsupported ClickHouse protocol %d", protocol)
	}
	options, err := ch.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parsing ClickHouse connection: %w", err)
	}
	if node.TLSConfig != nil {
		options.TLS = node.TLSConfig.Clone()
	}
	options.DialTimeout = 5 * time.Second
	options.ReadTimeout = 30 * time.Second
	return ch.OpenDB(options), nil
}

// Setup starts the resource and registers its cleanup with d only after it succeeds.
// The caller must wait for Setup to return before the test ends.
func Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Opt) (_ *Resource, err error) {
	if pool == nil || d == nil {
		return nil, fmt.Errorf("pool and cleaner must not be nil")
	}
	config := defaultConfig()
	for _, option := range opts {
		option(&config)
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "clickhouse-resource-")
	if err != nil {
		return nil, fmt.Errorf("creating ClickHouse configuration directory: %w", err)
	}
	var containers []*dockertest.Resource
	var network *dockertest.Network
	var client *http.Client
	result := &Resource{NetworkID: config.NetworkID}
	teardown := func(failed bool) {
		if config.PrintLogsOnError && failed {
			for _, container := range containers {
				printLogs(pool, d, container)
			}
		}
		for _, node := range result.Nodes {
			if node.DB != nil {
				if err := node.DB.Close(); err != nil {
					d.Log("Closing ClickHouse database:", err)
				}
			}
		}
		if client != nil {
			client.CloseIdleConnections()
		}
		for _, container := range slices.Backward(containers) {
			if err := pool.Purge(container); err != nil {
				d.Log("Purging ClickHouse container:", err)
			}
		}
		for _, node := range slices.Backward(result.Nodes) {
			if node.ContainerName != "" && node.ContainerID == "" {
				if err := pool.RemoveContainerByName(node.ContainerName); err != nil {
					d.Log("Purging incomplete ClickHouse container:", err)
				}
			}
		}
		if network != nil {
			if err := pool.RemoveNetwork(network); err != nil {
				d.Log("Removing ClickHouse network:", err)
			}
		}
		if err := os.RemoveAll(dir); err != nil {
			d.Log("Removing ClickHouse configuration:", err)
		}
	}
	// Setup may run outside the test goroutine, so it registers no cleanup until all state is written.
	// A panic leaves err nil, so it tears down before the panic continues.
	defer func() {
		if r := recover(); r != nil {
			teardown(true)
			panic(r)
		}
		if err != nil {
			teardown(true)
		}
	}()
	if err := os.Chmod(dir, 0o755); err != nil {
		return nil, fmt.Errorf("making ClickHouse configuration readable: %w", err)
	}
	prefix := filepath.Base(dir)
	result.Nodes = topology(prefix, config)
	result.Node = result.Nodes[0]
	if config.Shards > 0 {
		result.ClusterName = clusterName
		if config.NetworkID == "" {
			network, err = pool.CreateNetwork(prefix)
			if err != nil {
				return nil, fmt.Errorf("creating ClickHouse network: %w", err)
			}
			result.NetworkID = network.Network.ID
		}
	}
	var certPEM, keyPEM []byte
	if config.TLS {
		hostnames := make([]string, 0, len(result.Nodes))
		for _, node := range result.Nodes {
			hostnames = append(hostnames, node.Hostname)
		}
		var ca *tls.Certificate
		if config.CACertPEM != nil || config.CAKeyPEM != nil {
			if ca, err = parseCertificateAuthority(config.CACertPEM, config.CAKeyPEM); err != nil {
				return nil, err
			}
		}
		result.TLSConfig, result.CAPEM, certPEM, keyPEM, err = newTLSFixture(ca, config.BindIP, hostnames, !config.NoIPSANs)
		if err != nil {
			return nil, fmt.Errorf("generating ClickHouse TLS certificates: %w", err)
		}
	}
	transport := &http.Transport{TLSClientConfig: result.TLSConfig}
	client = &http.Client{Transport: transport, Timeout: 5 * time.Second}
	repository, tag, err := splitImage(config.Image)
	if err != nil {
		return nil, err
	}
	repository = registry.ImagePath(repository)
	if err := pullImage(pool, repository, tag); err != nil {
		return nil, fmt.Errorf("pulling ClickHouse image: %w", err)
	}
	for index, node := range result.Nodes {
		node.TLSConfig = result.TLSConfig
		mounts, err := writeConfig(dir, config, result.Nodes, index, certPEM, keyPEM)
		if err != nil {
			return nil, fmt.Errorf("writing ClickHouse configuration: %w", err)
		}
		httpPort, nativePort := "8123", "9000"
		if config.TLS {
			httpPort, nativePort = "8443", "9440"
		}
		ports := []string{httpPort, nativePort}
		if config.TLS && config.PlainHTTPPort {
			ports = append(ports, "8123")
		}
		node.ContainerName = node.Hostname
		container, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: repository, Tag: tag, Auth: registry.AuthConfiguration(),
			Name: node.Hostname, Hostname: node.Hostname, NetworkID: result.NetworkID,
			Env: append([]string{
				"CLICKHOUSE_USER=" + config.User, "CLICKHOUSE_PASSWORD=" + config.Password,
				"CLICKHOUSE_DB=" + config.Database, "CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1",
			}, config.Env...),
			Mounts: mounts, ExposedPorts: exposedPorts(ports),
			PortBindings: internal.IPv4PortBindings(ports, internal.WithBindIP(config.BindIP)),
		}, internal.DefaultHostConfig, func(host *docker.HostConfig) { host.Memory = config.Memory })
		if err != nil {
			orphan, inspectErr := pool.Client.InspectContainer(node.Hostname)
			if inspectErr == nil {
				node.ContainerID = orphan.ID
				containers = append(containers, &dockertest.Resource{Container: orphan})
			} else {
				d.Log("Inspecting incomplete ClickHouse startup:", inspectErr)
			}
			return nil, fmt.Errorf("starting ClickHouse node %s: %w", node.Hostname, err)
		}
		containers = append(containers, container)
		if container.Container, err = waitForPortBindings(context.Background(), pool.Client.InspectContainer, container.Container, ports, portBindingsTimeout); err != nil {
			return nil, err
		}
		node.Host = container.GetBoundIP(httpPort + "/tcp")
		if address := net.ParseIP(node.Host); address != nil && address.IsUnspecified() {
			node.Host = "127.0.0.1"
			if address.To4() == nil {
				node.Host = "::1"
			}
		}
		node.HTTPPort, node.NativePort = container.GetPort(httpPort+"/tcp"), container.GetPort(nativePort+"/tcp")
		node.PlainHTTPPort = container.GetPort("8123/tcp")
		node.ContainerName, node.ContainerID = strings.TrimPrefix(container.Container.Name, "/"), container.Container.ID
		if node.Host == "" || node.HTTPPort == "" || node.NativePort == "" {
			return nil, fmt.Errorf("reading ClickHouse node port bindings")
		}
		if err := verifyImage(pool, container, repository, tag); err != nil {
			return nil, err
		}
		node.DB, err = node.OpenDB(ch.Native)
		if err != nil {
			return nil, err
		}
	}
	for _, node := range result.Nodes {
		if err := pool.Retry(func() error { return ready(client, node) }); err != nil {
			return nil, fmt.Errorf("waiting for ClickHouse node %s: %w", node.Hostname, err)
		}
	}
	if config.Shards > 0 {
		if err := pool.Retry(func() error { return clusterReady(result) }); err != nil {
			return nil, fmt.Errorf("waiting for ClickHouse cluster: %w", err)
		}
	}
	d.Cleanup(func() { teardown(d.Failed()) })
	return result, nil
}

func ready(client *http.Client, node *Node) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	scheme := "http"
	if node.secure {
		scheme = "https"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, scheme+"://"+net.JoinHostPort(node.Host, node.HTTPPort)+"/ping", nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(node.User, node.Password)
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { httputil.CloseResponse(response) }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("ClickHouse ping returned status %d", response.StatusCode)
	}
	var one uint8
	if err := node.DB.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		return fmt.Errorf("querying ClickHouse as configured user: %w", err)
	}
	if one != 1 {
		return fmt.Errorf("ClickHouse SELECT 1 returned %d", one)
	}
	return nil
}

func clusterReady(result *Resource) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	for _, node := range result.Nodes {
		var count uint64
		if err := node.DB.QueryRowContext(ctx, "SELECT count() FROM system.clusters WHERE cluster = ?", result.ClusterName).Scan(&count); err != nil {
			return err
		}
		if count != uint64(len(result.Nodes)) {
			return fmt.Errorf("cluster has %d nodes, expected %d", count, len(result.Nodes))
		}
	}
	table := quoteIdentifier(result.Database) + ".`__dockertest_readiness`"
	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s ON CLUSTER %s (id UInt64) ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/dockertest_readiness', '{replica}') ORDER BY id", table, quoteIdentifier(result.ClusterName))
	if _, err := result.DB.ExecContext(ctx, query); err != nil {
		return err
	}
	for _, node := range result.Nodes {
		var count uint64
		if err := node.DB.QueryRowContext(ctx, "SELECT count() FROM system.replicas WHERE database = ? AND table = '__dockertest_readiness' AND is_readonly = 0 AND is_session_expired = 0", result.Database).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return fmt.Errorf("replicated table on %s is not ready", node.Hostname)
		}
	}
	marker := uint64(time.Now().UnixNano())
	insert := fmt.Sprintf("INSERT INTO %s (id) VALUES (?)", table) //nolint:gosec // table is quoteIdentifier output
	seenShards := make(map[string]bool)
	for _, node := range result.Nodes {
		if seenShards[node.Macros.Shard] {
			continue
		}
		seenShards[node.Macros.Shard] = true
		if _, err := node.DB.ExecContext(ctx, insert, marker); err != nil {
			return fmt.Errorf("inserting ClickHouse readiness row on %s: %w", node.Hostname, err)
		}
	}
	for _, node := range result.Nodes {
		if err := waitForReadinessRow(ctx, node, table, marker); err != nil {
			return err
		}
	}
	_, err := result.DB.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s ON CLUSTER %s SYNC", table, quoteIdentifier(result.ClusterName)))
	return err
}

func waitForReadinessRow(ctx context.Context, node *Node, table string, marker uint64) error {
	query := fmt.Sprintf("SELECT count() FROM %s WHERE id = ?", table) //nolint:gosec // table is quoteIdentifier output
	for {
		var count uint64
		if err := node.DB.QueryRowContext(ctx, query, marker).Scan(&count); err == nil && count == 1 {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("waiting for ClickHouse readiness row on replica %s: %w", node.Hostname, ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func quoteIdentifier(value string) string {
	return "`" + strings.NewReplacer("\\", "\\\\", "`", "\\`").Replace(value) + "`"
}

func writeConfig(dir string, config Config, nodes []*Node, index int, certPEM, keyPEM []byte) ([]string, error) {
	nodeDir := filepath.Join(dir, strconv.Itoa(index))
	if err := os.Mkdir(nodeDir, 0o755); err != nil {
		return nil, err
	}
	type mountedFile struct{ name, destination, content string }
	files := []mountedFile{
		{"resource.xml", "/etc/clickhouse-server/config.d/resource.xml", serverConfig(config, nodes, index)},
	}
	if config.ConfigXML != "" {
		files = append(files, mountedFile{"extra.xml", "/etc/clickhouse-server/config.d/zz_extra.xml", config.ConfigXML})
	}
	if config.UsersXML != "" {
		files = append(files, mountedFile{"users.xml", "/etc/clickhouse-server/users.d/zz_extra.xml", config.UsersXML})
	}
	if config.TLS {
		files = append(files,
			mountedFile{"server.pem", "/etc/clickhouse-server/certs/server.pem", string(certPEM)},
			mountedFile{"server.key", "/etc/clickhouse-server/certs/server.key", string(keyPEM)},
		)
	}
	mounts := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(nodeDir, file.name)
		if err := os.WriteFile(path, []byte(file.content), 0o644); err != nil {
			return nil, err
		}
		// The umask can clear the read bits, and the server reads the bind-mounted file as uid 101.
		if err := os.Chmod(path, 0o644); err != nil {
			return nil, err
		}
		mounts = append(mounts, path+":"+file.destination+":ro")
	}
	return mounts, nil
}

func pullImage(pool *dockertest.Pool, repository, tag string) error {
	if _, err := pool.Client.InspectImage(repository + ":" + tag); err == nil {
		return nil
	}
	if _, digest, pinned := strings.Cut(tag, "@"); pinned {
		return pool.Client.PullImage(docker.PullImageOptions{Repository: repository + "@" + digest}, registry.AuthConfiguration())
	}
	return pool.Client.PullImage(docker.PullImageOptions{Repository: repository, Tag: tag}, registry.AuthConfiguration())
}

func verifyImage(pool *dockertest.Pool, container *dockertest.Resource, repository, tag string) error {
	_, digest, pinned := strings.Cut(tag, "@")
	if !pinned {
		return nil
	}
	image, err := pool.Client.InspectImage(container.Container.Image)
	if err != nil {
		return fmt.Errorf("inspecting ClickHouse image: %w", err)
	}
	if !matchesDigest(image.RepoDigests, repository, digest) {
		return fmt.Errorf("ClickHouse image does not match requested digest %s", digest)
	}
	return nil
}

func matchesDigest(repoDigests []string, repository, digest string) bool {
	requested, err := reference.ParseNormalizedNamed(repository + "@" + digest)
	if err != nil {
		return false
	}
	return slices.ContainsFunc(repoDigests, func(repoDigest string) bool {
		actual, err := reference.ParseNormalizedNamed(repoDigest)
		return err == nil && actual.String() == requested.String()
	})
}

func printLogs(pool *dockertest.Pool, d resource.Logger, container *dockertest.Resource) {
	state, err := pool.Client.InspectContainer(container.Container.ID)
	if err != nil {
		d.Log("Inspecting ClickHouse container:", err)
	} else {
		d.Log("ClickHouse container state:", state.Name, state.State.Status)
	}
	var buffer bytes.Buffer
	if err := pool.Client.Logs(docker.LogsOptions{Container: container.Container.ID, Stdout: true, Stderr: true, OutputStream: &buffer, ErrorStream: &buffer}); err != nil {
		d.Log("Reading ClickHouse container logs:", err)
	}
	d.Log("ClickHouse container logs:", container.Container.Name, buffer.String())
}
