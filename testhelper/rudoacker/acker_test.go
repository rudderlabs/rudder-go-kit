package rudoacker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	clustertypes "github.com/rudderlabs/rudder-schemas/go/cluster"
)

func TestMigrationAcker_JobsDBs(t *testing.T) {
	newMigration := func(jobsDBFanout bool) clustertypes.PartitionMigration {
		return clustertypes.PartitionMigration{
			ID: "m1",
			Jobs: []*clustertypes.PartitionMigrationJobHeader{
				{JobID: "job-1", SourceNode: 0, TargetNode: 1},
			},
			AckKeyPrefix: "/ns/migration/ack/m1",
			Features:     clustertypes.PartitionMigrationFeatures{JobsDBFanout: jobsDBFanout},
		}
	}

	acksByNode := func(b *Builder[clustertypes.PartitionMigration], migration clustertypes.PartitionMigration) map[int]*clustertypes.PartitionMigrationAck {
		entries := b.mapFn("", migration)
		out := make(map[int]*clustertypes.PartitionMigrationAck, len(entries))
		for _, e := range entries {
			ack, ok := e.Value.(*clustertypes.PartitionMigrationAck)
			require.True(t, ok, "ack entry value must be *PartitionMigrationAck")
			out[ack.NodeIndex] = ack
		}
		return out
	}

	t.Run("legacy: nodes ack without JobsDBs by default", func(t *testing.T) {
		b := NewMigrationAcker(context.Background(), &errgroup.Group{}, nil, "ns")
		acks := acksByNode(b, newMigration(true))
		require.Len(t, acks, 2)
		require.Empty(t, acks[0].JobsDBs)
		require.Empty(t, acks[1].JobsDBs)
	})

	t.Run("source node declares configured JobsDBs when fan-out is enabled", func(t *testing.T) {
		b := NewMigrationAcker(context.Background(), &errgroup.Group{}, nil, "ns").
			WithJobsDBs(map[int][]string{0: {"gw", "rt"}})
		acks := acksByNode(b, newMigration(true))
		require.Len(t, acks, 2)
		require.Equal(t, []string{"gw", "rt"}, acks[0].JobsDBs, "source node 0 should declare JobsDBs")
		require.Empty(t, acks[1].JobsDBs, "target node 1 should not declare JobsDBs")
	})

	t.Run("configured JobsDBs are ignored when fan-out is disabled", func(t *testing.T) {
		b := NewMigrationAcker(context.Background(), &errgroup.Group{}, nil, "ns").
			WithJobsDBs(map[int][]string{0: {"gw", "rt"}})
		acks := acksByNode(b, newMigration(false))
		require.Len(t, acks, 2)
		require.Empty(t, acks[0].JobsDBs, "source node 0 should not declare JobsDBs when fan-out is disabled")
		require.Empty(t, acks[1].JobsDBs)
	})

	t.Run("nodes absent from the map keep legacy behaviour", func(t *testing.T) {
		b := NewMigrationAcker(context.Background(), &errgroup.Group{}, nil, "ns").
			WithJobsDBs(map[int][]string{2: {"gw"}})
		acks := acksByNode(b, newMigration(true))
		require.Empty(t, acks[0].JobsDBs)
		require.Empty(t, acks[1].JobsDBs)
	})

	t.Run("target nodes never declare JobsDBs even when configured", func(t *testing.T) {
		// node 1 is only a target node; configuring JobsDBs for it must not leak into its ack.
		b := NewMigrationAcker(context.Background(), &errgroup.Group{}, nil, "ns").
			WithJobsDBs(map[int][]string{1: {"gw"}})
		acks := acksByNode(b, newMigration(true))
		require.Len(t, acks, 2)
		require.Empty(t, acks[0].JobsDBs)
		require.Empty(t, acks[1].JobsDBs, "target node 1 should not declare JobsDBs")
	})
}
