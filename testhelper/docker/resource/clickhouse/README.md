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
  application-specific grants remain the caller's responsibility. The image
  entrypoint uses the database name unquoted in SQL and the user name as an XML
  element, so `Setup` rejects a database outside `[A-Za-z_][A-Za-z0-9_]*` and a
  user outside `[A-Za-z_][A-Za-z0-9_-]*`.
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
- `WithProductionDefaults()` turns off the test tuning described below. Use it
  when a test depends on ClickHouse's default pool sizes, caches or system log
  tables such as `system.part_log`.

## Test tuning

Each server is sized for short-lived tests by default. On 26.3 this cuts the
idle RSS of one server from about 650 MiB to about 160 MiB, its threads from
about 680 to about 100, and its idle CPU by about two thirds.

- Background pools: `background_pool_size` 5 with
  `background_merges_mutations_concurrency_ratio` 5, schedule 4, common 2, and
  move, fetches and distributed schedule 1 each. Mutations and `OPTIMIZE` need
  pool size × ratio of at least 25, so keep that product when you override them.
- Caches: a 16 MiB mark cache, and no uncompressed or index mark cache.
- Asynchronous metrics refresh every 600 s instead of every second.
- These system log tables are disabled: `metric_log`, `trace_log`, `text_log`,
  `asynchronous_metric_log`, `part_log`, `processors_profile_log`,
  `opentelemetry_span_log`, `query_thread_log`, `query_views_log`, `crash_log`
  and `background_schedule_pool_log`. `query_log` stays enabled.

The tuning sets no global thread pool limit, because ClickHouse hangs at startup
below about 2000 threads. It sets no `max_server_memory_usage` either, because
ClickHouse derives it from the container memory limit. The tuning lives in the
generated resource configuration, so a `WithConfig` document overrides any value.

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
dropped before returning. A failed `Setup` tears down its partial state before it
returns and registers nothing with `d`. A successful `Setup` registers one
`d.Cleanup` that closes clients, purges nodes, removes owned networks and deletes
files. Wait for `Setup` to return before the test ends.
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
