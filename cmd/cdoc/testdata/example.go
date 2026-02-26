// Package example showcases all cdoc annotation features.
//
// Each config call demonstrates a different capability:
//   - Type families: string, bool, int, float64, duration, reloadable
//   - Multiple hierarchical keys for the same setting
//   - Explicit environment variable keys (UPPERCASE_STYLE)
//   - Non-literal (dynamic) keys via //cdoc:key
//   - Non-literal (dynamic) defaults via //cdoc:default
//   - Group ordering via //cdoc:group with numeric prefix
//   - Ungrouped entries (no group directive)
//   - Missing desc (warns)
//   - Missing key for non-literal key (warns, entry skipped)
//   - Missing default for non-literal default (renders Go expression)
//   - Excluding entries via //cdoc:ignore
//   - Deprecated non-*Var methods are silently skipped
package example

import (
	"fmt"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
)

func setup(conf *config.Config, tenantID string, computedTimeout int) {
	// ── No group set — lands in "Ungrouped" ──────────────────────────

	//cdoc:desc Application deployment name
	conf.GetStringVar("rudder-app", "deploymentName", "RELEASE_NAME")

	// ── Server (order 1) ─────────────────────────────────────────────

	//cdoc:group 1 Server

	// String with multiple hierarchical keys
	//cdoc:desc Hostname the server binds to
	conf.GetStringVar("localhost", "server.http.host", "server.host")

	// Int with min value
	//cdoc:desc HTTP server port
	conf.GetIntVar(8080, 1, "server.http.port")

	// Duration
	//cdoc:desc Read header timeout
	conf.GetDurationVar(10, time.Second, "server.http.readHeaderTimeout")

	// Bool with an explicit env-var key
	//cdoc:desc Enable TLS
	conf.GetBoolVar(false, "server.http.tls.enabled", "TLS_ENABLED")

	// Float64
	//cdoc:desc Rate limit (requests per second)
	conf.GetFloat64Var(100.0, "server.http.rateLimit")

	// Reloadable int
	//cdoc:desc Maximum request body size
	conf.GetReloadableIntVar(1048576, 1, "server.http.maxRequestBodySize")

	// ── Tenant (order 2) — dynamic keys and defaults ─────────────────

	//cdoc:group 2 Tenant

	// Non-literal key via key
	//cdoc:desc Per-tenant request timeout
	//cdoc:key tenant.<id>.requestTimeout
	conf.GetDurationVar(30, time.Second, fmt.Sprintf("tenant.%s.requestTimeout", tenantID))

	// Non-literal key AND non-literal default via key + default
	//cdoc:desc Per-tenant max connections
	//cdoc:key tenant.<id>.maxConnections
	//cdoc:default 100
	conf.GetIntVar(computedTimeout, 1, fmt.Sprintf("tenant.%s.maxConnections", tenantID))

	// ── Warning scenarios (no group order — appears after ordered groups) ─

	//cdoc:group Warning scenarios

	// Missing desc — warns with -warn flag
	conf.GetBoolVar(true, "missingDescription")

	// Non-literal key without key — warns, entry is skipped
	//cdoc:desc Per-tenant cache TTL
	conf.GetDurationVar(60, time.Second, fmt.Sprintf("tenant.%s.cacheTTL", tenantID))

	// Non-literal default without default — renders Go expression as default
	//cdoc:desc Worker pool size
	conf.GetIntVar(computedTimeout, 1, "nonLiteralDefaultValue.withoutVardefault")

	// ── Ignored and deprecated entries ────────────────────────────────

	// Excluded from output
	//cdoc:ignore
	conf.GetStringVar("", "internal.secret")

	// Deprecated non-*Var methods — silently skipped by the tool
	conf.GetBool("legacy.enabled", true)
	conf.GetDuration("legacy.timeout", 30, time.Second)
}
