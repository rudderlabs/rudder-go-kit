# Config Documentation Generator

## Context

Configuration options are scattered across the codebase as `conf.GetXxxVar(...)` calls. The README had a manually maintained config table that could drift out of sync. This tool parses Go source code, extracts config options (keys, defaults, descriptions from annotations), and generates a grouped markdown table automatically.

## Architecture

A Go CLI tool at `cmd/configdoc/` that uses `go/ast` to find all config getter calls, extract their parameters, and output a grouped, ordered markdown table.

### Files

- **`cmd/configdoc/configdoc.go`** — Core logic: AST parsing, extraction, formatting, env var derivation
- **`cmd/configdoc/main.go`** — CLI entry point with flags
- **`cmd/configdoc/configdoc_test.go`** — Tests using inline Go source strings
- **`Makefile`** — `configdoc` target: `go run ./cmd/configdoc -root . -output docs/configuration.md -prefix PREFIX -warn`
- **`docs/configuration.md`** — Generated output (do not edit manually)
- **`README.md`** — Links to `docs/configuration.md`

## Annotations

All annotations use the `//configdoc:` prefix. Place them on the line(s) above or on the same line as the config getter call.

| Directive | Scope | Description |
|---|---|---|
| `//configdoc:group [N] <Name>` | Sticky (applies to all subsequent calls in the file until overridden) | Sets the group for config entries, with an optional numeric prefix for sort order. Ordered groups appear first (ascending), unordered groups appear last (alphabetically) |
| `//configdoc:description <text>` | Per-call (applies to the next config getter call) | Sets the description for a config entry |
| `//configdoc:varkey <key>` | Per-call (multiple allowed, consumed in order) | Provides a key for non-literal (dynamic) config key arguments |
| `//configdoc:vardefault <value>` | Per-call | Provides a default value for non-literal (dynamic) default arguments |
| `//configdoc:ignore` | Per-call | Excludes the config entry from output |

### Example usage

```go
//configdoc:group 3 HTTP server
//configdoc:description HTTP server port
port := s.conf.GetIntVar(8080, 1, "http.port")

//configdoc:description Read header timeout
timeout := s.conf.GetDurationVar(10, time.Second, "http.timeout")

//configdoc:ignore
conf.GetStringVar("", "k8s.client.key")

//configdoc:varkey workspace.<id>.timeout
conf.GetStringVar("30s", fmt.Sprintf("workspace.%s.timeout", wsID))

//configdoc:vardefault 5s
conf.GetStringVar(computeDefault(), "retry.interval")
```

### Annotation semantics

- **Group inheritance**: `//configdoc:group` is sticky — it applies to all subsequent config calls in the same file until another `//configdoc:group` overrides it. An optional numeric prefix sets the sort order (e.g. `//configdoc:group 1 Server`). Only one ordered occurrence per group is needed across the codebase.
- **Description**: Applies to the immediately following config call only (up to 3 lines above).
- **Varkey**: Consumed in order for non-literal key arguments. Static string literal keys are extracted normally and don't consume varkey directives. If there are non-literal args without matching varkeys, a warning is emitted.
- **Vardefault**: Overrides the extracted default value. Useful when the default argument is a non-literal expression (e.g. a function call or variable).
- **Deduplication**: The same config key can appear in multiple files. Only one occurrence needs annotations — the tool merges description, group, group order, and env keys across occurrences. Conflicting defaults or groups produce warnings.

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
go run ./cmd/configdoc [flags]
  -root string     Project root directory (default ".")
  -output string   Output file path (default: stdout)
  -prefix string   Environment variable prefix (default "PREFIX")
  -warn            Print warnings for missing descriptions/groups to stderr
```

## Limitations

- **static analysis only**: The tool uses `go/ast` (no type checker). It can only extract string literal key arguments. Non-literal keys (e.g. `fmt.Sprintf(...)`) are skipped unless a `//configdoc:varkey` directive is provided.
- **external constants**: Default values that reference external package constants (e.g. `backoff.DefaultInitialInterval`) are rendered as the Go expression, not as their resolved value.
- **single-file group scope**: Group inheritance is per-file. A `//configdoc:group` in one file does not affect another file.

## Verification

```bash
# Generate docs
make configdoc

# Run tests
go test ./cmd/configdoc/...

# Check for warnings (missing descriptions/groups)
go run ./cmd/configdoc -root . -prefix PREFIX -warn
```
