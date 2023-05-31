package diskutils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rudderlabs/rudder-go-kit/config"
	"github.com/rudderlabs/rudder-go-kit/logger"
	"github.com/stretchr/testify/require"
)

func initMisc() {
	config.Reset()
	logger.Reset()
}

func Test_CreteTempDir(t *testing.T) {
	initMisc()
	tmpDirPath, err := CreateTMPDIR()
	require.NoError(t, err)
	folderExists, err := FolderExists(tmpDirPath)
	require.NoError(t, err)
	require.True(t, folderExists)
}

func TestGetDiskUsage(t *testing.T) {
	initMisc()
	// Create a temp file
	tmpDirPath := t.TempDir()
	tempFilePath := filepath.Join(tmpDirPath, "tempFileForDiskUsage")
	f, err := os.OpenFile(tempFilePath, os.O_CREATE|os.O_RDWR, 0o755) // skipcq: GSC-G302
	require.NoError(t, err)

	defer func() { _ = f.Close() }()
	defer func() { _ = os.Remove(tempFilePath) }()

	err = f.Truncate(1024 * 1024)
	require.NoError(t, err)

	fileSize, err := os.Stat(tempFilePath)
	require.NoError(t, err)
	fileDiskUsage, err := GetDiskUsageOfFile(tempFilePath)
	require.NoError(t, err)
	require.EqualValues(t, 1024*1024, fileSize.Size())
	require.EqualValues(t, 0, fileDiskUsage)

	// write some bytes into the file
	_, err = f.WriteString("test")
	require.NoError(t, err)
	fileSize, err = os.Stat(tempFilePath)
	require.NoError(t, err)
	fileDiskUsage, err = GetDiskUsageOfFile(tempFilePath)
	require.NoError(t, err)

	require.EqualValues(t, 1024*1024, fileSize.Size())
	require.Greater(t, fileDiskUsage, int64(0))
}
