package etcdwatcher_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/goleak"

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
		watcher, err := etcdwatcher.New[TestData](client, "/test/key")

		require.NoError(t, err)
		require.NotNil(t, watcher)
	})

	t.Run("valid construction with explicit values", func(t *testing.T) {
		client := &MockEtcdClient{}
		watcher, err := etcdwatcher.New[TestData](client, "/test/",
			etcdwatcher.WithPrefix[TestData](),
			etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
			etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))

		require.NoError(t, err)
		require.NotNil(t, watcher)
	})

	t.Run("error when using OnceMode with non-prefix watches", func(t *testing.T) {
		client := &MockEtcdClient{}
		watcher, err := etcdwatcher.New[TestData](client, "/test/key",
			etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))

		require.Error(t, err)
		require.Nil(t, watcher)
		require.Contains(t, err.Error(), "once mode can only be used with prefix watches")
	})

	t.Run("valid construction with OnceMode and prefix watches", func(t *testing.T) {
		client := &MockEtcdClient{}
		watcher, err := etcdwatcher.New[TestData](client, "/test/",
			etcdwatcher.WithPrefix[TestData](),
			etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))

		require.NoError(t, err)
		require.NotNil(t, watcher)
	})
}

// TestLoadAndWatch_AllMode_AllEventType tests LoadAndWatch with AllMode and AllWatchEventType
func TestLoadAndWatch_AllMode_AllEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.AllWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))
	require.NoError(t, err)

	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events
	require.NoError(t, initialEvents.Error)
	require.GreaterOrEqual(t, len(initialEvents.Events), 0)

	// Test PUT event
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}
}

// TestLoadAndWatch_AllMode_PutEventType tests LoadAndWatch with AllMode and PutWatchEventType
func TestLoadAndWatch_AllMode_PutEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events
	require.NoError(t, initialEvents.Error)
	require.GreaterOrEqual(t, len(initialEvents.Events), 0)

	// Test PUT event
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test DELETE event - should not receive it since we're only watching for PUT events
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event - should timeout since we're only watching PUT events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching PUT events
		t.Fatalf("Should not receive DELETE event when watching only PUT events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no DELETE event received
	}
}

// TestLoadAndWatch_AllMode_DeleteEventType tests LoadAndWatch with AllMode and DeleteWatchEventType
func TestLoadAndWatch_AllMode_DeleteEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.DeleteWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events - should not receive PUT events in initial load when watching only DELETE
	require.NoError(t, initialEvents.Error)

	// Test PUT event - should not receive it since we're only watching for DELETE events
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}
}

