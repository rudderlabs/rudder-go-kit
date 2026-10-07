package clickhouse

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

// fakeDocker answers the Docker API calls that Setup makes for one container. inspect returns the
// container state for the nth inspect, starting at 1.
type fakeDocker struct {
	mu       sync.Mutex
	inspect  func(n int) map[string]any
	inspects int
	removed  []string
}

var apiVersionPrefix = regexp.MustCompile(`^/v[0-9.]+`)

func (f *fakeDocker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	path := apiVersionPrefix.ReplaceAllString(r.URL.Path, "")
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/images/"):
		_ = json.NewEncoder(w).Encode(map[string]any{"Id": "sha256:fake"})
	case r.Method == http.MethodPost && path == "/containers/create":
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]any{"Id": "fake"})
	case r.Method == http.MethodPost && path == "/containers/fake/start":
		w.WriteHeader(http.StatusNoContent)
	case r.Method == http.MethodGet && path == "/containers/fake/json":
		f.inspects++
		state := f.inspect(f.inspects)
		state["Id"], state["Name"], state["Image"] = "fake", "/fake", "sha256:fake"
		_ = json.NewEncoder(w).Encode(state)
	case r.Method == http.MethodDelete && path == "/containers/fake":
		f.removed = append(f.removed, "fake")
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "unexpected "+r.Method+" "+path, http.StatusNotFound)
	}
}

func fakeDockerPool(t *testing.T, fake *fakeDocker) *dockertest.Pool {
	t.Helper()
	server := httptest.NewServer(fake)
	t.Cleanup(server.Close)
	pool, err := dockertest.NewPool(server.URL)
	require.NoError(t, err)
	pool.MaxWait = time.Second
	return pool
}

func containerState(status string, ports map[string]any) map[string]any {
	return map[string]any{
		"State":           map[string]any{"Status": status, "Running": status == "running", "ExitCode": 137, "OOMKilled": status != "running"},
		"NetworkSettings": map[string]any{"Ports": ports},
	}
}

// closedPort returns a loopback port with no listener, so readiness fails at once.
func closedPort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())
	return port
}

func requireEmptyDir(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Empty(t, entries, "a failed Setup removes its configuration directory")
}

func TestSetupWithFakeDocker(t *testing.T) {
	t.Run("exited container fails fast and is removed", func(t *testing.T) {
		tmp := t.TempDir()
		t.Setenv("TMPDIR", tmp)
		fake := &fakeDocker{inspect: func(int) map[string]any { return containerState("exited", map[string]any{}) }}
		cleaner := &countingCleaner{}
		start := time.Now()
		_, err := Setup(fakeDockerPool(t, fake), cleaner, WithImage("fake/clickhouse:test"))
		require.ErrorContains(t, err, "exited")
		require.ErrorContains(t, err, "exit code 137")
		require.ErrorContains(t, err, "OOM killed")
		require.Less(t, time.Since(start), 5*time.Second, "a final state stops the wait for port bindings")
		require.Equal(t, []string{"fake"}, fake.removed)
		require.Zero(t, cleaner.cleanups)
		requireEmptyDir(t, tmp)
	})

	// Docker Desktop can report a started container before its bindings, either with no port keys or with an
	// empty binding list per port.
	for name, unbound := range map[string]map[string]any{
		"no port keys":        {},
		"empty binding lists": {"8123/tcp": []any{}, "9000/tcp": []any{}},
	} {
		t.Run("bindings appear later with "+name, func(t *testing.T) {
			tmp := t.TempDir()
			t.Setenv("TMPDIR", tmp)
			port := closedPort(t)
			bound := map[string]any{}
			for _, exposed := range []string{"8123/tcp", "9000/tcp"} {
				bound[exposed] = []any{map[string]any{"HostIp": "127.0.0.1", "HostPort": port}}
			}
			fake := &fakeDocker{inspect: func(n int) map[string]any {
				if n < 3 {
					return containerState("running", unbound)
				}
				return containerState("running", bound)
			}}
			_, err := Setup(fakeDockerPool(t, fake), &countingCleaner{}, WithImage("fake/clickhouse:test"))
			require.ErrorContains(t, err, "waiting for ClickHouse node", "Setup reads the late bindings and reaches readiness")
			require.GreaterOrEqual(t, fake.inspects, 3)
			require.Equal(t, []string{"fake"}, fake.removed)
			requireEmptyDir(t, tmp)
		})
	}
}
