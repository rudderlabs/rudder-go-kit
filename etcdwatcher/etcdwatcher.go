// Package etcdwatcher provides an abstraction for watching etcd keys and prefixes.
package etcdwatcher

import (
	"context"
	"fmt"
	"sync"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
}

// EventType represents the type of event that occurred.
type EventType string

const (
	// PutEvent represents a key creation or update event.
	PutEvent EventType = "PUT"
	// DeleteEvent represents a key deletion event.
	DeleteEvent EventType = "DELETE"
)

// WatchMode defines how the watcher handles events.
type WatchMode string

const (
	// AllMode emits all events including updates for keys that have already been emitted.
	AllMode WatchMode = "ALL"
	// OnceMode Emit each key at most once across the entire lifecycle (initial load + watch).
	// Subsequent events for that key are ignored, regardless of type.
	// this mode can only be used with prefix watches.
	OnceMode WatchMode = "ONCE"
	// NoneMode will only emit the existing keys at the time of start and then stop.
	NoneMode WatchMode = "NONE"
)

// Event represents a single etcd key event.
type Event[T any] struct {
	// Key is the etcd key that changed.
	Key string
	// Type is the type of event (PUT or DELETE).
	Type EventType
	// Value is the unmarshalled value of the key (for PUT events).
	Value T
	// Revision is the etcd revision of the event.
	Revision int64
}

// WatchEventType defines what types of events to watch for.
type WatchEventType string

const (
	// PutWatchEventType watches for put events only.
	PutWatchEventType WatchEventType = "PUT"
	// DeleteWatchEventType watches for delete events only.
	DeleteWatchEventType WatchEventType = "DELETE"
	// AllWatchEventType watches for both put and delete events.
	AllWatchEventType WatchEventType = "ALL"
)

// Config holds the configuration for the watcher.
type Config[T any] struct {
	// EventTypes specifies which event types to watch for.
	EventTypes WatchEventType
	// Mode specifies the watch mode.
	Mode WatchMode
	// Filter is an optional function to filter events.
	Filter func(Event[T]) bool
}

// Watcher provides methods for watching etcd keys.
type Watcher[T any] struct {
	client   etcdClient
	config   Config[T]
	key      string
	isPrefix bool
}

// New creates a new Watcher instance.
func New[T any](client etcdClient, key string, isPrefix bool, config Config[T]) (*Watcher[T], error) {
	// Set defaults
	if config.EventTypes == "" {
		config.EventTypes = AllWatchEventType
	}
	if config.Mode == "" {
		config.Mode = AllMode
	}

	if config.Mode == OnceMode && !isPrefix {
		return nil, fmt.Errorf("once mode can only be used with prefix watches")
	}

	return &Watcher[T]{
		client:   client,
		config:   config,
		key:      key,
		isPrefix: isPrefix,
	}, nil
}

// LoadAndWatch loads initial data and then starts watching for changes.
// It returns the initial events, a channel for subsequent events and a function to stop watching for events.
func (w *Watcher[T]) LoadAndWatch(ctx context.Context) ([]Event[T], <-chan Event[T], <-chan error, func()) {
	// Create a new context for this watch operation
	watchCtx, cancel := context.WithCancel(ctx)

	// Channel for events
	eventCh := make(chan Event[T])
	errorCh := make(chan error, 1) // Buffered to prevent blocking

	// Load initial data
	initialEvents, revision, err := w.loadInitialData(watchCtx)
	if err != nil {
		// Send error and close channels
		go func() {
			errorCh <- err
		}()
		return nil, eventCh, errorCh, cancel
	}

	// If in NoneMode, we don't need to watch for changes
	if w.config.Mode == NoneMode {
		close(eventCh)
		close(errorCh)
		return initialEvents, eventCh, errorCh, cancel
	}

	// Start watching in a separate goroutine
	go w.watch(watchCtx, initialEvents, eventCh, errorCh, revision)

	return initialEvents, eventCh, errorCh, cancel
}

// Watch watches for changes and returns all events(after watch is started) through a single channel.
func (w *Watcher[T]) Watch(ctx context.Context) (<-chan Event[T], <-chan error, func(), error) {
	if w.config.Mode == NoneMode {
		return nil, nil, nil, fmt.Errorf("cannot watch in NoneMode")
	}

	// Create a new context for this watch operation
	watchCtx, cancel := context.WithCancel(ctx)

	// Channel for events
	eventCh := make(chan Event[T])
	errorCh := make(chan error, 1) // Buffered to prevent blocking

	// Load initial data
	initialEvents, revision, err := w.loadInitialData(watchCtx)
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("loading initial data: %w", err)
	}

	// Start watching in a separate goroutine
	go w.watch(watchCtx, initialEvents, eventCh, errorCh, revision)

	return eventCh, errorCh, cancel, nil
}

