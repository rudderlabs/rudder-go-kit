// Package example showcases all configdoc annotation features.
//
// Each config call demonstrates a different capability:
//   - Type families: string, bool, int, float64, duration, reloadable
//   - Multiple hierarchical keys for the same setting
//   - Explicit environment variable keys (UPPERCASE_STYLE)
//   - Non-literal (dynamic) keys via //configdoc:varkey
//   - Non-literal (dynamic) defaults via //configdoc:vardefault
//   - Group ordering via //configdoc:group with numeric prefix
//   - Ungrouped entries (no group directive)
//   - Missing description (warns)
//   - Missing varkey for non-literal key (warns, entry skipped)
//   - Missing vardefault for non-literal default (renders Go expression)
//   - Excluding entries via //configdoc:ignore
//   - Deprecated non-*Var methods are silently skipped
package example

import (
	"fmt"
	"time"

	"github.com/rudderlabs/rudder-go-kit/config"
)

func setup(conf *config.Config, tenantID string, computedTimeout int) {
	// ── No group set — lands in "Ungrouped" ──────────────────────────

	//configdoc:description Application deployment name
	conf.GetStringVar("rudder-app", "deploymentName", "RELEASE_NAME")

	// ── Server (order 1) ─────────────────────────────────────────────

	//configdoc:group 1 Server

	// String with multiple hierarchical keys
	//configdoc:description Hostname the server binds to
	conf.GetStringVar("localhost", "server.http.host", "server.host")

	// Int with min value
	//configdoc:description HTTP server port
	conf.GetIntVar(8080, 1, "server.http.port")

	// Duration
	//configdoc:description Read header timeout
	conf.GetDurationVar(10, time.Second, "server.http.readHeaderTimeout")

	// Bool with an explicit env-var key
	//configdoc:description Enable TLS
	conf.GetBoolVar(false, "server.http.tls.enabled", "TLS_ENABLED")

	// Float64
	//configdoc:description Rate limit (requests per second)
	conf.GetFloat64Var(100.0, "server.http.rateLimit")

	// Reloadable int
	//configdoc:description Maximum request body size
	conf.GetReloadableIntVar(1048576, 0, "server.http.maxRequestBodySize")

	// ── Tenant (order 2) — dynamic keys and defaults ─────────────────

	//configdoc:group 2 Tenant

	// Non-literal key via varkey
	//configdoc:description Per-tenant request timeout
	//configdoc:varkey tenant.<id>.requestTimeout
	conf.GetDurationVar(30, time.Second, fmt.Sprintf("tenant.%s.requestTimeout", tenantID))

	// Non-literal key AND non-literal default via varkey + vardefault
	//configdoc:description Per-tenant max connections
	//configdoc:varkey tenant.<id>.maxConnections
	//configdoc:vardefault 100
	conf.GetIntVar(computedTimeout, 1, fmt.Sprintf("tenant.%s.maxConnections", tenantID))

	// ── Warning scenarios (no group order — appears after ordered groups) ─

	//configdoc:group Warning scenarios

	// Missing description — warns with -warn flag
	conf.GetBoolVar(true, "missingDescription")

	// Non-literal key without varkey — warns, entry is skipped
	//configdoc:description Per-tenant cache TTL
	conf.GetDurationVar(60, time.Second, fmt.Sprintf("tenant.%s.cacheTTL", tenantID))

	// Non-literal default without vardefault — renders Go expression as default
	//configdoc:description Worker pool size
	conf.GetIntVar(computedTimeout, 1, "nonLiteralDefaultValue.withoutVardefault")

	// ── Ignored and deprecated entries ────────────────────────────────

	// Excluded from output
	//configdoc:ignore
	conf.GetStringVar("", "internal.secret")

	// Deprecated non-*Var methods — silently skipped by the tool
	conf.GetBool("legacy.enabled", true)
	conf.GetDuration("legacy.timeout", 30, time.Second)
}
