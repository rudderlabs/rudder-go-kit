package maxprocs

import (
	"math"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/logger/mock_logger"
)

func TestSet_Default(t *testing.T) {
	before := runtime.GOMAXPROCS(0)  // Capture original value
	defer runtime.GOMAXPROCS(before) // Restore after test

	mockLog := requireLoggerInfo(t, 1.1, 1, 1.5, 0, 2)
	set("1100m", withLogger(mockLog))
	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1100m * 1.5 = 1.65 → ceil = 2
}

func TestSetWithConfig_Default(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "1100m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1.1, 1, 1.5, int64(numCPU), 2)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1100m * 1.5 = 1.65 → ceil = 2
}

func TestSet_WithInvalidCPURequest_Invalid1(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 1, 1, 1.5, 0, 2)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)

	set("invalid", withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSetWithConfig_WithInvalidCPURequest_Invalid1(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "invalid")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1, 1, 1.5, int64(numCPU), 2)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)

	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSet_WithInvalidCPURequest_Invalid2(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 1, 1, 1.5, 0, 2)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)

	set("invalid_m", withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSetWithConfig_WithInvalidCPURequest_Invalid2(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "invalid_m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1, 1, 1.5, int64(numCPU), 2)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)

	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSet_WithMinProcs(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 5, 1.5, 0, 5)
	set("100m",
		withMinProcs(5),
		withLogger(mockLog),
	)

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSetWithConfig_WithMinProcs(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "100m")
	cfg.Set("MinProcs", 5)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 0.1, 5, 1.5, int64(numCPU), 5)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSet_WithMultiplier(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.3, 1, 4, 0, 2)
	set("300m",
		withCPURequestsMultiplier(4),
		withLogger(mockLog),
	)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSetWithConfig_WithMultiplier(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "300m")
	cfg.Set("RequestsMultiplier", 4)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 0.3, 1, 4, int64(numCPU), 2)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSet_CustomRoundQuotaFunc(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int { return int(math.Floor(f)) }

	mockLog := requireLoggerInfo(t, 1.5, 1, 1.5, 0, 2)
	set("1500m",
		withRoundQuotaFunc(roundFloor),
		withLogger(mockLog),
	)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1500m * 1.5 = 2.25 → floor = 2
}

func TestSet_WithMaxProcs(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 10, 1, 1.5, 8, 8)
	set("10000m",
		withMaxProcs(8),
		withLogger(mockLog),
	)

	require.Equal(t, 8, runtime.GOMAXPROCS(0)) // 10000m * 1.5 = 15 → capped at 8
}

func TestSet_WithMaxProcsNoEffect(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 1, 1.5, 10, 1)
	set("100m",
		withMaxProcs(10),
		withLogger(mockLog),
	)

	require.Equal(t, 1, runtime.GOMAXPROCS(0)) // 100m * 1.5 = 0.15 → ceil = 1, max has no effect
}

func TestSetWithConfig_WithMaxProcs(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "10000m")
	cfg.Set("MaxProcs", 8)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 10, 1, 1.5, 8, 8)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 8, runtime.GOMAXPROCS(0)) // 10000m * 1.5 = 15 → capped at 8
}

func TestSetWithConfig_WithMaxProcsNoEffect(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "100m")
	cfg.Set("MaxProcs", 10)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 1, 1.5, 10, 1)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 1, runtime.GOMAXPROCS(0)) // 100m * 1.5 = 0.15 → ceil = 1, max has no effect
}

func TestSetWithConfig_CustomRoundQuotaFunc(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "1500m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int { return int(math.Floor(f)) }

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1.5, 1, 1.5, int64(numCPU), 2)
	setWithConfig(cfg,
		withRoundQuotaFunc(roundFloor),
		withLogger(mockLog),
	)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1500m * 1.5 = 2.25 → floor = 2
}

