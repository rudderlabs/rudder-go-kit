// Command rudoacker runs orchestrator ackers (migration, gateway, src-router, job)
// against an etcd cluster, simulating external service acknowledgments.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/samber/lo"
	cli "github.com/urfave/cli/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rudderlabs/rudder-go-kit/testhelper/rudoacker"
)

type ackerMode string

const (
	modeAll       ackerMode = "all"       // start all ackers (migration, gateway, srcrouter, job)
	modeMigration ackerMode = "migration" // start only migration acker
	modeGateway   ackerMode = "gateway"   // start only gateway acker
	modeSrcrouter ackerMode = "srcrouter" // start only srcrouter acker
	modeJob       ackerMode = "job"       // start only job acker
)

type runConfig struct {
	etcdEndpoints      []string
	namespace          string
	gatewayNodePattern string
	srcrouterNodes     []string
	minDelay           time.Duration
	maxDelay           time.Duration
	dialTimeout        time.Duration
}

type runFunc func(context.Context, ackerMode, runConfig) error

func main() {
	cmd := newCommand(runAckerMode, os.Stdout, os.Stderr)
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func newCommand(run runFunc, stdout, stderr io.Writer) *cli.Command {
	return &cli.Command{
		Name:      "rudoacker",
		Usage:     "Run orchestrator ackers against etcd",
		UsageText: "rudoacker <subcommand> [flags]",
		Writer:    stdout,
		ErrWriter: stderr,
		// Enables urfave/cli dynamic candidate generation via --generate-shell-completion.
		EnableShellCompletion:      true,
		ShellCompletionCommandName: "__internal-completion",
		ExitErrHandler: func(_ context.Context, _ *cli.Command, _ error) {
			// Let Run return errors instead of exiting immediately.
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "etcd-endpoints",
				Usage: "comma-separated list of etcd endpoints",
				Value: "localhost:2379",
			},
			&cli.DurationFlag{
				Name:  "min-delay",
				Usage: "minimum ack delay",
				Value: 100 * time.Millisecond,
			},
			&cli.DurationFlag{
				Name:  "max-delay",
				Usage: "maximum ack delay",
				Value: 1 * time.Second,
			},
			&cli.DurationFlag{
				Name:  "dial-timeout",
				Usage: "etcd dial timeout",
				Value: 5 * time.Second,
			},
		},
		Commands: []*cli.Command{
			{
				Name:  string(modeAll),
				Usage: "start all ackers (migration, gateway, srcrouter, job)",
				Flags: []cli.Flag{
					requiredNamespaceFlag(),
					gatewayNodePatternFlag(),
					srcrouterNodesFlag(),
				},
				Action: actionForMode(modeAll, run),
			},
			{
				Name:   string(modeMigration),
				Usage:  "start only migration acker",
				Flags:  []cli.Flag{requiredNamespaceFlag()},
				Action: actionForMode(modeMigration, run),
			},
			{
				Name:   string(modeGateway),
				Usage:  "start only gateway acker",
				Flags:  []cli.Flag{requiredNamespaceFlag(), gatewayNodePatternFlag()},
				Action: actionForMode(modeGateway, run),
			},
			{
				Name:   string(modeSrcrouter),
				Usage:  "start only srcrouter acker",
				Flags:  []cli.Flag{requiredNamespaceFlag(), srcrouterNodesFlag()},
				Action: actionForMode(modeSrcrouter, run),
			},
			{
				Name:   string(modeJob),
				Usage:  "start only job acker",
				Flags:  []cli.Flag{requiredNamespaceFlag()},
				Action: actionForMode(modeJob, run),
			},
		},
		Action: func(_ context.Context, cmd *cli.Command) error {
			_ = cli.ShowRootCommandHelp(cmd)
			return cli.Exit("subcommand is required", 1)
		},
	}
}

func actionForMode(mode ackerMode, run runFunc) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		cfg := readRunConfig(cmd)
		return runWithSignals(ctx, mode, cfg, run)
	}
}

func requiredNamespaceFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:     "namespace",
		Usage:    "etcd key namespace (release name)",
		Required: true,
	}
}

func gatewayNodePatternFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "gateway-node-pattern",
		Usage: "gateway node name pattern (e.g. gateway-*)",
		Value: "gateway-*",
	}
}

func srcrouterNodesFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:  "srcrouter-nodes",
		Usage: "comma-separated list of src-router node names",
		Value: "srcrouter-0",
	}
}

func runWithSignals(ctx context.Context, mode ackerMode, cfg runConfig, run runFunc) error {
	signalCtx, cancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return run(signalCtx, mode, cfg)
}

func readRunConfig(cmd *cli.Command) runConfig {
	return runConfig{
		etcdEndpoints:      splitCSV(cmd.String("etcd-endpoints")),
		namespace:          strings.TrimSpace(cmd.String("namespace")),
		gatewayNodePattern: strings.TrimSpace(cmd.String("gateway-node-pattern")),
		srcrouterNodes:     splitCSV(cmd.String("srcrouter-nodes")),
		minDelay:           cmd.Duration("min-delay"),
		maxDelay:           cmd.Duration("max-delay"),
		dialTimeout:        cmd.Duration("dial-timeout"),
	}
}

func runAckerMode(ctx context.Context, mode ackerMode, cfg runConfig) error {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.etcdEndpoints,
		DialTimeout: cfg.dialTimeout,
	})
	if err != nil {
		return fmt.Errorf("connecting to etcd: %w", err)
	}
	defer func() { _ = etcdClient.Close() }()

	logListener := func(ackKey string) {
		log.Printf("acked: %s", ackKey)
	}

	g, runCtx := errgroup.WithContext(ctx)

	startMigration := func() error {
		if err := rudoacker.NewMigrationAcker(runCtx, g, etcdClient, cfg.namespace).
			WithMinDelay(cfg.minDelay).WithMaxDelay(cfg.maxDelay).WithAckListener(logListener).
			Start(); err != nil {
			return fmt.Errorf("starting migration acker: %w", err)
		}
		log.Printf("migration acker started for namespace %s", cfg.namespace)
		return nil
	}
	startGateway := func() error {
		if err := rudoacker.NewGatewayAcker(runCtx, g, etcdClient, cfg.namespace, cfg.gatewayNodePattern).
			WithMinDelay(cfg.minDelay).WithMaxDelay(cfg.maxDelay).WithAckListener(logListener).
			Start(); err != nil {
			return fmt.Errorf("starting gateway acker: %w", err)
		}
		log.Printf("gateway acker started for namespace %s", cfg.namespace)
		return nil
	}
	startSrcrouter := func() error {
		if err := rudoacker.NewSrcrouterAcker(runCtx, g, etcdClient, cfg.namespace, cfg.srcrouterNodes).
			WithMinDelay(cfg.minDelay).WithMaxDelay(cfg.maxDelay).WithAckListener(logListener).
			Start(); err != nil {
			return fmt.Errorf("starting srcrouter acker: %w", err)
		}
		log.Printf("srcrouter acker started for namespace %s", cfg.namespace)
		return nil
	}
	startJob := func() error {
		if err := rudoacker.NewJobAcker(runCtx, g, etcdClient, cfg.namespace).
			WithMinDelay(cfg.minDelay).WithMaxDelay(cfg.maxDelay).WithAckListener(logListener).
			Start(); err != nil {
			return fmt.Errorf("starting job acker: %w", err)
		}
		log.Printf("job acker started for namespace %s", cfg.namespace)
		return nil
	}

	ackersByMode := map[ackerMode][]func() error{
		modeAll:       {startMigration, startGateway, startSrcrouter, startJob},
		modeMigration: {startMigration},
		modeGateway:   {startGateway},
		modeSrcrouter: {startSrcrouter},
		modeJob:       {startJob},
	}
	fns, ok := ackersByMode[mode]
	if !ok {
		return fmt.Errorf("unsupported mode %q", mode)
	}
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}

	log.Printf("ackers running (mode=%s, namespace=%s, endpoints=%s)", mode, cfg.namespace, strings.Join(cfg.etcdEndpoints, ","))
	if err := g.Wait(); err != nil {
		return fmt.Errorf("acker error: %w", err)
	}
	return nil
}

func splitCSV(s string) []string {
	cleaned := lo.FilterMap(strings.Split(s, ","), func(part string, _ int) (string, bool) {
		part = strings.TrimSpace(part)
		return part, part != ""
	})
	if len(cleaned) == 0 {
		return nil
	}
	return cleaned
}
