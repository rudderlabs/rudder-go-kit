package diskutils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/rudderlabs/rudder-go-kit/config"
)

// CreateTMPDIR creates tmp dir at path configured via RUDDER_TMPDIR env var
func CreateTMPDIR() (string, error) {
	tmpdirPath := strings.TrimSuffix(config.GetString("RUDDER_TMPDIR", ""), "/")
	// second chance: fallback to /tmp if this folder exists
	if tmpdirPath == "" {
		fallbackPath := "/tmp"
		_, err := os.Stat(fallbackPath)
		if err == nil {
			tmpdirPath = fallbackPath
		}
	}
	if tmpdirPath == "" {
		return os.UserHomeDir()
	}
	return tmpdirPath, nil
}

func GetBadgerDBUsage(dir string) (int64, int64, int64, error) {
	// Notes
	// Instead of using BadgerDB's internal function to get the disk usage, we are writing our own implementation because of the following reasons:
	// 1. BadgerDB internally creates a sparse memory backed file to store the data
	// 2. The size returned by the filepath.Walk used internally gives a misleading size because the file is mostly empty and doesn't consume any disk space
	lsmSize, err := DiskUsage(dir, ".sst")
	if err != nil {
		return 0, 0, 0, err
	}
	vlogSize, err := DiskUsage(dir, ".vlog")
	if err != nil {
		return 0, 0, 0, err
	}
	totSize, err := DiskUsage(dir)
	if err != nil {
		return 0, 0, 0, err
	}
	return lsmSize, vlogSize, totSize, nil
}

// DiskUsage calculates the path's disk usage recursively in bytes. If exts are provided, only files with matching extensions will be included in the result.
func DiskUsage(path string, ext ...string) (int64, error) {
	var totSize int64
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		size, _ := GetDiskUsageOfFile(path)
		if len(ext) == 0 {
			totSize += size
		} else {
			for _, e := range ext {
				if filepath.Ext(path) == e {
					totSize += size
				}
			}
		}
		return nil
	})
	return totSize, err
}

func GetDiskUsageOfFile(path string) (int64, error) {
	// Notes
	// 1. stat.Blocks is the number of stat.Blksize blocks allocated to the file
	// 2. stat.Blksize is the filesystem block size for this filesystem
	// 3. We compute the actual disk usage of a (sparse) file by multiplying the number of blocks allocated to the file with the block size. This computes a different value than the one returned by stat.Size particularly for sparse files.
	var stat syscall.Stat_t
	err := syscall.Stat(path, &stat)
	if err != nil {
		return 0, fmt.Errorf("unable to get file size %w", err)
	}
	return int64(stat.Blksize) * stat.Blocks / 8, nil //nolint:unconvert // In amd64 architecture stat.Blksize is int64 whereas in arm64 it is int32
}

// FolderExists Check if folder exists at particular path
func FolderExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}

// FileExists Check if file exists at particular path
func FileExists(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return !fileInfo.IsDir(), nil
}
