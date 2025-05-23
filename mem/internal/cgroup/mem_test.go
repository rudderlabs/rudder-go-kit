package cgroup_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/bytesize"
	"github.com/rudderlabs/rudder-go-kit/mem/internal/cgroup"
)

func TestCgroupMemory(t *testing.T) {
	t.Run("cgroups v1 with limit", func(t *testing.T) {
		basePath := "testdata/cgroups_v1_mem_limit"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, 25*bytesize.GB, limit, "when a limit is set, this limit should be returned")
		require.EqualValues(t, 7873486848, cgroup.GetMemoryUsage(basePath))
		require.EqualValues(t, 7873486848, cgroup.GetMemoryRSS(basePath))
	})

	t.Run("cgroups v1 with self limit", func(t *testing.T) {
		basePath := "testdata/cgroups_v1_mem_limit_proc_self"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, 25*bytesize.GB, limit, "when a limit is set, this limit should be returned")
		require.EqualValues(t, 9456156572, cgroup.GetMemoryUsage(basePath))
		require.EqualValues(t, 9456156572, cgroup.GetMemoryRSS(basePath))
	})

	t.Run("cgroups v1 with hierarchical limit", func(t *testing.T) {
		basePath := "testdata/cgroups_v1_mem_hierarchy"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, 25*bytesize.GB, limit, "when a hierarchical limit is set, this limit should be returned")
		require.EqualValues(t, 7873486848, cgroup.GetMemoryUsage(basePath))
		require.EqualValues(t, 7873486848, cgroup.GetMemoryRSS(basePath))
	})

	t.Run("cgroups v1 no limit", func(t *testing.T) {
		basePath := "testdata/cgroups_v1_mem_no_limit"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, totalMem, limit, "when no limit is set, total memory should be returned")
		require.EqualValues(t, 7873486848, cgroup.GetMemoryUsage(basePath))
		require.EqualValues(t, 7873486848, cgroup.GetMemoryRSS(basePath))
	})

	t.Run("cgroups v2 with limit", func(t *testing.T) {
		basePath := "testdata/cgroups_v2_mem_limit"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, 32*bytesize.GB, limit, "when a limit is set, this limit should be returned")
		require.EqualValues(t, 26071040, cgroup.GetMemoryUsage(basePath))
		require.EqualValues(t, 3145728, cgroup.GetMemoryRSS(basePath))
	})

	t.Run("cgroups v2 no limit", func(t *testing.T) {
		basePath := "testdata/cgroups_v2_mem_no_limit"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, totalMem, limit, "when no limit is set, total memory should be returned")
		require.EqualValues(t, 26071040, cgroup.GetMemoryUsage(basePath))
		require.EqualValues(t, 3145728, cgroup.GetMemoryRSS(basePath))
	})

	t.Run("no cgroups info", func(t *testing.T) {
		basePath := "testdata/invalid_path"
		totalMem := int(100 * bytesize.GB)
		limit := cgroup.GetMemoryLimit(basePath, totalMem)

		require.EqualValues(t, limit, limit, "when no cgroups info is available, this limit should be returned")
		require.EqualValues(t, 0, cgroup.GetMemoryUsage(basePath))
	})
}
