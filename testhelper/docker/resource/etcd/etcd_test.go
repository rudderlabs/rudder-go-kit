package etcd

import (
	"context"
	"testing"

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
