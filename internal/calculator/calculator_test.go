package calculator

import (
	"math"
	"testing"
)

func TestCompute_defaultHomeJS(t *testing.T) {
	in := Inputs{
		Nodes:            3,
		DedicatedMasters: 0,
		Shards:           3,
		ReplicasPerShard: 2,
		GBSize:           90,
		Documents:        10_000_000,
		ReadRPM:          3000,
		WriteRPM:         500,
		Clusters:         1,
	}
	res := Compute(in)

	// (90/3) * (3+6) GiB * 1e9 (decimal) bytes as in Vue
	wantBytes := int64(270 * 1000 * 1000000)
	if res.ClusterBytes != wantBytes {
		t.Fatalf("ClusterBytes: got %d want %d", res.ClusterBytes, wantBytes)
	}
	if math.Abs(res.GBPerPrimaryShard-30) > 1e-9 {
		t.Fatalf("GBPerPrimaryShard: got %v want 30", res.GBPerPrimaryShard)
	}
	if res.SizeWarning != SizeWarningHighWarn {
		t.Fatalf("SizeWarning: got %s want high-warning band", res.SizeWarning)
	}
	denom := float64(3*2 + 3)
	wantRead := 3000 / denom
	if math.Abs(res.ReadPerPiece-wantRead) > 1e-9 {
		t.Fatalf("ReadPerPiece: got %v want %v", res.ReadPerPiece, wantRead)
	}
	if res.WritePerShard != 500.0/3.0 {
		t.Fatalf("WritePerShard: got %v", res.WritePerShard)
	}
	if !res.HasExpectedNodes {
		t.Fatal("expected successful allocation for default scenario")
	}
	if res.PrimarySlotsFilled != 3 || res.ReplicaSlotsFilled != 6 {
		t.Fatalf("allocation counts: prim %d rep %d", res.PrimarySlotsFilled, res.ReplicaSlotsFilled)
	}
}

func TestCompute_clustersMultiplier(t *testing.T) {
	in := Inputs{
		Nodes: 1, Shards: 1, ReplicasPerShard: 0, GBSize: 10, Clusters: 2,
	}
	res := Compute(in)
	// single cluster 10 GiB primary * (1+0) copies = 10 * 1e9 bytes; * 2 clusters
	want := int64(20 * 1000 * 1000000)
	if res.ClusterBytes != want {
		t.Fatalf("got %d want %d", res.ClusterBytes, want)
	}
}

func TestShardSummaries_ceil(t *testing.T) {
	in := Inputs{
		Shards: 3, ReplicasPerShard: 2, WriteRPM: 500, ReadRPM: 3000, GBSize: 90, Documents: 10_000_000,
	}
	rows := ShardSummaries(in)
	if len(rows) != 3 {
		t.Fatalf("len %d", len(rows))
	}
	if rows[0].WriteRPM != math.Ceil(500.0/3.0) {
		t.Fatalf("WriteRPM ceil: got %v", rows[0].WriteRPM)
	}
	wantRead := math.Ceil(3000.0 / 9.0)
	if rows[0].ReadRPM != wantRead {
		t.Fatalf("ReadRPM: got %v want %v", rows[0].ReadRPM, wantRead)
	}
}

func TestMessages_needsShard(t *testing.T) {
	res := Compute(Inputs{Nodes: 3, Shards: 0, ReplicasPerShard: 1})
	if len(res.Messages) == 0 {
		t.Fatal("expected message for zero shards")
	}
}

func TestReplicaAllocation_noSameShardOnNode(t *testing.T) {
	res := Compute(Inputs{
		Nodes: 3, DedicatedMasters: 0, Shards: 3, ReplicasPerShard: 2, GBSize: 90,
	})
	for node, pieces := range res.Allocation {
		seen := map[string]bool{}
		for _, p := range pieces {
			key := p.Shard
			if seen[key] {
				t.Fatalf("node %d has duplicate shard %s", node, key)
			}
			seen[key] = true
		}
	}
}