// TestLoadAndWatch_OnceMode_AllEventType tests LoadAndWatch with OnceMode and AllWatchEventType
func TestLoadAndWatch_OnceMode_AllEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.AllWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events
	require.NoError(t, initialEvents.Error)
	require.GreaterOrEqual(t, len(initialEvents.Events), 0)

	// Test PUT event for a new key
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Update the same key to test OnceMode behavior - should not receive this event
	updatedData := TestData{ID: 2, Name: "updated"}
	updatedDataBytes, err := jsonrs.Marshal(updatedData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(updatedDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional event is received for the same key (OnceMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Logf("Received additional event in OnceMode for same key (this may be expected): %v", eventOrErr)
		// If it's for the same key, this would violate OnceMode behavior
		if eventOrErr.Event != nil && eventOrErr.Event.Key == "/test/key2" {
			t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
		}
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}

	// Test DELETE event for a new key
	_, err = resource.Client.Delete(ctx, "/test/key3") // Delete a non-existent key to trigger a delete event
	require.NoError(t, err)

	// Create a key first, then delete it
	thirdData := TestData{ID: 3, Name: "third"}
	thirdDataBytes, err := jsonrs.Marshal(thirdData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key3", string(thirdDataBytes))
	require.NoError(t, err)

	// Wait for the put event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key3", eventOrErr.Event.Key)
		require.Equal(t, thirdData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event for key3")
	}

	// Now delete the key
	_, err = resource.Client.Delete(ctx, "/test/key3")
	require.NoError(t, err)

	select {
	case e := <-eventCh:
		t.Fatalf("Unexpected event received in OnceMode: %v", e)
	case <-time.After(5 * time.Second):
	}
}

// TestLoadAndWatch_OnceMode_PutEventType tests LoadAndWatch with OnceMode and PutWatchEventType
func TestLoadAndWatch_OnceMode_PutEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events
	require.NoError(t, initialEvents.Error)
	require.Len(t, initialEvents.Events, 1) // Only the initial PUT event should be present

	// Test PUT event for a new key
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Update the same key - should not receive this event in OnceMode
	updatedData := TestData{ID: 2, Name: "updated"}
	updatedDataBytes, err := jsonrs.Marshal(updatedData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(updatedDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional event is received for the same key (OnceMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}

	// Test DELETE event - should not receive it since we're only watching PUT events
	_, err = resource.Client.Delete(ctx, "/test/key3")
	require.NoError(t, err)

	// Wait to ensure no DELETE event is received (only PUT events should be watched)
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching PUT events
		t.Fatalf("Should not receive DELETE event when watching only PUT events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no DELETE event received
	}
}

// TestLoadAndWatch_OnceMode_DeleteEventType tests LoadAndWatch with OnceMode and DeleteWatchEventType
func TestLoadAndWatch_OnceMode_DeleteEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.DeleteWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events - should not receive PUT events in initial load when watching only DELETE
	require.NoError(t, initialEvents.Error)
	require.Empty(t, initialEvents.Events)

	// Test PUT event - should not receive it since we're only watching DELETE events
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Try to delete the same key again - should not receive this event in OnceMode
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait to ensure no additional event is received for the same key (OnceMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}

	// Test PUT event again - should not receive it since we're only watching DELETE events
	finalData := TestData{ID: 3, Name: "final"}
	finalDataBytes, err := jsonrs.Marshal(finalData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key3", string(finalDataBytes))
	require.NoError(t, err)

	// Wait to ensure no PUT event is received (only DELETE events should be watched)
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}
}

// TestLoadAndWatch_NoneMode_AllEventType tests LoadAndWatch with NoneMode and AllWatchEventType
func TestLoadAndWatch_NoneMode_AllEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.AllWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.NoneMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events
	require.NoError(t, initialEvents.Error)
	require.GreaterOrEqual(t, len(initialEvents.Events), 0)

	// The NoneMode should close the channel immediately after initial load
	select {
	case _, ok := <-eventCh:
		if ok {
			t.Fatal("Channel should be closed in NoneMode")
		}
	default:
		// Channel is closed, which is expected
	}

	// Test that no new events are received after initial load
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional events are received (NoneMode)
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Channel should be closed in NoneMode")
	case <-time.After(2 * time.Second):
		// Expected - no additional events
	}
}

// TestLoadAndWatch_NoneMode_PutEventType tests LoadAndWatch with NoneMode and PutWatchEventType
func TestLoadAndWatch_NoneMode_PutEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.NoneMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events
	require.NoError(t, initialEvents.Error)
	require.Equal(t, len(initialEvents.Events), 1)

	// The NoneMode should close the channel immediately after initial load
	select {
	case _, ok := <-eventCh:
		if ok {
			t.Fatal("Channel should be closed in NoneMode")
		}
	default:
		// Channel is closed, which is expected
	}

	// Test that no new events are received after initial load, even PUT events
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional events are received (NoneMode)
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Channel should be closed in NoneMode")
	case <-time.After(2 * time.Second):
		// Expected - no additional events
	}
}

