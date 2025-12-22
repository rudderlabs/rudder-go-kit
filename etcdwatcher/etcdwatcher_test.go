package etcdwatcher_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/api/v3/mvccpb"

	"github.com/ory/dockertest/v3"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/etcd"

	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/jsonrs"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/rudderlabs/rudder-go-kit/etcdwatcher"
)

// TestData represents the structure of data stored in etcd for testing
type TestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// MockEtcdClient is a mock implementation of the etcd client
type MockEtcdClient struct {
	watchChans map[string]chan clientv3.WatchResponse
	getResp    *clientv3.GetResponse
	getErr     error
}

func (m *MockEtcdClient) Get(_ context.Context, _ string, _ ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	return m.getResp, m.getErr
}

func (m *MockEtcdClient) Watch(_ context.Context, key string, _ ...clientv3.OpOption) clientv3.WatchChan {
	return m.watchChans[key]
}

// Constructor tests
func TestWatcher_New(t *testing.T) {
	t.Run("valid construction with default values", func(t *testing.T) {
		client := &MockEtcdClient{}
		watcher, err := etcdwatcher.New[TestData](client, "/test/key", false, etcdwatcher.Config[TestData]{})

		require.NoError(t, err)
		require.NotNil(t, watcher)
	})

	t.Run("valid construction with explicit values", func(t *testing.T) {
		client := &MockEtcdClient{}
		config := etcdwatcher.Config[TestData]{
			EventTypes: etcdwatcher.PutWatchEventType,
			Mode:       etcdwatcher.AllMode,
		}
		watcher, err := etcdwatcher.New[TestData](client, "/test/key", false, config)

		require.NoError(t, err)
		require.NotNil(t, watcher)
	})

	t.Run("error when using OnceMode with non-prefix watches", func(t *testing.T) {
		client := &MockEtcdClient{}
		config := etcdwatcher.Config[TestData]{
			Mode: etcdwatcher.OnceMode,
		}
		watcher, err := etcdwatcher.New[TestData](client, "/test/key", false, config)

		require.Error(t, err)
		require.Nil(t, watcher)
		require.Contains(t, err.Error(), "once mode can only be used with prefix watches")
	})

	t.Run("valid construction with OnceMode and prefix watches", func(t *testing.T) {
		client := &MockEtcdClient{}
		config := etcdwatcher.Config[TestData]{
			Mode: etcdwatcher.OnceMode,
		}
		watcher, err := etcdwatcher.New[TestData](client, "/test/", true, config)

		require.NoError(t, err)
		require.NotNil(t, watcher)
	})
}

