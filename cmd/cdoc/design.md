# Config Documentation Generator

## Context

Configuration options are scattered across the codebase as `conf.GetXxxVar(...)` calls. The README had a manually maintained config table that could drift out of sync. This tool parses Go source code, extracts config options (keys, defaults, descriptions from annotations), and generates a grouped markdown table automatically.

## Architecture

A Go CLI tool at `cmd/cdoc/` that uses `go/ast` to find all config getter calls, extract their parameters, and output a grouped, ordered markdown table.

## Annotations

All annotations use the `//cdoc:` prefix. Place them on the line(s) above or on the same line as the config getter call.

| Directive | Scope | Description |
|---|---|---|
| `//cdoc:group [N] <Name>` | Sticky (applies to all subsequent calls in the file until overridden) | Sets the group for config entries, with an optional numeric prefix for sort order. Ordered groups appear first (ascending), unordered groups appear last (alphabetically) |
| `//cdoc:desc <text>` | Per-call (applies to the next config getter call) | Sets the description for a config entry |
| `//cdoc:key <key>[, <key>...]` | Per-call (multiple allowed, consumed in order) | Provides key override(s) for non-literal (dynamic) config key arguments |
| `//cdoc:default <value>` | Per-call | Provides a default value for non-literal (dynamic) default arguments |
| `//cdoc:ignore` | Per-call | Excludes the config entry from output |

### Example usage

```go
//cdoc:group 3 HTTP server
//cdoc:desc HTTP server port
port := s.conf.GetIntVar(8080, 1, "http.port")

//cdoc:desc Read header timeout
timeout := s.conf.GetDurationVar(10, time.Second, "http.timeout")

//cdoc:ignore
conf.GetStringVar("", "k8s.client.key")

//cdoc:key workspace.<id>.timeout
conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID))

//cdoc:default 5s
conf.GetStringVar(computeDefault(), "retry.interval")
```

### Annotation semantics

- **group inheritance**: `//cdoc:group` is sticky — it applies to all subsequent config calls in the same file until another `//cdoc:group` overrides it. An optional numeric prefix sets the sort order (e.g. `//cdoc:group 1 Server`). Only one ordered occurrence per group is needed across the codebase.
- **desc**: Applies to the immediately following config call only (up to 3 lines above).
- **key**: Consumed in order for non-literal key arguments. Static string literal keys are extracted normally and don't consume key directives. If there are non-literal args without matching keys, a warning is emitted.
- **default**: Overrides the extracted default value. Useful when the default argument is a non-literal expression (e.g. a function call or variable).
- **deduplication**: The same config key can appear in multiple files. Only one occurrence needs annotations — the tool merges desc, group, group order, and env keys across occurrences. Conflicting defaults or groups produce warnings.

## Config getter method families

The tool handles these `rudderlabs/rudder-go-kit/config` method signatures:

| Family | Methods | Args layout |
|---|---|---|
| simple | `GetStringVar`, `GetBoolVar`, `GetFloat64Var`, `GetStringSliceVar`, `GetReloadableStringVar`, `GetReloadableFloat64Var` | `(default, keys...)` |
| int | `GetIntVar`, `GetReloadableIntVar` | `(default, min, keys...)` |
| duration | `GetDurationVar`, `GetReloadableDurationVar` | `(quantity, unit, keys...)` |
| int64 | `GetInt64Var` | `(default, multiplier, keys...)` |

## Env variable derivation

Env variables are derived from config keys using the same algorithm as `ConfigKeyToEnv` from `rudder-go-kit/config/config_env.go`:

- camelCase is split into SNAKE_CASE: `healthCheckTimeout` → `HEALTH_CHECK_TIMEOUT`
- Dots become underscores: `http.port` → `HTTP_PORT`
- A configurable prefix is prepended: `HTTP_PORT` → `PREFIX_HTTP_PORT`
- Keys already in UPPERCASE_STYLE (e.g. `RELEASE_NAME`, `ETCD_HOSTS`) are displayed as-is without the prefix

The prefix is configurable via the `-prefix` CLI flag (default: `PREFIX`).

## Default value rendering

- **string/bool/int/float**: rendered as literal values
- **duration**: quantity + unit abbreviation (`10` + `time.Second` → `10s`)
- **int64 with multiplier**: `1` + `bytesize.MB` → `1MB`
- **external constants**: rendered as the Go expression (e.g. `backoff.DefaultInitialInterval`)
- **non-literal expressions**: rendered as Go source text

## CLI flags

```
go run ./cmd/cdoc [flags]
  -root string     Project root directory (default ".")
  -output string   Output file path (default: stdout)
  -prefix string   Environment variable prefix (default "PREFIX")
  -warn            Print warnings for missing descriptions/groups to stderr
```

## Limitations

- **static analysis only**: The tool uses `go/ast` (no type checker). It can only extract string literal key arguments. Non-literal keys (e.g. `fmt.Sprintf(...)`) are skipped unless a `//cdoc:key` directive is provided.
- **external constants**: Default values that reference external package constants (e.g. `backoff.DefaultInitialInterval`) are rendered as the Go expression, not as their resolved value.
- **single-file group scope**: Group inheritance is per-file. A `//cdoc:group` in one file does not affect another file.

## Verification

```bash
# Generate docs
make cdoc

# Run tests
go test ./cmd/cdoc/...

# Check for warnings (missing descriptions/groups)
go run ./cmd/cdoc -root . -prefix PREFIX -warn
```
