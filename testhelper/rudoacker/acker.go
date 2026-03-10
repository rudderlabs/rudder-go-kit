// Package rudoacker provides test helpers that simulate external services acknowledging
// migration pipeline stages. Each rudoacker watches an etcd prefix for requests and
// writes the corresponding acknowledgments after a random delay.
package rudoacker

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
	clientv3 "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"

	clustertypes "github.com/rudderlabs/rudder-schemas/go/cluster"

	"github.com/rudderlabs/rudder-go-kit/etcdwatcher"
	"github.com/rudderlabs/rudder-go-kit/jsonrs"
)

// WithMinDelay configures the minimum delay before acknowledging a request. Default is 100ms.
func (b *Builder[T]) WithMinDelay(d time.Duration) *Builder[T] {
	b.minDelay = d
	return b
}

// WithMaxDelay configures the maximum delay before acknowledging a request. Default is 1s.
func (b *Builder[T]) WithMaxDelay(d time.Duration) *Builder[T] {
	b.maxDelay = d
	return b
}

// WithAckListener provides a callback that is invoked with the ack key every time the acker acknowledges.
func (b *Builder[T]) WithAckListener(listener func(ackKey string)) *Builder[T] {
	b.ackListener = listener
	return b
}

// WithEventListener provides a typed callback that is invoked with the key and deserialized value
// every time the acker receives an event, before acknowledging.
func (b *Builder[T]) WithEventListener(listener func(key string, value T)) *Builder[T] {
	b.eventListener = listener
	return b
}

// NewMigrationAcker creates a builder that watches for new migration requests under the given namespace
// and acks them after a random delay to simulate processor nodes acknowledging migration requests.
func NewMigrationAcker(ctx context.Context, g *errgroup.Group,
	client *clientv3.Client, namespace string,
) *Builder[clustertypes.PartitionMigration] {
	prefix := "/" + namespace + "/migration/request/"
	return newBuilder(ctx, g, client, prefix, func(_ string, migration clustertypes.PartitionMigration) []ackEntry {
		involvedNodes := lo.Union(migration.SourceNodes(), migration.TargetNodes())
		entries := make([]ackEntry, 0, len(involvedNodes))
		for _, nodeIdx := range involvedNodes {
			nodeName := fmt.Sprintf("node-%d", nodeIdx)
			entries = append(entries, ackEntry{
				Key:   migration.AckKey(nodeName),
				Value: migration.Ack(nodeIdx, nodeName),
			})
		}
		return entries
	}, nil)
}

// NewGatewayAcker creates a builder that watches for reload gateway requests under the given namespace
// and acks them after a random delay to simulate gateway pods acknowledging reload requests.
// nodePattern (e.g. "gateway-*") is used to map node indices to pod names.
func NewGatewayAcker(ctx context.Context, g *errgroup.Group,
	client *clientv3.Client, namespace, nodePattern string,
) *Builder[clustertypes.ReloadGatewayCommand] {
	prefix := "/" + namespace + "/reload/gateway/request/"
	return newBuilder(ctx, g, client, prefix, func(_ string, cmd clustertypes.ReloadGatewayCommand) []ackEntry {
		entries := make([]ackEntry, 0, len(cmd.Nodes))
		for _, nodeIdx := range cmd.Nodes {
			nodeName := strings.ReplaceAll(nodePattern, "*", strconv.Itoa(nodeIdx))
			entries = append(entries, ackEntry{
				Key:   cmd.AckKey(nodeName),
				Value: cmd.Ack(nodeIdx, nodeName),
			})
		}
		return entries
	}, nil)
}

// NewSrcrouterAcker creates a builder that watches for reload srcrouter requests under the given namespace
// and acks them after a random delay to simulate src-router pods acknowledging reload requests.
// nodeNames is a list of src-router node names to ack as (e.g. "srcrouter-0", "srcrouter-1").
func NewSrcrouterAcker(ctx context.Context, g *errgroup.Group,
	client *clientv3.Client, namespace string, nodeNames []string,
) *Builder[clustertypes.ReloadSrcRouterCommand] {
	prefix := "/" + namespace + "/reload/src-router/request/"
	return newBuilder(ctx, g, client, prefix, func(_ string, cmd clustertypes.ReloadSrcRouterCommand) []ackEntry {
		entries := make([]ackEntry, 0, len(nodeNames))
		for _, nodeName := range nodeNames {
			entries = append(entries, ackEntry{
				Key:   cmd.AckKey(nodeName),
				Value: cmd.Ack(nodeName),
			})
		}
		return entries
	}, nil)
}

