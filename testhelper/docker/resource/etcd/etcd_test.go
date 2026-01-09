package etcd

import (
	"context"
	"testing"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

const (
	dummyKey   = "dummyKey"
	dummyValue = "dummyValue"
)

func TestETCD(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	etcdRes, err := Setup(pool, t)
	require.NoError(t, err)

	getResp, err := etcdRes.Client.Get(context.Background(), dummyKey) // no value should be present during start
	require.NoError(t, err)

	require.Zero(t, getResp.Kvs)

	_, err = etcdRes.Client.Put(context.Background(), dummyKey, dummyValue) // put value in to dummyKey
	require.NoError(t, err)

	getResp, err = etcdRes.Client.Get(context.Background(), dummyKey) // check value in dummyKey
	require.NoError(t, err)
	require.Len(t, getResp.Kvs, 1)
	require.Equal(t, string(getResp.Kvs[0].Value), dummyValue)
}

func TestETCD_Watch(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	etcdRes, err := Setup(pool, t)
	require.NoError(t, err)

	watchChan := etcdRes.Client.Watch(context.Background(), dummyKey)

	_, err = etcdRes.Client.Put(context.Background(), dummyKey, dummyValue) // put value in to dummyKey
	require.NoError(t, err)

	watchResp := <-watchChan
	require.Len(t, watchResp.Events, 1)
	require.Equal(t, string(watchResp.Events[0].Kv.Value), dummyValue)

	err = pool.Purge(etcdRes.Resource)
	require.NoError(t, err)

	select {
	case e, ok := <-watchChan:
		require.Error(t, e.Err())
		require.False(t, ok)
	case <-time.After(5 * time.Second):
		require.Fail(t, "watch channel is not closed")
	}
}
