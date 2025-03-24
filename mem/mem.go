package mem

import (
	"fmt"

	gomem "github.com/shirou/gopsutil/v3/mem"

	"github.com/rudderlabs/rudder-go-kit/mem/internal/cgroup"
)

// Stat represents memory statistics (cgroup aware)
type Stat struct {
	// Total memory in bytes
	Total uint64
	// Available memory in bytes, calculated as [Total] - [Used]
	Available uint64
	// Available memory in percentage
	AvailablePercent float64
	// Used memory in bytes. Includes [total_active_file], that is cache memory that has been identified as active by the kernel.
	Used uint64
	// Used memory in percentage
	UsedPercent float64
	// Size of RSS in bytes. Like Used, but without [total_active_file].
	RSS uint64
}

// RSSPercent returns the percentage of RSS memory
func (s *Stat) RSSPercent() float64 {
	return float64(s.RSS) * 100 / float64(s.Total)
}

// AvailableIgnoreCache returns the available memory in bytes, ignoring cache memory. This is calculated as [Total] - [RSS]
func (s *Stat) AvailableIgnoreCache() uint64 {
	return s.Total - s.RSS
}

// AvailableIgnoreCachePercent returns the available memory in percentage, ignoring cache memory. This is calculated as [Total] - [RSS]
func (s *Stat) AvailableIgnoreCachePercent() float64 {
	return float64(s.AvailableIgnoreCache()) * 100 / float64(s.Total)
}

// Get current memory statistics
func Get() (*Stat, error) {
	return _default.Get()
}

var _default *collector

func init() {
	_default = &collector{}
}

type collector struct {
	basePath string
}

// Get current memory statistics
func (c *collector) Get() (*Stat, error) {
	var stat Stat
	mem, err := gomem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory statistics: %w", err)
	}

	cgroupLimit := cgroup.GetMemoryLimit(c.basePath, int(mem.Total))
	if cgroupLimit < int(mem.Total) { // if cgroup limit is set read memory statistics from cgroup
		stat.Total = uint64(cgroupLimit)
		stat.Used = uint64(cgroup.GetMemoryUsage(c.basePath))
		if stat.Used > stat.Total {
			stat.Used = stat.Total
		}
		stat.RSS = uint64(cgroup.GetMemoryRSS(c.basePath))
		if stat.RSS == 0 || stat.RSS > stat.Used {
			stat.RSS = stat.Used
		}
		stat.Available = stat.Total - stat.Used
	} else {
		stat.Total = mem.Total
		stat.Available = mem.Available
		stat.Used = stat.Total - stat.Available
		stat.RSS = stat.Used
	}
	stat.AvailablePercent = float64(stat.Available) * 100 / float64(stat.Total)
	stat.UsedPercent = float64(stat.Used) * 100 / float64(stat.Total)
	return &stat, nil
}