func TestEnvironmentVariables(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	t.Setenv("MAXPROCS_LOGGER_CONSOLE_JSON_FORMAT", "true")
	t.Setenv("MAXPROCS_REQUESTS", "1100m")

	setDefault()
	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1100m * 1.5 = 1.65 → ceil = 2

	t.Setenv("MAXPROCS_MIN_PROCS", "5")
	setDefault()
	require.Equal(t, 5, runtime.GOMAXPROCS(0))

	t.Setenv("MAXPROCS_REQUESTS_MULTIPLIER", "6")
	setDefault()
	require.Equal(t, 7, runtime.GOMAXPROCS(0))
}

func TestSetWithConfig_ReadFromFile(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	tmpDir := t.TempDir()
	requestsFile := filepath.Join(tmpDir, "cpu_requests")
	require.NoError(t, os.WriteFile(requestsFile, []byte("1500"), 0o644))

	cfg := config.New()
	cfg.Set("RequestsFile", requestsFile)
	cfg.Set("Watch", false)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1.5, 1, 1.5, int64(numCPU), 3)
	mockLog.EXPECT().Infon("Using CPU requests from file",
		logger.NewStringField("requests", "1500m"),
		logger.NewStringField("file", requestsFile),
	).Times(1)

	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // 1500m * 1.5 = 2.25 → ceil = 3
}

func TestSetWithConfig_EmptyFile(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	tmpDir := t.TempDir()
	requestsFile := filepath.Join(tmpDir, "cpu_requests")
	require.NoError(t, os.WriteFile(requestsFile, []byte(""), 0o644))

	cfg := config.New()
	cfg.Set("Requests", "1100m")
	cfg.Set("RequestsFile", requestsFile)
	cfg.Set("Watch", false)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1.1, 1, 1.5, int64(numCPU), 2)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Falls back to Requests config: 1100m * 1.5 = 1.65 → ceil = 2
}

func TestSetWithConfig_NonExistentFile(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	cfg := config.New()
	cfg.Set("Requests", "2000m")
	cfg.Set("RequestsFile", "/non/existent/file")
	cfg.Set("Watch", false)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 2.0, 1, 1.5, int64(numCPU), 3)
	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Falls back to Requests config: 2000m * 1.5 = 3.0 → ceil = 3
}

func TestSetWithConfig_WatchDisabled(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	tmpDir := t.TempDir()
	requestsFile := filepath.Join(tmpDir, "cpu_requests")
	require.NoError(t, os.WriteFile(requestsFile, []byte("1000"), 0o644))

	cfg := config.New()
	cfg.Set("RequestsFile", requestsFile)
	cfg.Set("Watch", false)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1, 1, 1.5, int64(numCPU), 2)
	mockLog.EXPECT().Infon("Using CPU requests from file",
		logger.NewStringField("requests", "1000m"),
		logger.NewStringField("file", requestsFile),
	).Times(1)
	mockLog.EXPECT().Infon("Starting file watcher to monitor CPU requests changes", gomock.Any()).Times(0)

	setWithConfig(cfg, withLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1000m * 1.5 = 1.5 → ceil = 2
}

