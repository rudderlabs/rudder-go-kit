package config

import (
	"os"
	"strconv"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func BenchmarkNonReloadableChangeDetections(b *testing.B) {
	run := func(b *testing.B, name string, numKeys int) {
		f, err := os.CreateTemp(b.TempDir(), "*config.yaml")
		require.NoError(b, err)
		configData, err := os.ReadFile("testdata/config.yaml")
		require.NoError(b, err)
		err = os.WriteFile(f.Name(), configData, 0o644)
		require.NoError(b, err)
		f.Close()
		b.Setenv("CONFIG_PATH", f.Name())
		c := New()
		for range numKeys {
			// in every iteration we are registering 3 distinct configuration keys, each one with a different number of components (1,2,3)
			key1 := uuid.New().String()
			key2 := uuid.New().String()
			key3 := uuid.New().String()

			_ = c.GetStringVar("someValue", key1)
			_ = c.GetStringVar("someValue", key1, key2)
			_ = c.GetStringVar("someValue", key1, key2, key3)
		}
		// Run the benchmark
		b.Run(name+"_detection_with_"+strconv.Itoa(numKeys*3)+"_keys_registered", func(b *testing.B) {
			for b.Loop() {
				// trigger change detection
				c.Set("someValue", uuid.New().String())
			}
		})
	}

	for _, num := range []int{10, 100, 1000, 10000, 100000} {
		b.Setenv("CONFIG_ADVANCED_DETECTION", "false")
		run(b, "simple", num)
		b.Setenv("CONFIG_ADVANCED_DETECTION", "true")
		run(b, "advanced", num)
	}
}
