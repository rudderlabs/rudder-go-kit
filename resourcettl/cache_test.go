package resourcettl_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/rudderlabs/rudder-go-kit/resourcettl"
)

func TestCache(t *testing.T) {
	const key = "key"
	ttl := 10 * time.Millisecond
	t.Run("checkout, checkin, then expire", func(t *testing.T) {
		t.Run("using cleanup", func(t *testing.T) {
			producer := &MockProducer{}
			c := resourcettl.NewCache(producer.NewCleanuper, ttl)

			r1, checkin1, err1 := c.Checkout(key)
			require.NoError(t, err1, "it should be able to create a new resource")
			require.NotNil(t, r1, "it should return a resource")
			require.EqualValues(t, 1, producer.instances.Load(), "it should create a new resource")

			r2, checkin2, err2 := c.Checkout(key)
			require.NoError(t, err2, "it should be able to checkout the same resource")
			require.NotNil(t, r2, "it should return a resource")
			require.EqualValues(t, 1, producer.instances.Load(), "it shouldn't create a new resource")
			require.Equal(t, r1.id, r2.id, "it should return the same resource")

			time.Sleep(ttl + time.Millisecond)
			checkin1()
			checkin2()

			r3, checkin3, err3 := c.Checkout(key)
			require.NoError(t, err3, "it should be able to create a new resource")
			require.NotNil(t, r3, "it should return a resource")
			require.EqualValues(t, 2, producer.instances.Load(), "it should create a new resource since the previous one expired")
			require.NotEqual(t, r1.id, r3.id, "it should return a different resource")
			time.Sleep(time.Millisecond) // wait for async cleanup
			require.EqualValues(t, 1, r1.cleanups.Load(), "it should cleanup the expired resource")
			checkin3()
		})

		t.Run("using closer", func(t *testing.T) {
			producer := &MockProducer{}
			c := resourcettl.NewCache(producer.NewCloser, ttl)

			r1, checkin1, err1 := c.Checkout(key)
			require.NoError(t, err1, "it should be able to create a new resource")
			require.NotNil(t, r1, "it should return a resource")
			require.EqualValues(t, 1, producer.instances.Load(), "it should create a new resource")

			r2, checkin2, err2 := c.Checkout(key)
			require.NoError(t, err2, "it should be able to checkout the same resource")
			require.NotNil(t, r2, "it should return a resource")
			require.EqualValues(t, 1, producer.instances.Load(), "it shouldn't create a new resource")
			require.Equal(t, r1.id, r2.id, "it should return the same resource")

			time.Sleep(ttl + time.Millisecond)
			checkin1()
			checkin2()

			r3, checkin3, err3 := c.Checkout(key)
			require.NoError(t, err3, "it should be able to create a new resource")
			require.NotNil(t, r3, "it should return a resource")
			require.EqualValues(t, 2, producer.instances.Load(), "it should create a new resource since the previous one expired")
			require.NotEqual(t, r1.id, r3.id, "it should return a different resource")
			time.Sleep(time.Millisecond) // wait for async cleanup
			require.EqualValues(t, 1, r1.cleanups.Load(), "it should cleanup the expired resource")
			checkin3()
		})
	})

	t.Run("expire while being used", func(t *testing.T) {
		producer := &MockProducer{}
		c := resourcettl.NewCache(producer.NewCleanuper, ttl)

		r1, checkin1, err1 := c.Checkout(key)
		require.NoError(t, err1, "it should be able to create a new resource")
		require.NotNil(t, r1, "it should return a resource")
		require.EqualValues(t, 1, producer.instances.Load(), "it should create a new resource")

		r2, checkin2, err2 := c.Checkout(key)
		require.NoError(t, err2, "it should be able to checkout the same resource")
		require.NotNil(t, r2, "it should return a resource")
		require.EqualValues(t, 1, producer.instances.Load(), "it shouldn't create a new resource")
		require.Equal(t, r1.id, r2.id, "it should return the same resource")

		time.Sleep(ttl + time.Millisecond) // wait for expiration

		r3, checkin3, err3 := c.Checkout(key)
		require.NoError(t, err3, "it should be able to return a resource")
		require.NotNil(t, r3, "it should return a resource")
		require.EqualValues(t, 2, producer.instances.Load(), "it should create a new resource since the previous one expired")
		require.NotEqual(t, r1.id, r3.id, "it should return a different resource")
		require.EqualValues(t, 0, r1.cleanups.Load(), "it shouldn't cleanup the expired resource yet since it is being used by 2 clients")
		checkin1()
		time.Sleep(time.Millisecond) // wait for async cleanup
		require.EqualValues(t, 0, r1.cleanups.Load(), "it shouldn't cleanup the expired resource yet since it is being used by 1 clients")
		checkin2()
		time.Sleep(time.Millisecond) // wait for async cleanup
		require.EqualValues(t, 1, r1.cleanups.Load(), "it should cleanup the expired resource since it is not being used by any client")
		checkin3()
	})

	t.Run("invalidate", func(t *testing.T) {
		producer := &MockProducer{}
		c := resourcettl.NewCache(producer.NewCleanuper, ttl)

		r1, checkin1, err1 := c.Checkout(key)
		require.NoError(t, err1, "it should be able to create a new resource")
		require.NotNil(t, r1, "it should return a resource")
		require.EqualValues(t, 1, producer.instances.Load(), "it should create a new resource")

		r2, checkin2, err2 := c.Checkout(key)
		require.NoError(t, err2, "it should be able to checkout the same resource")
		require.NotNil(t, r2, "it should return a resource")
		require.EqualValues(t, 1, producer.instances.Load(), "it shouldn't create a new resource")
		require.Equal(t, r1.id, r2.id, "it should return the same resource")

		c.Invalidate(key)

		r3, checkin3, err3 := c.Checkout(key)
		require.NoError(t, err3, "it should be able to create a new resource")
		require.NotNil(t, r3, "it should return a resource")
		require.EqualValues(t, 2, producer.instances.Load(), "it should create a new resource since the previous one was invalidated")
		require.NotEqual(t, r1.id, r3.id, "it should return a different resource")
		time.Sleep(time.Millisecond) // wait for async cleanup
		require.EqualValues(t, 0, r1.cleanups.Load(), "it shouldn't cleanup the expired resource yet since it is being used by 2 clients")
		checkin1()
		time.Sleep(time.Millisecond) // wait for async cleanup
		require.EqualValues(t, 0, r1.cleanups.Load(), "it shouldn't cleanup the expired resource yet since it is being used by 1 client")
		checkin2()
		time.Sleep(time.Millisecond) // wait for async cleanup
		require.EqualValues(t, 1, r1.cleanups.Load(), "it should cleanup the expired resource")
		checkin3()
	})
}

type MockProducer struct {
	instances atomic.Int32
}

func (m *MockProducer) NewCleanuper(_ string) (*cleanuper, error) {
	m.instances.Add(1)
	return &cleanuper{id: uuid.NewString()}, nil
}

func (m *MockProducer) NewCloser(_ string) (*closer, error) {
	m.instances.Add(1)
	return &closer{id: uuid.NewString()}, nil
}

type cleanuper struct {
	id       string
	cleanups atomic.Int32
}

func (m *cleanuper) Cleanup() {
	m.cleanups.Add(1)
}

type closer struct {
	id       string
	cleanups atomic.Int32
}

func (m *closer) Close() error {
	m.cleanups.Add(1)
	return nil
}
