package gittest_test

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/gittest"
)

func TestGitServer(t *testing.T) {
	t.Run("http", func(t *testing.T) {
		s := gittest.NewHttpServer(t, "testdata/gitrepo")
		defer s.Close()
		tempDir := t.TempDir()
		url := s.URL
		require.NoError(t, exec.Command("git", "-c", "http.sslVerify=false", "clone", url, tempDir).Run(), "should be able to clone the repository")
		require.FileExists(t, tempDir+"/README.md", "README.md should exist in the cloned repository")
	})

	t.Run("https", func(t *testing.T) {
		s := gittest.NewHttpsServer(t, "testdata/gitrepo")
		defer s.Close()
		tempDir := t.TempDir()
		url := s.URL
		require.NoError(t, exec.Command("git", "-c", "http.sslVerify=false", "clone", url, tempDir).Run(), "should be able to clone the repository")
		require.FileExists(t, tempDir+"/README.md", "README.md should exist in the cloned repository")
	})
}
