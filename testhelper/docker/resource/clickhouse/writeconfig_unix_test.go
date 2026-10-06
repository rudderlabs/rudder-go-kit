//go:build unix

package clickhouse

import (
	"syscall"
	"testing"
)

// The server reads the bind-mounted files as uid 101, so their mode must not depend on the host umask.
// The umask is process-wide, so this test must not run in parallel.
func TestWriteConfigIgnoresUmask(t *testing.T) {
	previous := syscall.Umask(0o077)
	t.Cleanup(func() { syscall.Umask(previous) })
	TestWriteConfig(t)
}