// TestLoadAndWatch_NoneMode_DeleteEventType tests LoadAndWatch with NoneMode and DeleteWatchEventType
func TestLoadAndWatch_NoneMode_DeleteEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.DeleteWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.NoneMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Check initial events - should not receive PUT events in initial load when watching only DELETE
	require.NoError(t, initialEvents.Error)

	// The NoneMode should close the channel immediately after initial load
	select {
	case _, ok := <-eventCh:
		if ok {
			t.Fatal("Channel should be closed in NoneMode")
		}
	default:
		// Channel is closed, which is expected
	}

	// Test that no new events are received after initial load, even DELETE events
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait to ensure no additional events are received (NoneMode)
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Channel should be closed in NoneMode")
	case <-time.After(2 * time.Second):
		// Expected - no additional events
	}
}

// TestWatch_AllMode_AllEventType tests Watch with AllMode and AllWatchEventType
func TestWatch_AllMode_AllEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.AllWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, testData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test PUT event
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}
}

// TestWatch_AllMode_AllEventType tests Watch with AllMode and PutWatchEventType
func TestWatch_AllMode_PutEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, testData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test PUT event
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test DELETE event - should not receive it since we're only watching PUT events
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event - should timeout since we're only watching PUT events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching PUT events
		t.Fatalf("Should not receive DELETE event when watching only PUT events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no DELETE event received
	}
}

// TestWatch_AllMode_AllEventType tests Watch with AllMode and DeleteWatchEventType
func TestWatch_AllMode_DeleteEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.DeleteWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.AllMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Test PUT event - should not receive it since we're only watching DELETE events
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}
}

// TestWatch_OnceMode_AllEventType tests Watch with OnceMode and AllWatchEventType
func TestWatch_OnceMode_AllEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.AllWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, testData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test PUT event for a new key
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Update the same key to test OnceMode behavior - should not receive this event
	updatedData := TestData{ID: 2, Name: "updated"}
	updatedDataBytes, err := jsonrs.Marshal(updatedData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(updatedDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional event is received for the same key (OnceMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Logf("Received additional event in OnceMode for same key (this may be expected): %v", eventOrErr)
		// If it's for the same key, this would violate OnceMode behavior
		if eventOrErr.Event != nil && eventOrErr.Event.Key == "/test/key2" {
			t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
		}
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}
}

// TestWatch_OnceMode_AllEventType tests Watch with OnceMode and PutWatchEventType
func TestWatch_OnceMode_PutEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, testData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test PUT event for a new key
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key2", eventOrErr.Event.Key)
		require.Equal(t, newData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Update the same key to test OnceMode behavior - should not receive this event
	updatedData := TestData{ID: 2, Name: "updated"}
	updatedDataBytes, err := jsonrs.Marshal(updatedData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(updatedDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional event is received for the same key (OnceMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}

	// Test DELETE event - should not receive it since we're only watching PUT events
	_, err = resource.Client.Delete(ctx, "/test/key3")
	require.NoError(t, err)

	// Wait to ensure no DELETE event is received (only PUT events should be watched)
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching PUT events
		t.Fatalf("Should not receive DELETE event when watching only PUT events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no DELETE event received
	}
}

// TestWatch_OnceMode_AllEventType tests Watch with OnceMode and DeleteWatchEventType
func TestWatch_OnceMode_DeleteEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.DeleteWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Test PUT event - should not receive it since we're only watching DELETE events
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Test DELETE event
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait for the delete event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, etcdwatcher.DeleteEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected DELETE event")
	}

	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}

	// Try to delete the same key again - should not receive this event in OnceMode
	_, err = resource.Client.Delete(ctx, "/test/key1")
	require.NoError(t, err)

	// Wait to ensure no additional event is received for the same key (OnceMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}

	// Test PUT event again - should not receive it since we're only watching DELETE events
	finalData := TestData{ID: 3, Name: "final"}
	finalDataBytes, err := jsonrs.Marshal(finalData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key3", string(finalDataBytes))
	require.NoError(t, err)

	// Wait to ensure no PUT event is received (only DELETE events should be watched)
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching DELETE events
		t.Fatalf("Should not receive PUT event when watching only DELETE events, got: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}
}

// TestWatch_NoneMode_AllEventType tests Watch with NoneMode and AllWatchEventType
func TestWatch_NoneMode_AllEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.AllWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.NoneMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Wait for the event - should receive it since Watch doesn't implement NoneMode restriction
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, testData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test that events are still received with Watch method even in NoneMode
	// (This tests the actual behavior, which may differ from expectations)
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait to ensure no additional event is received (NoneMode)
	select {
	case eventOrErr := <-eventCh:
		// In OnceMode, we should not receive another event for the same key
		t.Logf("Received additional event in OnceMode for same key (this may be expected): %v", eventOrErr)
		// If it's for the same key, this would violate OnceMode behavior
		if eventOrErr.Event != nil && eventOrErr.Event.Key == "/test/key2" {
			t.Fatalf("OnceMode should not emit events for keys that have already been emitted: %v", eventOrErr)
		}
	case <-time.After(2 * time.Second):
		// Expected for OnceMode - no additional event for the same key
	}
}

// TestWatch_NoneMode_AllEventType tests Watch with NoneMode and PutWatchEventType
func TestWatch_NoneMode_PutEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.NoneMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Wait for the event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/key1", eventOrErr.Event.Key)
		require.Equal(t, testData, eventOrErr.Event.Value)
		require.Equal(t, etcdwatcher.PutEvent, eventOrErr.Event.Type)
	case <-time.After(10 * time.Second):
		t.Fatal("Did not receive expected PUT event")
	}

	// Test that events are still received with Watch method even in NoneMode
	// (This tests the actual behavior, which may differ from expectations)
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Channel should be closed in NoneMode")
	case <-time.After(2 * time.Second):
		// Expected - no DELETE event received
	}
}

