package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_MinsOfDay(t *testing.T) {
	require.Equal(t, MinsOfDay("00:00"), 0)
	require.Equal(t, MinsOfDay("00:01"), 1)
	require.Equal(t, MinsOfDay("00:59"), 59)
	require.Equal(t, MinsOfDay("01:00"), 60)
	require.Equal(t, MinsOfDay("23:59"), 1439)
	require.Equal(t, MinsOfDay("26:00"), 0)
}

func Test_StartOfDay(t *testing.T) {
	require.Equal(t, StartOfDay(time.Date(2023, 11, 4, 5, 23, 43, 3, time.Local)), time.Date(2023, time.November, 4, 0, 0, 0, 0, time.Local))
	require.Equal(t, StartOfDay(time.Date(2023, 12, 4, 5, 23, 43, 3, time.UTC)), time.Date(2023, time.December, 4, 0, 0, 0, 0, time.UTC))
	require.Equal(t, StartOfDay(time.Date(2023, 15, 4, 5, 23, 43, 3, time.Local)), time.Date(2024, time.March, 4, 0, 0, 0, 0, time.Local))
}

func Test_Now(t *testing.T) {
	require.InDelta(t, Now().Unix(), time.Now().UTC().Unix(), 1000)
}

func Test_GetElapsedMinsInThisDay(t *testing.T) {
	require.Equal(t, GetElapsedMinsInThisDay(time.Date(2023, 11, 4, 5, 23, 43, 3, time.Local)), 323)
	require.Equal(t, GetElapsedMinsInThisDay(time.Date(2023, 12, 4, 5, 25, 43, 3, time.UTC)), 325)
	// Giving an invalid month as input also does return the correct value
	require.Equal(t, GetElapsedMinsInThisDay(time.Date(2023, 15, 4, 5, 33, 43, 3, time.Local)), 333)
	// Providing an incorrect time does return 0
	require.Equal(t, GetElapsedMinsInThisDay(time.Date(2023, 15, 4, 5, 63, 43, 3, time.Local)), 0)
}