// NewJobAcker creates a builder that watches for new migration job keys under the given namespace
// and marks them as completed after a random delay to simulate processor nodes executing migration jobs.
func NewJobAcker(ctx context.Context, g *errgroup.Group,
	client *clientv3.Client, namespace string,
) *Builder[clustertypes.PartitionMigrationJob] {
	prefix := "/" + namespace + "/migration/job/"
	return newBuilder(ctx, g, client, prefix, func(key string, job clustertypes.PartitionMigrationJob) []ackEntry {
		job.Status = clustertypes.PartitionMigrationJobStatusCompleted
		return []ackEntry{{Key: key, Value: &job}}
	}, func(e *etcdwatcher.Event[clustertypes.PartitionMigrationJob]) bool {
		return e.Value.Status != clustertypes.PartitionMigrationJobStatusCompleted
	})
}

// Start begins watching for events and acknowledging them.
func (b *Builder[T]) Start() error {
	bld := etcdwatcher.NewBuilder[T](b.client, b.prefix).
		WithPrefix().
		WithWatchMode(etcdwatcher.OnceMode).
		WithWatchEventType(etcdwatcher.PutWatchEventType)
	if b.filter != nil {
		bld = bld.WithFilter(b.filter)
	}
	w, err := bld.Build()
	if err != nil {
		return fmt.Errorf("creating acker watcher: %w", err)
	}

	initialEvents, events, leave := w.LoadAndWatch(b.ctx)
	if initialEvents.Error != nil {
		leave()
		return fmt.Errorf("loading existing requests: %w", initialEvents.Error)
	}

	b.g.Go(func() error {
		defer leave()
		for _, event := range initialEvents.Events {
			if b.eventListener != nil {
				b.eventListener(event.Key, event.Value)
			}
			if err := b.writeAcks(mapEntries(b.mapFn, event.Key, event.Value)); err != nil {
				return err
			}
		}
		for event := range events {
			if event.Error != nil {
				return fmt.Errorf("watching requests: %w", event.Error)
			}
			if b.eventListener != nil {
				b.eventListener(event.Event.Key, event.Event.Value)
			}
			if err := b.writeAcks(mapEntries(b.mapFn, event.Event.Key, event.Event.Value)); err != nil {
				return err
			}
		}
		return nil
	})

	return nil
}

// Builder configures and starts an acker. The type parameter T is the event type
// and is always inferred from the constructor — callers never need to specify it.
type Builder[T any] struct {
	ctx    context.Context
	g      *errgroup.Group
	client *clientv3.Client
	prefix string
	mapFn  func(key string, value T) []ackEntry
	filter func(*etcdwatcher.Event[T]) bool

	eventListener func(key string, value T)
	ackListener   func(ackKey string)
	minDelay      time.Duration
	maxDelay      time.Duration
}

func newBuilder[T any](ctx context.Context, g *errgroup.Group,
	client *clientv3.Client, prefix string,
	mapFn func(key string, value T) []ackEntry,
	filter func(*etcdwatcher.Event[T]) bool,
) *Builder[T] {
	return &Builder[T]{
		ctx:      ctx,
		g:        g,
		client:   client,
		prefix:   prefix,
		mapFn:    mapFn,
		filter:   filter,
		minDelay: 100 * time.Millisecond,
		maxDelay: 1 * time.Second,
	}
}

func mapEntries[T any](mapFn func(string, T) []ackEntry, key string, value T) []ackEntry {
	return mapFn(key, value)
}

func (b *Builder[T]) writeAcks(entries []ackEntry) error {
	for _, e := range entries {
		if !sleepWithContext(b.ctx, randomDelay(b.minDelay, b.maxDelay)) {
			return nil
		}
		data, err := jsonrs.Marshal(e.Value)
		if err != nil {
			return fmt.Errorf("marshalling ack: %w", err)
		}
		if _, err := b.client.Put(b.ctx, e.Key, string(data)); err != nil {
			return fmt.Errorf("writing ack: %w", err)
		}
		if b.ackListener != nil {
			b.ackListener(e.Key)
		}
	}
	return nil
}

func randomDelay(minDelay, maxDelay time.Duration) time.Duration {
	if maxDelay <= minDelay {
		return minDelay
	}
	return minDelay + rand.N(maxDelay-minDelay)
}

func sleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

type ackEntry struct {
	Key   string
	Value any
}