func TestSetWithConfig_FileWatcherWithChanges(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	tmpDir := t.TempDir()
	requestsFile := filepath.Join(tmpDir, "cpu_requests")
	require.NoError(t, os.WriteFile(requestsFile, []byte("1000"), 0o644))

	cfg := config.New()
	cfg.Set("RequestsFile", requestsFile)
	cfg.Set("Watch", true)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1, 1, 1.5, int64(numCPU), 2)
	mockLog.EXPECT().Infon("Using CPU requests from file",
		logger.NewStringField("requests", "1000m"),
		logger.NewStringField("file", requestsFile),
	).Times(1)
	mockLog.EXPECT().Infon("Starting file watcher to monitor CPU requests changes",
		logger.NewStringField("file", requestsFile),
	).Times(1)
	mockLog.EXPECT().Withn(logger.NewStringField("file", requestsFile)).Return(mockLog).Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured", // 2nd call is for the watcher
		logger.NewFloatField("cpuRequests", 2),
		logger.NewFloatField("multiplier", 1.5),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("maxProcs", int64(numCPU)),
		logger.NewIntField("result", 3),
		logger.NewIntField("GOMAXPROCS", 3),
	).MinTimes(1)

	watcherIsSetup := make(chan struct{})
	mockLog.EXPECT().Debugn("Watching file for changes").Do(func(_ string, _ ...logger.Field) {
		close(watcherIsSetup)
	}).Times(1)

	setWithConfig(cfg, withLogger(mockLog))
	initialProcs := runtime.GOMAXPROCS(0)
	require.Equal(t, 2, initialProcs) // 1000m * 1.5 = 1.5 → ceil = 2

	// Update the file
	select {
	case <-watcherIsSetup:
	case <-time.After(5 * time.Second):
		t.Fatalf("File watcher was not setup within 5 seconds")
	}
	require.NoError(t, os.WriteFile(requestsFile, []byte("2000"), 0o644))
	require.Eventually(t, func() bool {
		return runtime.GOMAXPROCS(0) == 3 // 2000m * 1.5 = 3.0 → ceil = 3
	}, 5*time.Second, 1*time.Second)
}

func TestSetWithConfig_FileWatcherWithSignal(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	tmpDir := t.TempDir()
	requestsFile := filepath.Join(tmpDir, "cpu_requests")
	require.NoError(t, os.WriteFile(requestsFile, []byte("1000"), 0o644))

	cfg := config.New()
	cfg.Set("RequestsFile", requestsFile)
	cfg.Set("Watch", true)

	stop := make(chan os.Signal, 1)

	numCPU := runtime.NumCPU()
	mockLog := requireLoggerInfo(t, 1, 1, 1.5, int64(numCPU), 2)
	mockLog.EXPECT().Infon("Using CPU requests from file",
		logger.NewStringField("requests", "1000m"),
		logger.NewStringField("file", requestsFile),
	).Times(1)
	mockLog.EXPECT().Infon("Starting file watcher to monitor CPU requests changes",
		logger.NewStringField("file", requestsFile),
	).Times(1)
	mockLog.EXPECT().Withn(logger.NewStringField("file", requestsFile)).Return(mockLog).Times(1)

	watcherIsSetup := make(chan struct{})
	mockLog.EXPECT().Debugn("Watching file for changes").Do(func(_ string, _ ...logger.Field) {
		close(watcherIsSetup)
	}).Times(1)
	mockLog.EXPECT().Infon("Received signal, stopping file watcher").Times(1)

	watcherStopped := make(chan struct{})
	mockLog.EXPECT().Debugn("Stopped watching file for changes").Do(func(_ string, _ ...logger.Field) {
		close(watcherStopped)
	}).Times(1)

	setWithConfig(cfg, withLogger(mockLog), withStopFileWatcher(stop))

	// Wait for watcher to be setup
	select {
	case <-watcherIsSetup:
	case <-time.After(5 * time.Second):
		t.Fatalf("File watcher was not setup within 5 seconds")
	}

	// Send signal to stop the watcher
	stop <- os.Interrupt

	// Wait for watcher to stop
	select {
	case <-watcherStopped:
	case <-time.After(5 * time.Second):
		t.Fatalf("File watcher did not stop within 5 seconds")
	}

	require.Equal(t, 2, runtime.GOMAXPROCS(0))
}

func requireLoggerInfo(t testing.TB,
	cpuRequests float64,
	minProcs int64,
	multiplier float64,
	maxProcs int64,
	result int64,
) *mock_logger.MockLogger {
	t.Helper()
	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", cpuRequests),
		logger.NewFloatField("multiplier", multiplier),
		logger.NewIntField("minProcs", minProcs),
		logger.NewIntField("maxProcs", maxProcs),
		logger.NewIntField("result", result),
		logger.NewIntField("GOMAXPROCS", result),
	).Times(1)
	return mockLog
}
