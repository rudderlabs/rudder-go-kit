# ClickHouse test resource

`Setup(pool *dockertest.Pool, d resource.Cleaner, opts ...Opt) (*Resource, error)`
starts one server by default, or a cluster with `WithCluster(shards, replicas)`.
The default image is the shared 26.3.33.24 fixture pin:
`clickhouse/clickhouse-server:26.3@sha256:810861a2e2d0188744f5f23b2d3ec9ff95812bcb9ddbb8fed13a377a7f305893`.
This is a reproducible fixture version, not a claim about the latest release.

```go
pool, err := dockertest.NewPool("")
require.NoError(t, err)
r, err := clickhouse.Setup(pool, t,
    clickhouse.WithCluster(2, 2),
    clickhouse.WithTLS(),
    clickhouse.WithPrintLogsOnError(true),
)
require.NoError(t, err)
var one uint8
require.NoError(t, r.DB.QueryRow("SELECT 1").Scan(&one))
```

## API

- `Config` and `Opt func(*Config)` follow the Postgres resource convention.
- `Resource` embeds the first `*Node`, so `DB`, `Host`, `HTTPPort`, `NativePort`,
  `User`, `Password`, `Database`, container identity and connection helpers work
  identically for single servers and clusters. `Nodes` contains every node;
  `ClusterName` is empty outside cluster mode. Ports are strings, like Postgres.
- Each node has its Docker `Hostname` and `Macros` (`Cluster`, `Shard`, `Replica`).
  Replicas in one shard share the shard macro; replica macros are unique.
- `NativeDSN()` and `HTTPDSN()` escape credentials and database names. With TLS,
  they select secure native/HTTPS connections, but do not embed CA trust.
  `OpenDB(ch.Native)` / `OpenDB(ch.HTTP)` return standard `*sql.DB` handles using
  clickhouse-go/v2 and the generated CA. Callers must close these additional
  handles; `Setup` closes the native `DB` on each node automatically.
- `TLSConfig` and `CAPEM` expose per-resource trust for other clients. No global
  TLS registration, environment-based root replacement, or skip-verification.

## Options

- `WithTag(tag)` uses the standard server repository and replaces the default
  digest pin. `WithImage(reference)` accepts custom repositories, tags and
  digests. Later options win. Docker Hub mirror configuration is respected.
- `WithNetwork(network)` reuses a caller-owned network; `WithBindIP(ip)` controls
  published addresses (default `127.0.0.1`). Container-internal ports are not
  published when TLS is enabled: only HTTPS 8443 and secure native 9440 are bound.
- `WithUser`, `WithPassword`, `WithDatabase` configure the fixture administrator
  and database. Defaults: `rudder`, `password`, `rudderdb`. Scoped accounts and
  application-specific grants remain the caller's responsibility.
- `WithTLS()` generates a throwaway CA/server certificate valid for localhost,
  loopback, the configured bind IP and the Docker node hostnames.
- `WithCluster(shards, replicas)` requires positive dimensions. For one shard
  with multiple replicas use `WithCluster(1, n)`; no redundant replicas option.
- `WithConfig(xml)` and `WithUsersConfig(xml)` mount additional complete
  `<clickhouse>...</clickhouse>` documents in `config.d` and `users.d`. They
  preserve image defaults and load after the generated resource configuration.
  Do not override resource-managed ports, credentials, topology or TLS paths.
- `WithEnv(env...)` supplies additional environment variables (e.g. `TZ=UTC`).
  Credential/initialization variables are rejected; use the explicit options.
- `WithMemory(bytes)` sets Docker memory per server (zero uses Docker's default).
  `WithPrintLogsOnError(bool)` prints state and logs on test or setup failure.

## Cluster and lifecycle

All nodes join one network. An automatically created network is removed after
the containers; a supplied network is never removed. The first server embeds a
single-member **ClickHouse Keeper** on internal ports 9181/9234, using the same
digest-pinned image and avoiding another image/version/configuration dependency.
This exercises replication and distributed DDL, **not Keeper high availability**.
The XML client section remains named `zookeeper`: this is ClickHouse's
Keeper-compatible coordination protocol configuration, not a ZooKeeper process.

Generated `remote_servers` entries use native 9000 and interserver replication
uses HTTP 9009 on the private Docker network. TLS secures published client ports,
not node-to-node traffic. Do not use this fixture with production credentials or
on an untrusted shared network. Test keys must be readable by the container UID.

Readiness uses `pool.Retry`: authenticated `/ping` and native `SELECT 1` on every
node, every node's `system.clusters` membership, and a temporary
`ReplicatedMergeTree` created `ON CLUSTER`, checked writable on every node and
dropped before returning. `d.Cleanup` handles successful and partial setups,
closing clients, purging nodes, removing owned networks and deleting files.
Set `pool.MaxWait` before `Setup` if a different retry budget is needed.

## Verification

```sh
go build ./...
go vet ./testhelper/docker/resource/clickhouse/...
go test -short ./testhelper/docker/resource/clickhouse/...
go test -v ./testhelper/docker/resource/clickhouse/...
```

Unit tests cover configuration validation, DSN parsing, topology XML, certificate
verification and mounted files. Docker tests cover native and HTTP queries,
verified TLS (including untrusted-CA rejection), a two-shard/two-replica cluster,
distributed DDL, replica writes/reads, shard isolation and caller-owned network
cleanup. Integration tests skip with `-short` or when Docker cannot be reached;
image/startup/query failures on an accessible daemon fail rather than skip.

The implementation adds upstream `github.com/ClickHouse/clickhouse-go/v2 v2.48.0`
and its required module graph/checksums from the local cache. Offline
`go mod tidy` cannot complete here because the driver's transitive test graph
needs uncached `github.com/testcontainers/testcontainers-go v0.43.0`. Run tidy
with network access before integration verification; no local `replace` or
vendored driver is needed.

Prior art: go-kit Postgres/MySQL/MinIO/Redis and Kafka resources; rudder-sources
`clickhouseresource` (image pin, TLS, administrator readiness); sqlconnect-go
`chtest` (single-file configuration mounts, memory caps; fault proxies excluded);
rudder-server ClickHouse cluster XML (remote servers/macros, modernized to Keeper);
rudder-mini (environment credentials and HTTP healthcheck).