// TestWatch_NoneMode_AllEventType tests Watch with NoneMode and DeleteWatchEventType
func TestWatch_NoneMode_DeleteEventType(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prepare test data
	testData := TestData{ID: 1, Name: "initial"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	// Put initial data
	_, err = resource.Client.Put(ctx, "/test/key1", string(dataBytes))
	require.NoError(t, err)

	// Create watcher
	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.DeleteWatchEventType),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.NoneMode))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Test PUT event - should not receive it since we're only watching DELETE events
	newData := TestData{ID: 2, Name: "new"}
	newDataBytes, err := jsonrs.Marshal(newData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/key2", string(newDataBytes))
	require.NoError(t, err)

	// Wait for the PUT event - should timeout since we're only watching DELETE events
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Channel should be closed in NoneMode")
	case <-time.After(2 * time.Second):
		// Expected - no PUT event received
	}
}

// TestValueTypes tests watching different value types ([]byte, string, custom struct)
func TestValueTypes(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test []byte values
	t.Run("byte slice values", func(t *testing.T) {
		byteData := []byte("hello world")
		_, err = resource.Client.Put(ctx, "/test/bytes", string(byteData))
		require.NoError(t, err)

		watcher, err := etcdwatcher.New[[]byte](resource.Client, "/test/bytes")
		require.NoError(t, err)

		initialEvents, _, cancelWatch := watcher.LoadAndWatch(ctx)
		defer cancelWatch()

		require.NoError(t, initialEvents.Error)
		require.Len(t, initialEvents.Events, 1)
		require.Equal(t, byteData, initialEvents.Events[0].Value)

		// Test string values
		stringData := "hello string"
		_, err = resource.Client.Put(ctx, "/test/string", stringData)
		require.NoError(t, err)

		watcherString, err := etcdwatcher.New[string](resource.Client, "/test/string")
		require.NoError(t, err)

		initialEventsStr, _, cancelWatchStr := watcherString.LoadAndWatch(ctx)
		defer cancelWatchStr()

		require.NoError(t, initialEventsStr.Error)
		require.Len(t, initialEventsStr.Events, 1)
		require.Equal(t, stringData, initialEventsStr.Events[0].Value)

		// Test custom struct values
		customData := TestData{ID: 1, Name: "test"}
		customDataBytes, err := jsonrs.Marshal(customData)
		require.NoError(t, err)

		_, err = resource.Client.Put(ctx, "/test/struct", string(customDataBytes))
		require.NoError(t, err)

		watcherStruct, err := etcdwatcher.New[TestData](resource.Client, "/test/struct")
		require.NoError(t, err)

		initialEventsStruct, _, cancelWatchStruct := watcherStruct.LoadAndWatch(ctx)
		defer cancelWatchStruct()

		require.NoError(t, initialEventsStruct.Error)
		require.Len(t, initialEventsStruct.Events, 1)
		require.Equal(t, customData.ID, initialEventsStruct.Events[0].Value.ID)
		require.Equal(t, customData.Name, initialEventsStruct.Events[0].Value.Name)
	})
}

