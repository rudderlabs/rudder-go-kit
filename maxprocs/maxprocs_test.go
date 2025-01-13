package maxprocs_test

import (
	"math"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger/mock_logger"
	"github.com/rudderlabs/rudder-go-kit/maxprocs"
)

func TestSet_Default(t *testing.T) {
	before := runtime.GOMAXPROCS(0)  // Capture original value
	defer runtime.GOMAXPROCS(before) // Restore after test

	maxprocs.Set("500m")
	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 500m * 3 = 1.5 → ceil = 2
}

func TestSetWithConfig_Default(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "500m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	maxprocs.SetWithConfig(cfg)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 500m * 3 = 1.5 → ceil = 2
}

func TestSet_WithInvalidCPURequest_Invalid1(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with ParseFloat, using default value").Times(1)

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

	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Defaults to 1 * 3 → ceil = 3
}

func TestSet_WithInvalidCPURequest_Invalid2(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	ctrl := gomock.NewController(t)
	mockLog := mock_logger.NewMockLogger(ctrl)
	mockLog.EXPECT().Warnn("unable to parse CPU requests with Atoi, using default value").Times(1)

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

	maxprocs.SetWithConfig(cfg, maxprocs.WithLogger(mockLog))

	require.Equal(t, 3, runtime.GOMAXPROCS(0)) // Defaults to 1 * 3 → ceil = 3
}

func TestSet_WithMinProcs(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	maxprocs.Set("100m", maxprocs.WithMinProcs(5))

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSetWithConfig_WithMinProcs(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "100m")
	cfg.Set("MaxProcs.MinProcs", 5)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	maxprocs.SetWithConfig(cfg)

	require.Equal(t, 5, runtime.GOMAXPROCS(0)) // MinProcs overrides calculated value
}

func TestSet_WithMultiplier(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	maxprocs.Set("300m", maxprocs.WithCPURequestsMultiplier(4))

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSetWithConfig_WithMultiplier(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "300m")
	cfg.Set("MaxProcs.CPURequestsMultiplier", 4)

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	maxprocs.SetWithConfig(cfg)

	require.Equal(t, 2, runtime.GOMAXPROCS(0)) // 300m * 4 = 1.2 → ceil = 2
}

func TestSet_CustomRoundQuotaFunc(t *testing.T) {
	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int {
		return int(math.Floor(f))
	}

	maxprocs.Set("1500m", maxprocs.WithRoundQuotaFunc(roundFloor))

	require.Equal(t, 4, runtime.GOMAXPROCS(0)) // 1500m * 3 = 4.5 → floor = 4
}

func TestSetWithConfig_CustomRoundQuotaFunc(t *testing.T) {
	cfg := config.New()
	cfg.Set("MaxProcs.CPURequests", "1500m")

	before := runtime.GOMAXPROCS(0)
	defer runtime.GOMAXPROCS(before)

	roundFloor := func(f float64) int {
		return int(math.Floor(f))
	}

	maxprocs.SetWithConfig(cfg, maxprocs.WithRoundQuotaFunc(roundFloor))

	require.Equal(t, 4, runtime.GOMAXPROCS(0)) // 1500m * 3 = 4.5 → floor = 4
}
