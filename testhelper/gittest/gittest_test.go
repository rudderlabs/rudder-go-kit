package gittest_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
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
		out, err := execCmd("git", "clone", url, tempDir)
		require.NoErrorf(t, err, "should be able to clone the repository: %s", out)
		require.FileExists(t, tempDir+"/README.md", "README.md should exist in the cloned repository")

		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("Hello, World!"), 0o644), "should be able to write to file1.txt")
		out, err = execCmd("git", "-C", tempDir, "add", "file1.txt")
		require.NoErrorf(t, err, "should be able to add file1.txt: %s", out)

		out, err = execCmd("git", "-C", tempDir, "commit", "-m", "add file1.txt")
		require.NoErrorf(t, err, "should be able to commit file1.txt: %s", out)

		out, err = execCmd("git", "-C", tempDir, "push", "origin", "main")
		require.NoErrorf(t, err, "should be able to push the main branch: %s", out)

		out, err = execCmd("git", "-C", tempDir, "checkout", "-b", "develop")
		require.NoErrorf(t, err, "should be able to create a develop repository: %s", out)

		out, err = execCmd("git", "-C", tempDir, "push", "origin", "develop:develop")
		require.NoErrorf(t, err, "should be able to push the develop branch: %s", out)

		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("Hello, World!"), 0o644), "should be able to write to file2.txt")
		out, err = execCmd("git", "-C", tempDir, "add", "file2.txt")
		require.NoErrorf(t, err, "should be able to add file2.txt: %s", out)

		out, err = execCmd("git", "-C", tempDir, "commit", "-m", "add file2.txt")
		require.NoErrorf(t, err, "should be able to commit file2.txt: %s", out)

		out, err = execCmd("git", "-C", tempDir, "push", "origin", "develop")
		require.NoErrorf(t, err, "should be able to push the develop branch: %s", out)

		out, err = execCmd("git", "-C", tempDir, "tag", "-a", "v1.0.0", "-m", "v1.0.0")
		require.NoErrorf(t, err, "should be able to create a tag: %s", out)

		out, err = execCmd("git", "-C", tempDir, "push", "origin", "v1.0.0")
		require.NoErrorf(t, err, "should be able to push the tag: %s", out)
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

func execCmd(name string, args ...string) (string, error) { // nolint: unparam
	var buf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}