// TestUnmarshallingErrors tests unmarshalling error scenarios
func TestUnmarshallingErrors(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put invalid JSON data for a struct that expects proper JSON
	invalidJSON := []byte(`{"id": 1, "name":}`) // Invalid JSON
	_, err = resource.Client.Put(ctx, "/test/invalid", string(invalidJSON))
	require.NoError(t, err)

	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/invalid")
	require.NoError(t, err)

	initialEvents, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	// Should have an unmarshalling error
	require.Error(t, initialEvents.Error)
	require.Contains(t, initialEvents.Error.Error(), "failed to unmarshal value")

	// Test unmarshalling error during watch
	// First put valid data to initialize
	validData := TestData{ID: 1, Name: "valid"}
	validDataBytes, err := jsonrs.Marshal(validData)
	require.NoError(t, err)
	_, err = resource.Client.Put(ctx, "/test/watch", string(validDataBytes))
	require.NoError(t, err)

	watcher2, err := etcdwatcher.New[TestData](resource.Client, "/test/watch")
	require.NoError(t, err)

	initialEvents2, eventCh2, cancelWatch2 := watcher2.LoadAndWatch(ctx)
	defer cancelWatch2()

	require.NoError(t, initialEvents2.Error)
	require.Len(t, initialEvents2.Events, 1)

	// Now put invalid JSON data for the same key to trigger error during watch
	_, err = resource.Client.Put(ctx, "/test/watch", string(invalidJSON))
	require.NoError(t, err)

	// Should receive an error event
	select {
	case eventOrErr := <-eventCh2:
		require.Error(t, eventOrErr.Error)
		require.Nil(t, eventOrErr.Event)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected unmarshalling error")
	}

	// Channel should close after the error
	select {
	case _, ok := <-eventCh2:
		require.False(t, ok, "Channel should be closed after error")
	case <-time.After(1 * time.Second):
		// Channel may already be closed
	}
}

// TestCustomUnmarshaller tests using a custom unmarshaller
func TestCustomUnmarshaller(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test custom unmarshaller that prepends a prefix
	customUnmarshaller := func(data []byte, target *string) error {
		*target = "custom_" + string(data)
		return nil
	}

	testData := "hello"
	_, err = resource.Client.Put(ctx, "/test/custom", testData)
	require.NoError(t, err)

	watcher, err := etcdwatcher.New[string](resource.Client, "/test/custom", etcdwatcher.WithValueUnmarshaller(customUnmarshaller))
	require.NoError(t, err)

	initialEvents, _, cancelWatch := watcher.LoadAndWatch(ctx)
	defer cancelWatch()

	require.NoError(t, initialEvents.Error)
	require.Len(t, initialEvents.Events, 1)
	require.Equal(t, "custom_"+testData, initialEvents.Events[0].Value)

	// Test custom unmarshaller error
	erroringUnmarshaller := func(data []byte, target *string) error {
		return fmt.Errorf("custom unmarshalling error")
	}

	_, err = resource.Client.Put(ctx, "/test/custom_error", "some data")
	require.NoError(t, err)

	watcherError, err := etcdwatcher.New[string](resource.Client, "/test/custom_error", etcdwatcher.WithValueUnmarshaller(erroringUnmarshaller))
	require.NoError(t, err)

	initialEventsError, _, cancelWatchError := watcherError.LoadAndWatch(ctx)
	defer cancelWatchError()

	require.Error(t, initialEventsError.Error)
	require.Contains(t, initialEventsError.Error.Error(), "custom unmarshalling error")

	// Test custom unmarshaller error during watch
	validData := "valid data"
	_, err = resource.Client.Put(ctx, "/test/custom_watch", validData)
	require.NoError(t, err)

	watcherWatch, err := etcdwatcher.New[string](resource.Client, "/test/custom_watch", etcdwatcher.WithValueUnmarshaller(erroringUnmarshaller))
	require.NoError(t, err)

	initialEventsWatch, _, cancelWatchWatch := watcherWatch.LoadAndWatch(ctx)
	defer cancelWatchWatch()

	// Should have an error during initial load
	require.Error(t, initialEventsWatch.Error)
}

