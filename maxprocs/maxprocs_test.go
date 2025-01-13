package maxprocs_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/rudderlabs/rudder-go-kit/logger/mock_logger"
	"github.com/rudderlabs/rudder-go-kit/maxprocs"
)

func TestSet_Default(t *testing.T) {
	before := runtime.GOMAXPROCS(0)  // Capture original value
	defer runtime.GOMAXPROCS(before) // Restore after test

	mockLog := requireLoggerInfo(t, 1.1, 1, 3, 4)
	maxprocs.Set("1100m", maxprocs.WithLogger(mockLog))
	require.Equal(t, 4, runtime.GOMAXPROCS(0)) // 1100m * 3 = 3.3 → ceil = 4
}

func TestSetWithConfig_Default(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "1100m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 1.1, 1, 3, 4)
	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 4, runtime.GOMAXPROCS(0)) // 1100m * 3 = 3.3 → ceil = 4
}

func TestSet_WithInvalidCPURequest_Invalid1(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 3),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 3),
		logger.NewIntField("GOMAXPROCS", 3),
	).Times(1)

	maxprocs.Set("invalid", maxprocs.WithLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Defaults to 1 * 3 → ceil = 3
}

func TestSetWithConfig_WithInvalidCPURequest_Invalid1(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "invalid")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 3),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 3),
		logger.NewIntField("GOMAXPROCS", 3),
	).Times(1)

	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Defaults to 1 * 3 → ceil = 3
}

func TestSet_WithInvalidCPURequest_Invalid2(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 3),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 3),
		logger.NewIntField("GOMAXPROCS", 3),
	).Times(1)

	maxprocs.Set("invalid_m", maxprocs.WithLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Defaults to 1 * 3 → ceil = 3
}

func TestSetWithConfig_WithInvalidCPURequest_Invalid2(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "invalid_m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)
	mockLog.EXPECT().Infon("GOMAXPROCS has been configured",
		logger.NewFloatField("cpuRequests", 1),
		logger.NewFloatField("multiplier", 3),
		logger.NewIntField("minProcs", 1),
		logger.NewIntField("result", 3),
		logger.NewIntField("GOMAXPROCS", 3),
	).Times(1)

	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Defaults to 1 * 3 → ceil = 3
}

func TestSet_WithMinProcs(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 5, 3, 5)
	maxprocs.Set("100m",
		maxprocs.WithMinProcs(5),
		maxprocs.WithLogger(mockLog),
	)

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSetWithConfig_WithMinProcs(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "100m")
	cfg.Set("MaxProcs.MinProcs", 5)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.1, 5, 3, 5)
	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSet_WithMultiplier(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.3, 1, 4, 2)
	maxprocs.Set("300m",
		maxprocs.WithCPURequestsMultiplier(4),
		maxprocs.WithLogger(mockLog),
	)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSetWithConfig_WithMultiplier(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "300m")
	cfg.Set("MaxProcs.CPURequestsMultiplier", 4)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	mockLog := requireLoggerInfo(t, 0.3, 1, 4, 2)
	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSet_CustomRoundQuotaFunc(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int { return int(math.Floor(f)) }

	mockLog := requireLoggerInfo(t, 1.5, 1, 3, 4)
	maxprocs.Set("1500m",
		maxprocs.WithRoundQuotaFunc(roundFloor),
		maxprocs.WithLogger(mockLog),
	)

	require.Equal(t, 4, runtime.GOMAXPROCS(0)) // 1500m * 3 = 4.5 → floor = 4
}

func TestSetWithConfig_CustomRoundQuotaFunc(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "1500m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int { return int(math.Floor(f)) }

	mockLog := requireLoggerInfo(t, 1.5, 1, 3, 4)
	maxprocs.SetWithConfig(cfg,
		maxprocs.WithRoundQuotaFunc(roundFloor),
		maxprocs.WithLogger(mockLog),
	)

	require.Equal(t, 4, runtime.GOMAXPROCS(0)) // 1500m * 3 = 4.5 → floor = 4
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