// loadInitialData loads the initial data for the given key or prefix.
func (w *Watcher[T]) loadInitialData(ctx context.Context) ([]Event[T], int64, error) {
	var resp *clientv3.GetResponse
	var err error

	if w.isPrefix {
		resp, err = w.client.Get(ctx, w.key, clientv3.WithPrefix())
	} else {
		resp, err = w.client.Get(ctx, w.key)
	}

	if err != nil {
		return nil, 0, err
	}

	events := make([]Event[T], 0, len(resp.Kvs))
	emittedKeys := make(map[string]bool)

	for _, kv := range resp.Kvs {
		// In OnceMode or NoneMode, track emitted keys to avoid duplicates
		if w.config.Mode == OnceMode {
			if emittedKeys[string(kv.Key)] {
				continue
			}
			emittedKeys[string(kv.Key)] = true
		}

		event := Event[T]{
			Key:      string(kv.Key),
			Type:     PutEvent,
			Revision: kv.ModRevision,
		}

		err = jsonrs.Unmarshal(kv.Value, &event.Value)
		if err != nil {
			// Skip values that can't be unmarshalled
			continue
		}

		// Apply filter if present
		if w.config.Filter != nil && !w.config.Filter(event) {
			continue
		}

		events = append(events, event)
	}

	return events, resp.Header.GetRevision(), nil
}

// watch starts watching etcd for changes.
func (w *Watcher[T]) watch(ctx context.Context, initialEvents []Event[T], eventCh chan<- Event[T], errorCh chan<- error, revision int64) {
	defer func() {
		close(eventCh)
		close(errorCh)
	}()
	// Track emitted keys for OnceMode
	emittedKeys := make(map[string]bool)
	var emittedKeysMutex sync.Mutex

	// Initialize with keys from initial events if in OnceMode
	if w.config.Mode == OnceMode {
		emittedKeysMutex.Lock()
		for _, event := range initialEvents {
			emittedKeys[event.Key] = true
		}
		emittedKeysMutex.Unlock()
	}

	// Prepare watch options
	opts := []clientv3.OpOption{
		clientv3.WithRev(revision + 1),
	}
	if w.isPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}

	// Create a watch channel
	watchCh := w.client.Watch(ctx, w.key, opts...)

	// Process events
	for {
		select {
		case <-ctx.Done():
			return
		case watchResp, ok := <-watchCh:
			if !ok {
				// Watch channel closed
				return
			}

			if err := watchResp.Err(); err != nil {
				select {
				case errorCh <- err:
				case <-ctx.Done():
				}
				return
			}

			for _, ev := range watchResp.Events {
				eventType := PutEvent
				if ev.Type == mvccpb.DELETE {
					eventType = DeleteEvent
				}

				// Check if we should process this event type
				shouldProcess := false
				switch w.config.EventTypes {
				case PutWatchEventType:
					shouldProcess = eventType == PutEvent
				case DeleteWatchEventType:
					shouldProcess = eventType == DeleteEvent
				case AllWatchEventType:
					shouldProcess = true
				}

				if !shouldProcess {
					continue
				}

				event := Event[T]{
					Key:      string(ev.Kv.Key),
					Type:     eventType,
					Revision: ev.Kv.ModRevision,
				}

				// For PUT events, unmarshal the value
				if eventType == PutEvent {
					err := jsonrs.Unmarshal(ev.Kv.Value, &event.Value)
					if err != nil {
						errorCh <- err
						continue
					}
				}

				// Apply filter if present
				if w.config.Filter != nil && !w.config.Filter(event) {
					continue
				}

				// Handle OnceMode - skip already emitted keys
				if w.config.Mode == OnceMode {
					emittedKeysMutex.Lock()
					if eventType == PutEvent && emittedKeys[event.Key] {
						emittedKeysMutex.Unlock()
						continue
					}
					if eventType == PutEvent {
						emittedKeys[event.Key] = true
					}
					emittedKeysMutex.Unlock()
				}

				// Send event
				select {
				case eventCh <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}