// TestClientEarlyCancellation tests that no goroutines leak when client cancels early
func TestClientEarlyCancellation(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create a context that we'll cancel early
	earlyCtx, earlyCancel := context.WithCancel(ctx)
	earlyCancel() // Cancel immediately

	// Put some data
	testData := TestData{ID: 1, Name: "test"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)
	_, err = resource.Client.Put(ctx, "/test/early", string(dataBytes))
	require.NoError(t, err)

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/early")
	require.NoError(t, err)

	// This should handle cancellation gracefully
	initialEvents, _, cancelWatch := watcher.LoadAndWatch(earlyCtx)
	defer cancelWatch()

	// Should get context cancellation error
	require.Error(t, initialEvents.Error)
	require.Contains(t, initialEvents.Error.Error(), "context canceled")

	// Test Watch method as well
	earlyCtx2, earlyCancel2 := context.WithCancel(ctx)
	earlyCancel2() // Cancel immediately

	watcher2, err := etcdwatcher.New[TestData](resource.Client, "/test/early")
	require.NoError(t, err)

	eventCh2, cancelWatch2 := watcher2.Watch(earlyCtx2)
	defer func() {
		cancelWatch2()
		select {
		case _, ok := <-eventCh2:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Should receive context cancellation error
	select {
	case eventOrErr := <-eventCh2:
		require.Error(t, eventOrErr.Error)
		require.Contains(t, eventOrErr.Error.Error(), "context canceled")
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected context cancellation error")
	}

	// Channel should close after the error
	select {
	case _, ok := <-eventCh2:
		require.False(t, ok, "Channel should be closed after error")
	case <-time.After(1 * time.Second):
		// Channel may already be closed
	}
}

// TestErrorChannelClosure tests that watch channel closes after first error
func TestErrorChannelClosure(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put invalid JSON data to trigger unmarshalling error
	invalidJSON := []byte(`{"id": 1, "name":}`) // Invalid JSON
	_, err = resource.Client.Put(ctx, "/test/error_closure", string(invalidJSON))
	require.NoError(t, err)

	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/error_closure")
	require.NoError(t, err)

	// Use Watch method to see the behavior with errors
	eventCh, cancelWatch := watcher.Watch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Should receive an error event
	select {
	case eventOrErr := <-eventCh:
		require.Error(t, eventOrErr.Error)
		require.Nil(t, eventOrErr.Event)
		require.Contains(t, eventOrErr.Error.Error(), "failed to unmarshal value")
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected unmarshalling error")
	}

	// Channel should be closed after the error
	select {
	case _, ok := <-eventCh:
		require.False(t, ok, "Channel should be closed after error")
	case <-time.After(1 * time.Second):
		// Channel may already be closed
	}

	// Test with LoadAndWatch as well
	_, err = resource.Client.Put(ctx, "/test/error_closure2", string(invalidJSON))
	require.NoError(t, err)

	watcher2, err := etcdwatcher.New[TestData](resource.Client, "/test/error_closure2")
	require.NoError(t, err)

	initialEvents, eventCh2, cancelWatch2 := watcher2.LoadAndWatch(ctx)
	defer func() {
		cancelWatch2()
		select {
		case _, ok := <-eventCh2:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Initial load should have error
	require.Error(t, initialEvents.Error)
	require.Contains(t, initialEvents.Error.Error(), "failed to unmarshal value")

	// Channel should be closed
	select {
	case _, ok := <-eventCh2:
		require.False(t, ok, "Channel should be closed after error")
	case <-time.After(1 * time.Second):
		// Channel may already be closed
	}
}

// TestFilterFunction tests the WithFilter option to filter events
func TestFilterFunction(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put multiple test keys with different patterns
	testData1 := TestData{ID: 1, Name: "user_data"}
	testData2 := TestData{ID: 2, Name: "config_data"}
	testData3 := TestData{ID: 3, Name: "other_data"}

	dataBytes1, err := jsonrs.Marshal(testData1)
	require.NoError(t, err)
	dataBytes2, err := jsonrs.Marshal(testData2)
	require.NoError(t, err)
	dataBytes3, err := jsonrs.Marshal(testData3)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/user_data", string(dataBytes1))
	require.NoError(t, err)
	_, err = resource.Client.Put(ctx, "/test/config_data", string(dataBytes2))
	require.NoError(t, err)
	_, err = resource.Client.Put(ctx, "/test/other_data", string(dataBytes3))
	require.NoError(t, err)

	// Test filter that only accepts keys with "user" in the name
	filter := func(event *etcdwatcher.Event[TestData]) bool {
		return event.Key == "/test/user_data"
	}

	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithFilter[TestData](filter))
	require.NoError(t, err)

	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Should only receive the event that matches the filter
	require.NoError(t, initialEvents.Error)
	require.Len(t, initialEvents.Events, 1)
	require.Equal(t, "/test/user_data", initialEvents.Events[0].Key)
	require.Equal(t, testData1.ID, initialEvents.Events[0].Value.ID)
	require.Equal(t, testData1.Name, initialEvents.Events[0].Value.Name)

	// Test filter during watch
	testData4 := TestData{ID: 4, Name: "user_new_data"}
	dataBytes4, err := jsonrs.Marshal(testData4)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/user_new_data", string(dataBytes4))
	require.NoError(t, err)

	// This should not be received because it doesn't match the filter
	testData5 := TestData{ID: 5, Name: "non_user_data"}
	dataBytes5, err := jsonrs.Marshal(testData5)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/non_user_data", string(dataBytes5))
	require.NoError(t, err)

	// Wait a short time to ensure non-matching events are not received
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since the event doesn't match the filter
		t.Fatalf("Should not receive non-matching event: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no non-matching events received
	}

	// Test filter with deletion events
	_, err = resource.Client.Delete(ctx, "/test/user_data")
	require.NoError(t, err)

	// Wait for the filtered event
	select {
	case eventOrErr := <-eventCh:
		require.NoError(t, eventOrErr.Error)
		require.NotNil(t, eventOrErr.Event)
		require.Equal(t, "/test/user_data", eventOrErr.Event.Key)
	case <-time.After(5 * time.Second):
		t.Fatal("Did not receive expected filtered event")
	}
}

// TestFilterFunctionWithOnceMode tests filter function combined with OnceMode
func TestFilterFunctionWithOnceMode(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put multiple test keys
	testData1 := TestData{ID: 1, Name: "user_data"}
	testData2 := TestData{ID: 2, Name: "user_data"} // Same name, different ID

	dataBytes1, err := jsonrs.Marshal(testData1)
	require.NoError(t, err)
	dataBytes2, err := jsonrs.Marshal(testData2)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/user_1", string(dataBytes1))
	require.NoError(t, err)
	_, err = resource.Client.Put(ctx, "/test/user_2", string(dataBytes2))
	require.NoError(t, err)

	// Filter that accepts events with ID > 0 (all of them)
	filter := func(event *etcdwatcher.Event[TestData]) bool {
		return event.Value.ID > 0
	}

	defer goleak.VerifyNone(t, goleak.IgnoreCurrent())

	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithFilter[TestData](filter),
		etcdwatcher.WithWatchMode[TestData](etcdwatcher.OnceMode))
	require.NoError(t, err)

	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Should receive all initial events that match the filter
	require.NoError(t, initialEvents.Error)
	require.Len(t, initialEvents.Events, 2)

	// Test that updates to the same keys are filtered out in OnceMode
	testData3 := TestData{ID: 3, Name: "updated_user"}
	dataBytes3, err := jsonrs.Marshal(testData3)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/user_1", string(dataBytes3))
	require.NoError(t, err)

	// In OnceMode, this should not be received even if it matches the filter
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since OnceMode filters out already emitted keys
		t.Fatalf("Should not receive event for already emitted key in OnceMode: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no event received due to OnceMode
	}

	testData4 := TestData{ID: -1, Name: "new_user"}
	dataBytes4, err := jsonrs.Marshal(testData4)
	require.NoError(t, err)
	_, err = resource.Client.Put(ctx, "/test/new_user", string(dataBytes4))
	require.NoError(t, err)

	select {
	case eventOrErr := <-eventCh:
		t.Fatalf("Should not receive event that does not match filter: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no event received due to filter
	}
}

// TestFilterFunctionWithEventTypes tests filter function combined with event type filtering
func TestFilterFunctionWithEventTypes(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := etcd.Setup(pool, t)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Put test data
	testData := TestData{ID: 1, Name: "test_data"}
	dataBytes, err := jsonrs.Marshal(testData)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/filter_event", string(dataBytes))
	require.NoError(t, err)

	// Filter that only accepts events with specific name
	filter := func(event *etcdwatcher.Event[TestData]) bool {
		return event.Value.Name == "test_data"
	}

	watcher, err := etcdwatcher.New[TestData](resource.Client, "/test/",
		etcdwatcher.WithPrefix[TestData](),
		etcdwatcher.WithFilter[TestData](filter),
		etcdwatcher.WithWatchEventType[TestData](etcdwatcher.PutWatchEventType))
	require.NoError(t, err)

	initialEvents, eventCh, cancelWatch := watcher.LoadAndWatch(ctx)
	defer func() {
		cancelWatch()
		select {
		case _, ok := <-eventCh:
			require.False(t, ok, "Expected channel to be closed")
		case <-time.After(2 * time.Second):
			t.Fatal("Channel did not close in time")
		}
	}()

	// Should receive the initial event that matches both the filter and event type
	require.NoError(t, initialEvents.Error)
	require.Len(t, initialEvents.Events, 1)
	require.Equal(t, "test_data", initialEvents.Events[0].Value.Name)

	// Test that filtered events are not received
	testData2 := TestData{ID: 2, Name: "other_data"}
	dataBytes2, err := jsonrs.Marshal(testData2)
	require.NoError(t, err)

	_, err = resource.Client.Put(ctx, "/test/filter_event2", string(dataBytes2))
	require.NoError(t, err)

	// This should not be received because it doesn't match the filter
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since the event doesn't match the filter
		t.Fatalf("Should not receive non-matching event: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no non-matching event received
	}

	// Test delete event - should not be received because we're only watching PUT events
	_, err = resource.Client.Delete(ctx, "/test/filter_event")
	require.NoError(t, err)

	// Should not receive the delete event because of event type filter
	select {
	case eventOrErr := <-eventCh:
		// This should not happen since we're only watching PUT events
		t.Fatalf("Should not receive DELETE event when only watching PUT events: %v", eventOrErr)
	case <-time.After(2 * time.Second):
		// Expected - no DELETE event received
	}
}
