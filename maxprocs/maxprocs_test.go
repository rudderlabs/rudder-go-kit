package maxprocs

import (
	"math"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/logger/mock_logger"
)

func TestSet_Default(t *testing.T) {
	before := runtime.GOMAXPROCS(0)  // Capture original value
	defer runtime.GOMAXPROCS(before) // Restore after test

	mockLog := requireLoggerInfo(t, 1.1, 1, 1.5, 2)
	Set("1100m", WithLogger(mockLog))
	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1100m * 1.5 = 1.65 → ceil = 2
}

func TestSetWithConfig_Default(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "1100m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 1.1, 1, 1.5, 2)
	SetWithConfig(cfg, WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1100m * 1.5 = 1.65 → ceil = 2
}

func TestSet_WithInvalidCPURequest_Invalid1(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 1.5),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 2),
		logger.NewIntField("GOMAXPROCS", 2),
	).Times(1)

	Set("invalid", WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSetWithConfig_WithInvalidCPURequest_Invalid1(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "invalid")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 1.5),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 2),
		logger.NewIntField("GOMAXPROCS", 2),
	).Times(1)

	SetWithConfig(cfg, WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSet_WithInvalidCPURequest_Invalid2(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 1.5),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 2),
		logger.NewIntField("GOMAXPROCS", 2),
	).Times(1)

	Set("invalid_m", WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSetWithConfig_WithInvalidCPURequest_Invalid2(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "invalid_m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 1.5),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 2),
		logger.NewIntField("GOMAXPROCS", 2),
	).Times(1)

	SetWithConfig(cfg, WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // Defaults to 1 * 1.5 → ceil = 2
}

func TestSet_WithMinProcs(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 5, 1.5, 5)
	Set("100m",
		WithMinProcs(5),
		WithLogger(mockLog),
	)

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSetWithConfig_WithMinProcs(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "100m")
	cfg.Set("MinProcs", 5)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 5, 1.5, 5)
	SetWithConfig(cfg, WithLogger(mockLog))

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSet_WithMultiplier(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.3, 1, 4, 2)
	Set("300m",
		WithCPURequestsMultiplier(4),
		WithLogger(mockLog),
	)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSetWithConfig_WithMultiplier(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "300m")
	cfg.Set("RequestsMultiplier", 4)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.3, 1, 4, 2)
	SetWithConfig(cfg, WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSet_CustomRoundQuotaFunc(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int { return int(math.Floor(f)) }

	mockLog := requireLoggerInfo(t, 1.5, 1, 1.5, 2)
	Set("1500m",
		WithRoundQuotaFunc(roundFloor),
		WithLogger(mockLog),
	)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 1500m * 1.5 = 2.25 → floor = 2
}

func TestSetWithConfig_CustomRoundQuotaFunc(t *testing.T) {
	cfg := config.New()
	cfg.Set("Requests", "1500m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int { return int(math.Floor(f)) }

	mockLog := requireLoggerInfo(t, 1.5, 1, 1.5, 2)
	SetWithConfig(cfg,
		WithRoundQuotaFunc(roundFloor),
		WithLogger(mockLog),
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

func requireLoggerInfo(t testing.TB,
	cpuRequests float64,
	minProcs int64,
	multiplier float64,
	required int64,
) logger.Logger {
	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", cpuRequests),
		logger.NewFloatField("multiplier", multiplier),
		logger.NewIntField("minProcs", minProcs),
		logger.NewIntField("result", required),
		logger.NewIntField("GOMAXPROCS", required),
	).Times(1)
	return mockLog
}