// Integration tests - Basic functionality
func TestWatcher_BasicFunctionality(t *testing.T) {
	// Setup etcd container
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	// Test basic functionality
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put some initial data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	config := etcdwatcher.Config[TestData]{
		EventTypes: etcdwatcher.AllWatchEventType,
		Mode:       etcdwatcher.AllMode,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	// Load and watch
	initialEvents, eventCh, errCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Check initial data
	require.Len(t, initialEvents, 1)
	require.Equal(t, "/test/key1", initialEvents[0].Key)
	require.Equal(t, testData, initialEvents[0].Value)
	require.Equal(t, etcdwatcher.PutEvent, initialEvents[0].Type)

	// Add new data and verify it's received
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/key2", event.Key)
		require.Equal(t, newData, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case err := <-errCh:
		require.NoError(t, err, "received unexpected error from watcher")
	case <-time.After(60 * time.Second):
		t.Fatal("Did not receive expected event")
	}
}

func TestWatcher_DeleteEvents(t *testing.T) {
	// Setup etcd container
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	// Test delete event handling
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put initial data
	testData := TestData{ID: 1, Name: "to_be_deleted"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/delete_key", string(dataBytes))
	require.NoError(t, err)

	// Create watcher with delete event type
	config := etcdwatcher.Config[TestData]{
		EventTypes: etcdwatcher.AllWatchEventType,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	_, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Delete the key
	_, err = resource.Client.Delete(ctx, "/test/delete_key")
	require.NoError(t, err)

	// Wait for delete event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/delete_key", event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, event.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected delete event")
	}
}

// Mode tests
func TestWatcher_OnceMode(t *testing.T) {
	// Setup etcd container
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	// Test OnceMode behavior
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put initial data
	testData1 := TestData{ID: 1, Name: "initial"}
	dataBytes1, err := jsonrs.Marshal(testData1)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/once_key", string(dataBytes1))
	require.NoError(t, err)

	// Create watcher with OnceMode
	config := etcdwatcher.Config[TestData]{
		Mode: etcdwatcher.OnceMode,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	initialEvents, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Check initial data
	require.Len(t, initialEvents, 1)
	require.Equal(t, "/test/once_key", initialEvents[0].Key)
	require.Equal(t, testData1, initialEvents[0].Value)

	// Update the same key - should be ignored in OnceMode
	testData2 := TestData{ID: 2, Name: "updated"}
	dataBytes2, err := jsonrs.Marshal(testData2)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/once_key", string(dataBytes2))
	require.NoError(t, err)

	// Add a new key - should be received
	testData3 := TestData{ID: 3, Name: "new"}
	dataBytes3, err := jsonrs.Marshal(testData3)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/once_key_new", string(dataBytes3))
	require.NoError(t, err)

	// Should only receive the new key, not the update to existing key
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/once_key_new", event.Key)
		require.Equal(t, testData3, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected event for new key")
	}

	// Ensure we don't receive any more events
	select {
	case event := <-eventCh:
		t.Fatalf("Received unexpected event: %+v", event)
	case <-time.After(100 * time.Millisecond):
		// Expected - no more events
	}
}

func TestWatcher_NoneMode(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put some initial data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher with NoneMode
	config := etcdwatcher.Config[TestData]{
		Mode: etcdwatcher.NoneMode,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	// Load and watch
	initialEvents, eventCh, errCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Check initial data is loaded
	require.Len(t, initialEvents, 1)
	require.Equal(t, "/test/key1", initialEvents[0].Key)
	require.Equal(t, testData, initialEvents[0].Value)

	// Add new data - should not be received since we're in NoneMode
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Verify no events are received (channel should be nil or empty)
	select {
	case _, ok := <-eventCh:
		require.False(t, ok)
	case _, ok := <-errCh:
		require.False(t, ok)
	case <-time.After(100 * time.Millisecond):
		// Expected - no events in NoneMode
	}
}

// Event type tests
func TestWatcher_PutWatchEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create watcher with PutWatchEventType
	config := etcdwatcher.Config[TestData]{
		EventTypes: etcdwatcher.PutWatchEventType,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	_, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Add new data - should be received
	newData := TestData{ID: 1, Name: "put_only"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key1", string(newDataBytes))
	require.NoError(t, err)

	// Should receive PUT event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/key1", event.Key)
		require.Equal(t, newData, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Delete the key - should NOT be received
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Verify no DELETE event is received
	select {
	case event := <-eventCh:
		t.Fatalf("Received unexpected DELETE event with PutWatchEventType: %+v", event)
	case <-time.After(5 * time.Second):
		// Expected - no DELETE events with PutWatchEventType
	}
}

func TestWatcher_DeleteWatchEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put initial data
	testData := TestData{ID: 1, Name: "delete_only"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher with DeleteWatchEventType
	config := etcdwatcher.Config[TestData]{
		EventTypes: etcdwatcher.DeleteWatchEventType,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	_, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Delete the key - should be received
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Should receive DELETE event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/key1", event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, event.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}

	// Add new data - should NOT be received
	newData := TestData{ID: 2, Name: "new_put"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Verify no PUT event is received
	select {
	case event := <-eventCh:
		t.Fatalf("Received unexpected PUT event with DeleteWatchEventType: %+v", event)
	case <-time.After(5 * time.Second):
		// Expected - no PUT events with DeleteWatchEventType
	}
}

// Filter function tests
func TestWatcher_FilterFunctionAllFiltered(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Define filter that rejects all events
	config := etcdwatcher.Config[TestData]{
		Filter: func(event etcdwatcher.Event[TestData]) bool {
			return false // Reject all
		},
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	_, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Put data that should all be filtered out
	testData := TestData{ID: 1, Name: "filtered_out"}
	testDataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/filter_key", string(testDataBytes))
	require.NoError(t, err)

	// Ensure we don't receive any events
	select {
	case event := <-eventCh:
		t.Fatalf("Received unexpected event despite filter rejecting all: %+v", event)
	case <-time.After(100 * time.Millisecond):
		// Expected - no events due to filter
	}
}

func TestWatcher_NilFilterFunction(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create config with nil filter (default behavior)
	config := etcdwatcher.Config[TestData]{
		Filter: nil, // Explicitly nil
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	_, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Put data - should be received since there's no filter
	testData := TestData{ID: 1, Name: "no_filter"}
	testDataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/no_filter_key", string(testDataBytes))
	require.NoError(t, err)

	// Should receive the event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/no_filter_key", event.Key)
		require.Equal(t, testData, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case <-time.After(60 * time.Second):
		t.Fatal("Did not receive expected event with nil filter")
	}
}

// Non-prefix key watching
func TestWatcher_NonPrefixKey(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put some data at exact key
	testData := TestData{ID: 1, Name: "exact_key"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/exact_key", string(dataBytes))
	require.NoError(t, err)

	// Create watcher for exact key (non-prefix)
	config := etcdwatcher.Config[TestData]{}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/exact_key", false, config)
	require.NoError(t, err)

	// Load and watch
	initialEvents, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Check initial data
	require.Len(t, initialEvents, 1)
	require.Equal(t, "/test/exact_key", initialEvents[0].Key)
	require.Equal(t, testData, initialEvents[0].Value)

	// Update the exact key - should be received
	updatedData := TestData{ID: 2, Name: "updated"}
	updatedDataBytes, err := jsonrs.Marshal(updatedData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/exact_key", string(updatedDataBytes))
	require.NoError(t, err)

	// Should receive update event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/exact_key", event.Key)
		require.Equal(t, updatedData, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected update event")
	}
}

// Malformed JSON handling
func TestWatcher_MalformedJSON(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put good data
	goodData := TestData{ID: 1, Name: "good"}
	goodDataBytes, err := jsonrs.Marshal(goodData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/good_key", string(goodDataBytes))
	require.NoError(t, err)

	// Create watcher
	config := etcdwatcher.Config[TestData]{}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	// Load and watch
	initialEvents, eventCh, errCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Should only get the good data (malformed JSON should be skipped)
	require.Len(t, initialEvents, 1)
	require.Equal(t, "/test/good_key", initialEvents[0].Key)
	require.Equal(t, goodData, initialEvents[0].Value)

	// Add another malformed entry - should not cause issues
	_, err = resource.Client.Put(ctx, "/test/bad_key", "{ invalid json }")
	require.NoError(t, err)

	// Ensure we don't receive anything for the bad JSON
	select {
	case event := <-eventCh:
		t.Fatalf("Received unexpected event for malformed JSON: %+v", event)
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(100 * time.Millisecond):
		// Expected - no events for malformed JSON
	}
}

// Empty response handling
func TestWatcher_EmptyResponse(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create watcher for a prefix with no existing keys
	config := etcdwatcher.Config[TestData]{}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/empty/", true, config)
	require.NoError(t, err)

	initialEvents, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Should get empty initial events
	require.Len(t, initialEvents, 0)

	// Add a new key and verify it's received
	testData := TestData{ID: 1, Name: "new"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/empty/new_key", string(dataBytes))
	require.NoError(t, err)

	// Should receive the new event
	select {
	case event := <-eventCh:
		require.Equal(t, "/empty/new_key", event.Key)
		require.Equal(t, testData, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected event")
	}
}

// Context cancellation
func TestWatcher_ContextCancellation(t *testing.T) {
	// Setup etcd container
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	// Test context cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create watcher
	config := etcdwatcher.Config[TestData]{}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	_, eventCh, _, stop := watcher.LoadAndWatch(ctx)

	stop()

	// Channel should be closed or we should get zero-value events
	select {
	case event, ok := <-eventCh:
		// Either channel is closed (ok == false) or we get a zero-value event
		if ok {
			// If we got an event, it should be empty/zero-value
			require.Empty(t, event.Key)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Channel was not closed in reasonable time")
	}
}

// Watch method tests
func TestWatcher_WatchMethod(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Put some initial data
	testData := TestData{ID: 1, Name: "watch_method"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	config := etcdwatcher.Config[TestData]{}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	// Use Watch method instead of LoadAndWatch
	eventCh, errCh, stop, err := watcher.Watch(ctx)
	require.NoError(t, err)
	defer stop()

	// Add new data and verify it's received
	newData := TestData{ID: 2, Name: "new_in_watch"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/key2", event.Key)
		require.Equal(t, newData, event.Value)
		require.Equal(t, etcdwatcher.PutEvent, event.Type)
	case err := <-errCh:
		require.NoError(t, err, "received unexpected error from watcher")
	case <-time.After(15 * time.Second):
		t.Fatal("Did not receive expected event")
	}
}

func TestWatcher_WatchWithNoneMode(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create watcher with NoneMode
	config := etcdwatcher.Config[TestData]{
		Mode: etcdwatcher.NoneMode,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	// Watch method should return an error in NoneMode
	eventCh, errCh, stop, err := watcher.Watch(ctx)
	require.Error(t, err)
	require.Nil(t, eventCh)
	require.Nil(t, errCh)
	require.Nil(t, stop)
	require.Contains(t, err.Error(), "cannot watch in NoneMode")
}

// Error handling tests
func TestWatcher_GetError_LoadAndWatch(t *testing.T) {
	client := &MockEtcdClient{
		getErr: fmt.Errorf("get error"),
	}
	watcher, err := etcdwatcher.New[TestData](client, "/test/key", false, etcdwatcher.Config[TestData]{})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	initialEvents, eventCh, errCh, stop := watcher.LoadAndWatch(ctx)
	defer stop()

	// Should receive error
	select {
	case err := <-errCh:
		require.Error(t, err)
		require.Contains(t, err.Error(), "get error")
	case <-time.After(1 * time.Second):
		t.Fatal("Did not receive expected error")
	}

	// Initial events should be nil
	require.Nil(t, initialEvents)

	// Event channel should be available but empty
	select {
	case event := <-eventCh:
		require.Empty(t, event.Key) // Zero value
	default:
		// Channel is open but empty, which is expected
	}
}

func TestWatcher_GetError_Watch(t *testing.T) {
	client := &MockEtcdClient{
		getErr: fmt.Errorf("get error"),
	}
	watcher, err := etcdwatcher.New[TestData](client, "/test/key", false, etcdwatcher.Config[TestData]{})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, _, _, err = watcher.Watch(ctx)
	require.Error(t, err)
}

func TestWatcher_WatchResponseError(t *testing.T) {
	client := &MockEtcdClient{
		watchChans: map[string]chan clientv3.WatchResponse{
			"/test/key": make(chan clientv3.WatchResponse, 1),
		},
		getResp: &clientv3.GetResponse{
			Header: &etcdserverpb.ResponseHeader{
				Revision: 1,
			},
			Kvs: []*mvccpb.KeyValue{},
		},
	}
	watcher, err := etcdwatcher.New[TestData](client, "/test/key", false, etcdwatcher.Config[TestData]{})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, eventCh, errCh, stop := watcher.LoadAndWatch(ctx)
	defer stop()

	// Send an error response on the watch channel
	go func() {
		client.watchChans["/test/key"] <- clientv3.WatchResponse{
			Header: etcdserverpb.ResponseHeader{
				Revision: 2,
			},
			Canceled: true,
		}
		close(client.watchChans["/test/key"])
	}()

	// Should receive the error
	select {
	case err := <-errCh:
		require.Error(t, err)
		require.Contains(t, err.Error(), "future revision")
	case <-time.After(1 * time.Second):
		t.Fatal("Did not receive expected error from watch response")
	}

	// Event channel should be closed
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Event channel should be closed after error")
	case <-time.After(100 * time.Millisecond):
		// Channel might already be closed
	}
}

// Special tests
func TestWatcher_OnceModeWithDeleteEvents(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put initial data
	testData1 := TestData{ID: 1, Name: "to_be_deleted"}
	dataBytes1, err := jsonrs.Marshal(testData1)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/once_delete_key", string(dataBytes1))
	require.NoError(t, err)

	// Create watcher with OnceMode
	config := etcdwatcher.Config[TestData]{
		Mode:       etcdwatcher.OnceMode,
		EventTypes: etcdwatcher.AllWatchEventType,
	}
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/", true, config)
	require.NoError(t, err)

	initialEvents, eventCh, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Check initial data
	require.Len(t, initialEvents, 1)
	require.Equal(t, "/test/once_delete_key", initialEvents[0].Key)
	require.Equal(t, testData1, initialEvents[0].Value)

	// Delete the key - should be received since it was only emitted once (not filtered by OnceMode for PUT)
	_, err = resource.Client.Delete(ctx, "/test/once_delete_key")
	require.NoError(t, err)

	// Should receive the DELETE event
	select {
	case event := <-eventCh:
		require.Equal(t, "/test/once_delete_key", event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, event.Type)
		// Value should be zero-value for delete events
		require.Equal(t, TestData{}, event.Value)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected DELETE event in OnceMode")
	}

	// Put the key again - should NOT be received in OnceMode
	testData2 := TestData{ID: 2, Name: "updated_after_delete"}
	dataBytes2, err := jsonrs.Marshal(testData2)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/once_delete_key", string(dataBytes2))
	require.NoError(t, err)

	// Verify we don't receive the update event (key was already emitted in PUT or DELETE)
	select {
	case event := <-eventCh:
		t.Fatalf("Received unexpected event in OnceMode after DELETE: %+v", event)
	case <-time.After(100 * time.Millisecond):
		// Expected - no more events for the same key in OnceMode
	}
}
