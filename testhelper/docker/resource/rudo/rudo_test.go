package rudo_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	etcdclient "go.etcd.io/etcd/client/v3"
	"golang.org/x/sync/errgroup"

	"github.com/rudderlabs/rudder-schemas/go/cluster"

	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/etcd"
	"github.com/rudderlabs/rudder-go-kit/testhelper/docker/resource/rudo"
	"github.com/rudderlabs/rudder-go-kit/testhelper/localip"
	"github.com/rudderlabs/rudder-go-kit/testhelper/rudoacker"
)

func TestRudoResource(t *testing.T) {
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	namepace := uuid.New().String()

	bindIP := localip.GetLocalIP()
	// start etcd container
	etcdContainer, err := etcd.Setup(pool, t, etcd.WithBindIP(localip.GetLocalIP()))
	require.NoError(t, err)

	// start rudo container with ETCD_HOSTS env var pointing to the etcd container
	rudoContainer, err := rudo.Setup(pool, t,
		rudo.WithBindIP(bindIP),
		rudo.WithEtcdHosts(etcdContainer.Hosts),
		rudo.WithReleaseName(namepace),
		rudo.WithStaticWorkspaces([]string{"workspace1"}),
		rudo.WithSrcRouterNodes([]string{"srcrouter-node-0"}),
		rudo.WithAssignmentStrategy("single-node-least-loaded"),
	)
	require.NoError(t, err)

	// start a srcrouter acker
	var mu sync.Mutex
	var ackKeys []string
	var events []cluster.ReloadSrcRouterCommand
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)
	err = rudoacker.NewSrcrouterAcker(ctx, g, etcdContainer.Client, namepace, []string{"srcrouter-node-0"}).
		WithAckListener(func(ackKey string) {
			mu.Lock()
			defer mu.Unlock()
			ackKeys = append(ackKeys, ackKey)
		}).
		WithEventListener(func(key string, value cluster.ReloadSrcRouterCommand) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, value)
		}).
		Start()
	require.NoError(t, err)

	// wait until partitions are created for workspace1
	prefix := "/" + namepace + "/workspace_part_map/"
	require.Eventuallyf(t, func() bool {
		resp, err := etcdContainer.Client.Get(context.Background(), prefix, etcdclient.WithPrefix())
		if err != nil {
			return false
		}
		for _, kv := range resp.Kvs {
			if strings.HasSuffix(string(kv.Key), "workspace1") {
				return true
			}
		}
		return false
	}, 60*time.Second, 1*time.Second, "partition mapping for workspace1 not found in etcd at prefix %s, rudo URL: %s", prefix, rudoContainer.URL)

	require.Eventually(t, func() bool {
		mu.Lock()
		defer mu.Unlock()
		return len(ackKeys) == 1 && len(events) == 1
	}, 60*time.Second, 1*time.Second)
	require.Len(t, ackKeys, 1)
	require.Len(t, events, 1)

	// create a migration for workspace1
	migrationInfo, err := rudoContainer.CreateMigration(ctx, []rudo.WorkspaceMigration{{
		WorkspaceID: "workspace1",
		Migrations: []rudo.Migration{
			{
				Src: rudo.Src{
					ServerID:      0,
					PartitionIdxs: []int{0},
				},
				Dst: rudo.Dst{
					ServerID: 1,
				},
			},
		},
	}})
	require.NoError(t, err)
	require.NotEmpty(t, migrationInfo.ID)

	migrations, err := rudoContainer.ListMigrations(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, migrations)
	require.Equal(t, migrationInfo.ID, migrations[0].ID)

	cancel()
	require.NoError(t, g.Wait())
}
