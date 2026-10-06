package clickhouse

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/distribution/reference"
	"github.com/ory/dockertest/v3/docker"
)

const DefaultImage = "clickhouse/clickhouse-server:26.3@sha256:810861a2e2d0188744f5f23b2d3ec9ff95812bcb9ddbb8fed13a377a7f305893"

// The image entrypoint uses CLICKHOUSE_DB unquoted in CREATE DATABASE and CLICKHOUSE_USER as an XML element name
// whose dots ClickHouse would read as path separators. Other names make the container fail to start.
var (
	databasePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	userPattern     = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]*$`)
)

type Opt func(*Config)

type Config struct {
	Image            string
	NetworkID        string
	BindIP           string
	PrintLogsOnError bool
	User             string
	Password         string
	Database         string
	TLS              bool
	Shards           int
	Replicas         int
	ConfigXML        string
	UsersXML         string
	Env              []string
	Memory           int64

	cluster bool
}

func WithTag(tag string) Opt {
	return func(config *Config) { config.Image = "clickhouse/clickhouse-server:" + tag }
}

func WithImage(image string) Opt {
	return func(config *Config) { config.Image = image }
}

func WithNetwork(network *docker.Network) Opt {
	return func(config *Config) {
		if network != nil {
			config.NetworkID = network.ID
		}
	}
}

func WithBindIP(bindIP string) Opt {
	return func(config *Config) { config.BindIP = bindIP }
}

func WithPrintLogsOnError(enabled bool) Opt {
	return func(config *Config) { config.PrintLogsOnError = enabled }
}

func WithUser(user string) Opt {
	return func(config *Config) { config.User = user }
}

func WithPassword(password string) Opt {
	return func(config *Config) { config.Password = password }
}

func WithDatabase(database string) Opt {
	return func(config *Config) { config.Database = database }
}

func WithTLS() Opt {
	return func(config *Config) { config.TLS = true }
}

func WithCluster(shards, replicas int) Opt {
	return func(config *Config) {
		config.Shards, config.Replicas = shards, replicas
		config.cluster = true
	}
}

func WithConfig(document string) Opt {
	return func(config *Config) { config.ConfigXML = document }
}

func WithUsersConfig(document string) Opt {
	return func(config *Config) { config.UsersXML = document }
}

func WithEnv(env ...string) Opt {
	return func(config *Config) { config.Env = append(config.Env, env...) }
}

func WithMemory(memory int64) Opt {
	return func(config *Config) { config.Memory = memory }
}

func defaultConfig() Config {
	return Config{Image: DefaultImage, User: "rudder", Password: "password", Database: "rudderdb"}
}

func (config Config) validate() error {
	if _, _, err := splitImage(config.Image); err != nil {
		return err
	}
	if !userPattern.MatchString(config.User) {
		return fmt.Errorf("user %q must match %s", config.User, userPattern)
	}
	if !databasePattern.MatchString(config.Database) {
		return fmt.Errorf("database %q must match %s", config.Database, databasePattern)
	}
	if config.TLS && config.Password == "" {
		return fmt.Errorf("ClickHouse TLS requires a non-empty password")
	}
	if config.Memory < 0 {
		return fmt.Errorf("memory must not be negative")
	}
	if config.Shards < 0 || config.Replicas < 0 || (config.Shards == 0) != (config.Replicas == 0) || (config.cluster && config.Shards == 0) {
		return fmt.Errorf("cluster shards and replicas must both be positive")
	}
	if config.Shards > 0 && config.Replicas > int(^uint(0)>>1)/config.Shards {
		return fmt.Errorf("cluster node count overflows int")
	}
	for _, document := range []string{config.ConfigXML, config.UsersXML} {
		if err := validateXML(document); err != nil {
			return fmt.Errorf("validating ClickHouse XML: %w", err)
		}
	}
	for _, env := range config.Env {
		key, _, found := strings.Cut(env, "=")
		if !found || key == "" {
			return fmt.Errorf("environment variable %q must have KEY=value form", env)
		}
		switch key {
		case "CLICKHOUSE_USER", "CLICKHOUSE_PASSWORD", "CLICKHOUSE_PASSWORD_FILE", "CLICKHOUSE_DB",
			"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT", "CLICKHOUSE_SKIP_USER_SETUP":
			return fmt.Errorf("use credential options instead of environment variable %s", key)
		}
	}
	return nil
}

func validateXML(document string) error {
	if document == "" {
		return nil
	}
	decoder := xml.NewDecoder(strings.NewReader(document))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		switch element := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if element.Name.Local != "clickhouse" {
					return fmt.Errorf("document root must be clickhouse")
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(element)) != "" {
				return fmt.Errorf("text outside document root")
			}
		}
	}
	if roots != 1 {
		return fmt.Errorf("expected one clickhouse document root")
	}
	return nil
}

func splitImage(image string) (string, string, error) {
	parsed, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return "", "", fmt.Errorf("parsing ClickHouse image: %w", err)
	}
	name, digest, pinned := strings.Cut(image, "@")
	repository, tag := name, "latest"
	if separator := strings.LastIndex(name, ":"); separator > strings.LastIndex(name, "/") {
		repository, tag = name[:separator], name[separator+1:]
	}
	if pinned {
		if _, ok := parsed.(reference.Digested); !ok {
			return "", "", fmt.Errorf("image digest is invalid")
		}
		tag += "@" + digest
	}
	return repository, tag, nil
}
