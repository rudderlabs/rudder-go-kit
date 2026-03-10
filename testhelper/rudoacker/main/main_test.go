package main

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	cli "github.com/urfave/cli/v3"
)

func TestAckerCommandRoutes(t *testing.T) {
	testCases := []struct {
		name         string
		args         []string
		expectedMode ackerMode
		validateCfg  func(t *testing.T, cfg runConfig)
	}{
		{
			name:         "All",
			args:         []string{"acker", "all", "--namespace", "ns-all", "--gateway-node-pattern", "gateway-*", "--srcrouter-nodes", "srcrouter-0,srcrouter-1", "--etcd-endpoints", "e1:2379,e2:2379"},
			expectedMode: modeAll,
			validateCfg: func(t *testing.T, cfg runConfig) {
				t.Helper()
				require.Equal(t, "ns-all", cfg.namespace)
				require.Equal(t, "gateway-*", cfg.gatewayNodePattern)
				require.Equal(t, []string{"srcrouter-0", "srcrouter-1"}, cfg.srcrouterNodes)
				require.Equal(t, []string{"e1:2379", "e2:2379"}, cfg.etcdEndpoints)
			},
		},
		{
			name:         "Migration",
			args:         []string{"acker", "migration", "--namespace", "ns-mig"},
			expectedMode: modeMigration,
			validateCfg: func(t *testing.T, cfg runConfig) {
				t.Helper()
				require.Equal(t, "ns-mig", cfg.namespace)
				require.Equal(t, []string{"localhost:2379"}, cfg.etcdEndpoints)
				require.Equal(t, 100*time.Millisecond, cfg.minDelay)
				require.Equal(t, 1*time.Second, cfg.maxDelay)
				require.Equal(t, 5*time.Second, cfg.dialTimeout)
			},
		},
		{
			name:         "Gateway",
			args:         []string{"acker", "gateway", "--namespace", "ns-gw", "--gateway-node-pattern", "gw-*"},
			expectedMode: modeGateway,
			validateCfg: func(t *testing.T, cfg runConfig) {
				t.Helper()
				require.Equal(t, "ns-gw", cfg.namespace)
				require.Equal(t, "gw-*", cfg.gatewayNodePattern)
			},
		},
		{
			name:         "Srcrouter",
			args:         []string{"acker", "srcrouter", "--namespace", "ns-src", "--srcrouter-nodes", "a,b,c"},
			expectedMode: modeSrcrouter,
			validateCfg: func(t *testing.T, cfg runConfig) {
				t.Helper()
				require.Equal(t, "ns-src", cfg.namespace)
				require.Equal(t, []string{"a", "b", "c"}, cfg.srcrouterNodes)
			},
		},
		{
			name:         "Job",
			args:         []string{"acker", "job", "--namespace", "ns-job"},
			expectedMode: modeJob,
			validateCfg: func(t *testing.T, cfg runConfig) {
				t.Helper()
				require.Equal(t, "ns-job", cfg.namespace)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var gotMode ackerMode
			var gotCfg runConfig
			cmd := newTestCommand(func(_ context.Context, mode ackerMode, cfg runConfig) error {
				gotMode = mode
				gotCfg = cfg
				return nil
			})

			err := cmd.Run(context.Background(), tc.args)
			require.NoError(t, err)
			require.Equal(t, tc.expectedMode, gotMode)
			tc.validateCfg(t, gotCfg)
		})
	}
}

func TestAckerCommandErrors(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "RootWithoutSubcommand",
			args: []string{"acker"},
		},
		{
			name: "MissingNamespace",
			args: []string{"acker", "job"},
		},
		{
			name: "UnknownSubcommand",
			args: []string{"acker", "unknown"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newTestCommand(func(_ context.Context, _ ackerMode, _ runConfig) error {
				t.Fatalf("run should not be called for %s", tc.name)
				return nil
			})
			err := cmd.Run(context.Background(), tc.args)
			require.Error(t, err)
		})
	}
}

func newTestCommand(run runFunc) *cli.Command {
	return newCommand(run, &bytes.Buffer{}, &bytes.Buffer{})
}

func TestSplitCSV(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{"a, b, c", []string{"a", "b", "c"}},
		{" a , b ", []string{"a", "b"}},
		{"single", []string{"single"}},
		{"", nil},
		{",", nil},
		{" , , ", nil},
	}
	for _, tc := range cases {
		got := splitCSV(tc.input)
		require.Equal(t, tc.expected, got, "splitCSV(%q)", tc.input)
	}
}
