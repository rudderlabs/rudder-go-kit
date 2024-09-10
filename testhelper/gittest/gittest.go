// Package gittest provides a test helper for creating a git server that serves a git repository.
package gittest

import (
	"bytes"
	"encoding/pem"
	"fmt"
	"net"
	"net/http/cgi"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/testhelper/httptest"
)

type Server struct {
	*httptest.Server
	URL      string
	rootPath string
}

// NewHttpServer creates a new httptest.Server that serves a git repository from the given sourcePath.
func NewHttpServer(t testing.TB, sourcePath string) *Server {
	return newServer(t, sourcePath, false)
}

// NewHttpsServer creates a new httptest.Server that serves a git repository from the given sourcePath.
func NewHttpsServer(t testing.TB, sourcePath string) *Server {
	return newServer(t, sourcePath, true)
}

// newServer creates a new httptest.Server that serves a git repository from the given sourcePath.
func newServer(t testing.TB, sourcePath string, secure bool) *Server {
	t.Helper()
	tempDir := t.TempDir()
	org := "org"
	repo := filepath.Base(sourcePath)
	if !strings.HasSuffix(repo, ".git") {
		repo = repo + ".git"
	}
	source := sourcePath
	if !strings.HasSuffix(sourcePath, "/") {
		source = source + "/"
	}

	workingDir := filepath.Join(tempDir, "workdir")
	require.NoErrorf(t, os.MkdirAll(workingDir, os.ModePerm), "should be able to create %s", workingDir)
	gitRoot := filepath.Join(tempDir, org, repo)
	require.NoErrorf(t, os.MkdirAll(gitRoot, os.ModePerm), "should be able to create %s", gitRoot)

	out, err := execCmd("rsync", "--recursive", source, workingDir)
	require.NoErrorf(t, err, "should be able to copy %s to %s: %s", source, workingDir, out)

	gitPath, err := exec.LookPath("git")
	require.NoError(t, err, "should be able to find git in PATH")
	out, err = execCmd("git", "init", workingDir)
	require.NoErrorf(t, err, "should be able to initialize git repository: %s", out)
	out, err = execCmd("git", "-C", workingDir, "branch", "-m", "main")
	require.NoErrorf(t, err, "should be able to rename the default branch to main: %s", out)
	out, err = execCmd("git", "-C", workingDir, "add", ".")
	require.NoErrorf(t, err, "should be able to add files to git repository: %s", out)
	commitCmd := exec.Command("git", "-C", workingDir, "commit", "-m", "initial commit")
	commitCmd.Env = append(commitCmd.Env, "GIT_AUTHOR_NAME=git test", "GIT_COMMITTER_NAME=git test", "GIT_AUTHOR_EMAIL=gittest@example.com", "GIT_COMMITTER_EMAIL=gittest@example.com")
	require.NoError(t, commitCmd.Run(), "should be able to commit files to git repository")

	out, err = execCmd("git", "clone", "--bare", workingDir, gitRoot)
	require.NoErrorf(t, err, "should be able to clone the bare repository: %s", out)

	handler := &cgi.Handler{
		Path: gitPath,
		Args: []string{"http-backend"},
		Env: []string{
			fmt.Sprintf("GIT_PROJECT_ROOT=%s", tempDir),
			"GIT_HTTP_EXPORT_ALL=true",
			"REMOTE_USER=git",
		},
	}

	localIP := getLocalIP(t)
	var s *httptest.Server
	if !secure {
		s = httptest.NewServer(handler)
	} else {
		s = httptest.NewTLSServer(localIP, handler)
		certPath := filepath.Join(tempDir, "server.crt")
		require.NoError(t, writeServerCA(s, certPath))
		t.Setenv("SSL_CERT_FILE", certPath)

	}
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(serverURL.Host)
	require.NoError(t, err)
	serverURL.Host = net.JoinHostPort(getLocalIP(t), port)
	url := serverURL.String() + "/" + org + "/" + repo
	return &Server{
		Server:   s,
		URL:      url,
		rootPath: gitRoot,
	}
}

// getServerCA returns a byte slice containing the PEM encoding of the server's CA certificate
func (s *Server) GetServerCA() []byte {
	return getServerCA(s.Server)
}

func (s *Server) GetLatestCommitHash(t testing.TB) string {
	commitHashCmd := exec.Command("git", "-C", s.rootPath, "rev-parse", "HEAD")
	commitHash, err := commitHashCmd.Output()
	require.NoError(t, err, "should be able to get the latest commit hash")
	return strings.TrimSpace(string(commitHash))
}

func getServerCA(server *httptest.Server) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.TLS.Certificates[0].Certificate[0]})
}

// writeServerCA writes the PEM-encoded server certificate to a given path
func writeServerCA(server *httptest.Server, path string) error {
	certOut, err := os.Create(path)
	if err != nil {
		return err
	}
	defer certOut.Close()

	if _, err := certOut.Write(getServerCA(server)); err != nil {
		return err
	}

	return nil
}

func getLocalIP(t testing.TB) string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, conn.Close())
	}()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func execCmd(name string, args ...string) (string, error) {
	var buf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}
