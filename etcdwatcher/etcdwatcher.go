// Package etcdwatcher provides an abstraction for watching etcd keys and prefixes.
package etcdwatcher

import (
	"context"
	"fmt"

	"github.com/rudderlabs/rudder-go-kit/async"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type etcdClient interface {
	Get(ctx context.Context, key string, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan
	Txn(ctx context.Context) clientv3.Txn
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
	// Version is the version of the key. A deletion resets the version to zero.
	Version int64
}

type EventOrError[T any] struct {
	Event *Event[T]
	Error error
}
type EventsOrError[T any] struct {
	Events []*Event[T]
	Error  error
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

// config holds the configuration for the watcher.
type config[T any] struct {
	// EventTypes specifies which event types to watch for.
	eventTypes WatchEventType
	// Mode specifies the watch mode.
	mode WatchMode
	// Filter is an optional function to filter events.
	filter func(*Event[T]) bool
	// UnmarshalValue is a function to unmarshal the value of a key.
	unmarshalValue func([]byte) (T, error)
	// isPrefix indicates whether the watcher is watching a prefix or a single key.
	isPrefix bool
	// txnGate is an optional function that returns conditions that must be met for a PUT event to pass through.
	txnGate func(T) []clientv3.Cmp
}

// Watcher provides methods for watching etcd keys.
type Watcher[T any] struct {
	client etcdClient
	config config[T]
	key    string
}

// LoadAndWatch loads initial data and then starts watching for changes.
// It returns the initial events, a channel for subsequent events and a function to stop watching for events.
func (w *Watcher[T]) LoadAndWatch(ctx context.Context) (EventsOrError[T], <-chan EventOrError[T], func()) {
	eventSender := &async.SingleSender[EventOrError[T]]{}
	ctx, watchChan, _ := eventSender.Begin(ctx)

	// Channel for events
	initialEvents := EventsOrError[T]{}
	// Load initial data
	events, revision, err := w.loadInitialData(ctx)
	if err != nil {
		eventSender.Close()
		initialEvents.Error = err
		return initialEvents, watchChan, eventSender.Close
	}
	initialEvents.Events = events

	// If in NoneMode, we don't need to watch for changes
	if w.config.mode == NoneMode {
		eventSender.Close()
		return initialEvents, watchChan, eventSender.Close
	}

	// Start watching in a separate goroutine
	go w.watch(ctx, events, eventSender, revision)

	return initialEvents, watchChan, eventSender.Close
}

// Watch watches for changes and returns all events(including initial set of events) through a channel.
func (w *Watcher[T]) Watch(ctx context.Context) (<-chan EventOrError[T], func()) {
	eventSender := &async.SingleSender[EventOrError[T]]{}
	ctx, watchChan, _ := eventSender.Begin(ctx)

	go func() {
		// Load initial data
		initialEvents, revision, err := w.loadInitialData(ctx)
		if err != nil {
			eventSender.Send(EventOrError[T]{Error: err})
			eventSender.Close()
			return
		}

		if len(initialEvents) > 0 {
			for _, event := range initialEvents {
				eventSender.Send(EventOrError[T]{Event: event})
			}
		}

		if w.config.mode == NoneMode {
			eventSender.Close()
			return
		}

		w.watch(ctx, initialEvents, eventSender, revision)
	}()

	return watchChan, eventSender.Close
}

// loadInitialData loads the initial data for the given key or prefix.
func (w *Watcher[T]) loadInitialData(ctx context.Context) ([]*Event[T], int64, error) {
	var resp *clientv3.GetResponse
	var err error

	if w.config.isPrefix {
		resp, err = w.client.Get(ctx, w.key, clientv3.WithPrefix())
	} else {
		resp, err = w.client.Get(ctx, w.key)
	}

	if err != nil {
		return nil, 0, err
	}

	events := make([]*Event[T], 0, len(resp.Kvs))

	// If only watching for delete events, return early as there are no delete events in initial load
	if w.config.eventTypes == DeleteWatchEventType {
		return events, resp.Header.GetRevision(), nil
	}

	for _, kv := range resp.Kvs {
		event := &Event[T]{
			Key:      string(kv.Key),
			Type:     PutEvent,
			Revision: kv.ModRevision,
			Version:  kv.Version,
		}

		event.Value, err = w.config.unmarshalValue(kv.Value)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to unmarshal value: %w", err)
		}

		// Apply filter if present
		if w.config.filter != nil && !w.config.filter(event) {
			continue
		}

		// Apply txn gate if configured
		if w.config.txnGate != nil {
			event, err = w.applyTxnGate(ctx, event)
			if err != nil {
				return nil, 0, err
			}
			if event == nil {
				continue
			}
		}

		events = append(events, event)
	}

	return events, resp.Header.GetRevision(), nil
}

// watch starts watching etcd for changes.
func (w *Watcher[T]) watch(ctx context.Context, initialEvents []*Event[T], eventSender *async.SingleSender[EventOrError[T]], revision int64) {
	defer eventSender.Close()
	// Track emitted keys for OnceMode
	emittedKeys := make(map[string]struct{})

	// Initialize with keys from initial events if in OnceMode
	if w.config.mode == OnceMode {
		for _, event := range initialEvents {
			emittedKeys[event.Key] = struct{}{}
		}
	}

	// Prepare watch options
	opts := []clientv3.OpOption{
		clientv3.WithRev(revision + 1),
	}
	if w.config.isPrefix {
		opts = append(opts, clientv3.WithPrefix())
	}

	// Create a watch channel
	watchCh := w.client.Watch(ctx, w.key, opts...)

	// Process events
	for watchResp := range watchCh {
		if err := watchResp.Err(); err != nil {
			eventSender.Send(EventOrError[T]{Error: err})
			return
		}

		for _, ev := range watchResp.Events {
			eventType := PutEvent
			if ev.Type == mvccpb.DELETE {
				eventType = DeleteEvent
			}

			// Check if we should process this event type
			shouldProcess := false
			switch w.config.eventTypes {
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

			event := &Event[T]{
				Key:      string(ev.Kv.Key),
				Type:     eventType,
				Revision: ev.Kv.ModRevision,
				Version:  ev.Kv.Version,
			}

			// For PUT events, unmarshal the value
			if eventType == PutEvent {
				value, err := w.config.unmarshalValue(ev.Kv.Value)
				if err != nil {
					eventSender.Send(EventOrError[T]{Error: err})
					return
				}
				event.Value = value
			}

			// Apply filter if present
			if w.config.filter != nil && !w.config.filter(event) {
				continue
			}

			// Apply txn gate if configured (only for PUT events)
			if eventType == PutEvent && w.config.txnGate != nil {
				var err error
				event, err = w.applyTxnGate(ctx, event)
				if err != nil {
					eventSender.Send(EventOrError[T]{Error: err})
					return
				}
				if event == nil {
					continue
				}
			}

			// Handle OnceMode - skip already emitted keys
			if w.config.mode == OnceMode {
				if _, emitted := emittedKeys[event.Key]; emitted {
					continue
				}
				emittedKeys[event.Key] = struct{}{}
			}

			// Send event
			eventSender.Send(EventOrError[T]{Event: event})
		}
	}
}

// applyTxnGate executes the transactional gate for a PUT event.
// It returns the updated event if the transaction succeeds and the event passes filtering,
// nil if the event should be skipped, or an error if something went wrong.
func (w *Watcher[T]) applyTxnGate(ctx context.Context, event *Event[T]) (*Event[T], error) {
	cmps := w.config.txnGate(event.Value)
	if len(cmps) == 0 {
		return event, nil
	}

	txnResp, err := w.client.Txn(ctx).
		If(cmps...).
		Then(clientv3.OpGet(event.Key)).
		Commit()
	if err != nil {
		return nil, fmt.Errorf("executing txn gate: %w", err)
	}

	if !txnResp.Succeeded { // Transaction conditions not met, skip this event
		return nil, nil // nolint: nilnil
	}

	rangeResp := txnResp.Responses[0].GetResponseRange()
	if len(rangeResp.Kvs) == 0 {
		return nil, nil // nolint: nilnil
	}

	kv := rangeResp.Kvs[0]
	value, err := w.config.unmarshalValue(kv.Value)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling txn gate value: %w", err)
	}

	event = &Event[T]{
		Key:      string(kv.Key),
		Type:     PutEvent,
		Value:    value,
		Revision: kv.ModRevision,
		Version:  kv.Version,
	}

	if w.config.filter != nil && !w.config.filter(event) {
		return nil, nil // nolint: nilnil
	}

	return event, nil
}

func defaultUnmarshalValue[T any](data []byte) (T, error) {
	var target T
	switch v := any(&target).(type) {
	case *[]byte:
		*v = append([]byte(nil), data...)
		return target, nil
	case *string:
		*v = string(data)
		return target, nil
	default:
		err := jsonrs.Unmarshal(data, &target)
		return target, err
	}
}

// NewBuilder creates a new fluent builder for Watcher[T].
func NewBuilder[T any](client etcdClient, key string) *Builder[T] {
	return &Builder[T]{
		client: client,
		key:    key,
		config: config[T]{
			eventTypes:     AllWatchEventType,
			mode:           AllMode,
			unmarshalValue: defaultUnmarshalValue[T],
		},
	}
}

// Builder is a fluent builder for Watcher[T].
type Builder[T any] struct {
	client etcdClient
	key    string
	config config[T]
}

// WithPrefix configures the watcher to watch a prefix. By default, the watcher watches a single key.
func (b *Builder[T]) WithPrefix() *Builder[T] {
	b.config.isPrefix = true
	return b
}

// WithWatchEventType configures the watcher to watch specific event types (PUT, DELETE, or ALL). By default, it watches for ALL event types.
func (b *Builder[T]) WithWatchEventType(eventType WatchEventType) *Builder[T] {
	b.config.eventTypes = eventType
	return b
}

// WithWatchMode configures the watcher to use a specific watch mode (ALL, ONCE, or NONE). By default, it uses ALL mode.
func (b *Builder[T]) WithWatchMode(mode WatchMode) *Builder[T] {
	b.config.mode = mode
	return b
}

// WithFilter configures the watcher to use a filter function to filter events.
// Only events for which the filter function returns false will be filtered.
func (b *Builder[T]) WithFilter(filter func(*Event[T]) bool) *Builder[T] {
	b.config.filter = filter
	return b
}

// WithTxnGate configures the watcher to use a transactional gate for PUT events.
// After a PUT event passes filtering, the watcher calls the provided function with the event's
// value to obtain If conditions, then executes an etcd Txn with those conditions (if any) and a Then
// that re-fetches the key. If the transaction's conditions aren't met, the event is silently
// filtered out. If successful, the re-fetched value is unmarshalled and re-filtered.
func (b *Builder[T]) WithTxnGate(txnGate func(T) []clientv3.Cmp) *Builder[T] {
	b.config.txnGate = txnGate
	return b
}

// WithValueUnmarshaller configures the watcher to use a custom unmarshaller function for values.
func (b *Builder[T]) WithValueUnmarshaller(unmarshalFunc func([]byte, *T) error) *Builder[T] {
	b.config.unmarshalValue = func(data []byte) (T, error) {
		var target T
		err := unmarshalFunc(data, &target)
		return target, err
	}
	return b
}

// Build creates a new Watcher instance with the configured options.
func (b *Builder[T]) Build() (*Watcher[T], error) {
	if b.config.mode == OnceMode && !b.config.isPrefix {
		return nil, fmt.Errorf("once mode can only be used with prefix watches")
	}

	return &Watcher[T]{
		client: b.client,
		config: b.config,
		key:    b.key,
	}, nil
}
