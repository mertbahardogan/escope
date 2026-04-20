package calculator

import (
	"math"
	"testing"
)

func TestCompute_defaultHomeJS(t *testing.T) {
	const writeTotalRPS = 500.0 / 60.0
	in := Inputs{
		DataNodes:          3,
		DedicatedMasters:   0,
		Shards:             3,
		ReplicasPerShard:   2,
		GBSize:             90,
		Documents:          10_000_000,
		ReadRPS:            3000.0 / 60.0,
		WriteRPS:           writeTotalRPS,
		RAMGiBPerDataNode:  64,
		DiskGiBPerDataNode: 2000,
	}
	res := Compute(in)

	wantBytes := int64(270 * 1000 * 1000000)
	if res.ClusterBytes != wantBytes {
		t.Fatalf("ClusterBytes: got %d want %d", res.ClusterBytes, wantBytes)
	}
	if math.Abs(res.GBPerPrimaryShard-30) > 1e-9 {
		t.Fatalf("GBPerPrimaryShard: got %v want 30", res.GBPerPrimaryShard)
	}
	if res.SizeWarning != SizeWarningNone {
		t.Fatalf("SizeWarning: got %s want %s", res.SizeWarning, SizeWarningNone)
	}
	denom := float64(3*2 + 3)
	wantRead := (3000.0 / 60.0) / denom
	if math.Abs(res.ReadPerPiece-wantRead) > 1e-9 {
		t.Fatalf("ReadPerPiece: got %v want %v", res.ReadPerPiece, wantRead)
	}
	if math.Abs(res.WritePerShard-writeTotalRPS/3.0) > 1e-9 {
		t.Fatalf("WritePerShard: got %v", res.WritePerShard)
	}
	if !res.HasExpectedNodes {
		t.Fatal("expected successful allocation for default scenario")
	}
	if res.PrimarySlotsFilled != 3 || res.ReplicaSlotsFilled != 6 {
		t.Fatalf("allocation counts: prim %d rep %d", res.PrimarySlotsFilled, res.ReplicaSlotsFilled)
	}
}

func TestMessages_needsShard(t *testing.T) {
	res := Compute(Inputs{DataNodes: 3, Shards: 0, ReplicasPerShard: 1})
	if len(res.Messages) == 0 {
		t.Fatal("expected message for zero shards")
	}
}

func TestReplicaAllocation_noSameShardOnNode(t *testing.T) {
	res := Compute(Inputs{
		DataNodes: 3, DedicatedMasters: 0, Shards: 3, ReplicasPerShard: 2, GBSize: 90,
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

func TestNodeResourceViews_heapDetailAndCover(t *testing.T) {
	in := Inputs{
		DataNodes:          2,
		Shards:             2,
		ReplicasPerShard:   1,
		GBSize:             40,
		Documents:          100,
		ReadRPS:            10,
		WriteRPS:           10,
		RAMGiBPerDataNode:  64,
		DiskGiBPerDataNode: 500,
	}
	res := Compute(in)
	rows := NodeSummaries(in, res.Allocation)
	views := NodeResourceViews(in, rows)
	if len(views) != 2 {
		t.Fatalf("views len %d", len(views))
	}
	for _, v := range views {
		if v.DiskUsePct < 0 || v.DiskUsePct > 100 {
			t.Fatalf("disk pct %v", v.DiskUsePct)
		}
		sum := v.FieldDataGiB + v.QueryBufferGiB + v.IndexBufferGiB + v.HeapAvailGiB
		if math.Abs(sum-v.HeapCapGiB) > 1e-6 {
			t.Fatalf("heap detail sum %v vs cap %v", sum, v.HeapCapGiB)
		}
		if v.PageCacheCoversHotPct < 0 || v.PageCacheCoversHotPct > 100 {
			t.Fatalf("cover pct %v", v.PageCacheCoversHotPct)
		}
	}
}
