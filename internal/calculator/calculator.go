package calculator

import (
	"math"
	"strconv"
)

type Inputs struct {
	DataNodes          int
	DedicatedMasters   int
	Shards             int
	ReplicasPerShard   int // R in "number_of_replicas"
	GBSize             int // total primary data size across all primary shards (GiB scale per original UI)
	Documents          int64
	ReadRPS            float64
	WriteRPS           float64
	RAMGiBPerDataNode  float64
	DiskGiBPerDataNode float64
}

func (in Inputs) TotalNodes() int {
	return in.DataNodes + in.DedicatedMasters
}

// PieceKind identifies a primary shard or a replica copy on a data node.
type PieceKind string

const (
	KindShard   PieceKind = "shard"
	KindReplica PieceKind = "replica"
)

// Piece is one primary shard or one replica assignment.
type Piece struct {
	Key   string
	Name  string
	Kind  PieceKind
	Shard string // e.g. "shard-1"
}

// SizeWarning matches Shard.vue sizeWarning categories.
type SizeWarning string

const (
	SizeWarningNone       SizeWarning = ""
	SizeWarningHighDanger SizeWarning = "high-danger"
	SizeWarningLowDanger  SizeWarning = "low-danger"
	SizeWarningLowWarn    SizeWarning = "low-warning"
)

type Result struct {
	GBPerPrimaryShard float64
	ClusterBytes      int64 // total stored size including replica copies (same formula as Vue clusterSize)

	ReadPerPiece  float64 // RPS per primary or replica piece
	WritePerShard float64

	SizeWarning SizeWarning

	DataNodes          int
	Allocation         map[int][]Piece // 1-based data node index -> pieces in allocation order
	ReplicaSlotsFilled int
	PrimarySlotsFilled int

	ExpectedMinNodes int
	HasExpectedNodes bool

	Messages []string
}

func Compute(in Inputs) Result {
	out := Result{
		Allocation: map[int][]Piece{},
		Messages:   nil,
	}

	dataNodes := in.DataNodes
	if dataNodes < 0 {
		dataNodes = 0
	}
	out.DataNodes = dataNodes

	out.ExpectedMinNodes = in.DedicatedMasters
	if in.Shards > 0 {
		out.ExpectedMinNodes++
	}

	gbPer := 0.0
	if in.Shards > 0 {
		gbPer = float64(in.GBSize) / float64(in.Shards)
	}
	out.GBPerPrimaryShard = gbPer

	clusterGiB := gbPer * float64(in.Shards+in.Shards*in.ReplicasPerShard)
	out.ClusterBytes = int64(clusterGiB * 1000 * 1000000)

	out.ReadPerPiece = 0
	denom := float64(in.Shards*in.ReplicasPerShard + in.Shards)
	if denom > 0 && in.Shards > 0 {
		out.ReadPerPiece = in.ReadRPS / denom
	}
	if in.Shards > 0 {
		out.WritePerShard = in.WriteRPS / float64(in.Shards)
	}

	out.SizeWarning = sizeWarning(gbPer, in.Shards)

	out.Allocation, out.PrimarySlotsFilled, out.ReplicaSlotsFilled = allocatePieces(
		dataNodes, in.Shards, in.ReplicasPerShard,
	)

	expectedReplicas := in.Shards * in.ReplicasPerShard
	replicasOK := out.ReplicaSlotsFilled == expectedReplicas && out.PrimarySlotsFilled == in.Shards
	nodesOK := in.TotalNodes() >= out.ExpectedMinNodes && replicasOK
	out.HasExpectedNodes = nodesOK

	out.Messages = buildMessages(in, out)

	return out
}

func sizeWarning(gbPerShard float64, shards int) SizeWarning {
	if gbPerShard > 50 {
		return SizeWarningHighDanger
	}
	if shards < 2 {
		return SizeWarningNone
	}
	if gbPerShard < 8 {
		return SizeWarningLowDanger
	}
	if gbPerShard < 13 {
		return SizeWarningLowWarn
	}
	return SizeWarningNone
}

func allocatePieces(dataNodes, shards, replicasPer int) (map[int][]Piece, int, int) {
	m := map[int][]Piece{}
	for n := 1; n <= dataNodes; n++ {
		m[n] = []Piece{}
	}

	if dataNodes == 0 || shards == 0 {
		return m, 0, 0
	}

	pendingShards := shards
	pendingReplicas := shards * replicasPer

	type repJob struct {
		key   string
		name  string
		shard string
	}
	var repQueue []repJob
	for s := 1; s <= shards; s++ {
		for r := 1; r <= replicasPer; r++ {
			repQueue = append(repQueue, repJob{
				key:   "replica-" + strconv.Itoa(r) + "-for-shard-" + strconv.Itoa(s),
				name:  "replica (shard " + strconv.Itoa(s) + ")",
				shard: "shard-" + strconv.Itoa(s),
			})
		}
	}

	maxTries := (pendingShards*replicasPer + pendingShards) * 3
	if maxTries < 1 {
		maxTries = 1
	}

	for maxTries > 0 && (pendingShards > 0 || pendingReplicas > 0) {
		maxTries--
		for node := 1; node <= dataNodes; node++ {
			if pendingShards > 0 {
				idx := shards - pendingShards + 1
				m[node] = append(m[node], Piece{
					Key:   "shard-" + strconv.Itoa(idx),
					Name:  "shard " + strconv.Itoa(idx),
					Kind:  KindShard,
					Shard: "shard-" + strconv.Itoa(idx),
				})
				pendingShards--
			} else if pendingReplicas > 0 {
				placed := false
				for r := 0; r < len(repQueue); r++ {
					job := repQueue[r]
					if canAllocateReplica(m[node], job.key, job.shard) {
						m[node] = append(m[node], Piece{
							Key:   job.key,
							Name:  job.name,
							Kind:  KindReplica,
							Shard: job.shard,
						})
						repQueue = append(repQueue[:r], repQueue[r+1:]...)
						pendingReplicas--
						placed = true
						break
					}
				}
				_ = placed
			} else {
				break
			}
		}
	}

	prim := 0
	repl := 0
	for _, pieces := range m {
		for _, p := range pieces {
			switch p.Kind {
			case KindShard:
				prim++
			case KindReplica:
				repl++
			}
		}
	}
	return m, prim, repl
}

func canAllocateReplica(current []Piece, newKey, newShard string) bool {
	for _, p := range current {
		if p.Shard == newShard {
			return false
		}
	}
	for _, p := range current {
		if p.Key == newKey {
			return false
		}
	}
	return true
}

func buildMessages(in Inputs, out Result) []string {
	var msgs []string
	if in.Shards == 0 {
		msgs = append(msgs, "The cluster needs at least one shard.")
	}
	if out.DataNodes == 0 && in.Shards > 0 {
		msgs = append(msgs, "The cluster has not being used!")
	}
	return msgs
}

// NodeSummaries returns read/write/docs/bytes totals per data node (Node.vue computed).
func NodeSummaries(in Inputs, alloc map[int][]Piece) []NodeSummary {
	if in.Shards == 0 {
		return nil
	}
	r := in.ReplicasPerShard
	denom := float64(in.Shards*r + in.Shards)
	readPerPiece := 0.0
	if denom > 0 {
		readPerPiece = in.ReadRPS / denom
	}
	writePerShard := in.WriteRPS / float64(in.Shards)
	docsPerShard := float64(in.Documents) / float64(in.Shards)
	gbPerShard := float64(in.GBSize) / float64(in.Shards)

	var rows []NodeSummary
	maxN := in.DataNodes
	for node := 1; node <= maxN; node++ {
		pieces := alloc[node]
		var prim, rep int
		for _, p := range pieces {
			switch p.Kind {
			case KindShard:
				prim++
			case KindReplica:
				rep++
			}
		}
		pieceCount := prim + rep
		sum := NodeSummary{
			NodeIndex: node,
			Primaries: prim,
			Replicas:  rep,
			ReadRPS:   float64(pieceCount) * readPerPiece,
			WriteRPS:  float64(prim) * writePerShard,
			Docs:      float64(prim) * docsPerShard,
			BytesAll:  int64(float64(pieceCount) * gbPerShard * 1000 * 1000000),
			BytesPrim: int64(float64(prim) * gbPerShard * 1000 * 1000000),
		}
		rows = append(rows, sum)
	}
	return rows
}

// NodeSummary is one row of per-node rollup (aligned with Node.vue).
type NodeSummary struct {
	NodeIndex int
	Primaries int
	Replicas  int
	ReadRPS   float64
	WriteRPS  float64
	Docs      float64
	BytesAll  int64
	BytesPrim int64
}

type NodeResourceView struct {
	NodeIndex int

	RAMGiB         float64
	HeapCapGiB     float64
	OSPageCacheGiB float64 // RAM minus JVM heap cap (OS page cache budget)

	DataGiB    float64
	DiskCapGiB float64
	DiskUsePct float64

	// JVM heap sub-segments (sum to HeapCapGiB).
	FieldDataGiB   float64
	QueryBufferGiB float64
	IndexBufferGiB float64
	HeapAvailGiB   float64

	// PageCacheCoversHotPct: min(100, OSPageCacheGiB/dataGiB*100) — share of on-node data that fits OS cache.
	PageCacheCoversHotPct float64
	HeapOfRAMPct          float64
}

func NodeResourceViews(in Inputs, rows []NodeSummary) []NodeResourceView {
	if len(rows) == 0 {
		return nil
	}
	ram := in.RAMGiBPerDataNode
	if ram < 1 {
		ram = 1
	}
	diskCap := in.DiskGiBPerDataNode
	if diskCap < 1 {
		diskCap = 1
	}
	heapCap := math.Min(ram*0.5, 31)
	osPage := math.Max(0, ram-heapCap)

	out := make([]NodeResourceView, 0, len(rows))
	for _, row := range rows {
		dataGiB := float64(row.BytesAll) / 1e9
		diskPct := math.Min(100, dataGiB/diskCap*100)

		coverPct := 100.0
		if dataGiB > 1e-9 {
			coverPct = math.Min(100, osPage/dataGiB*100)
		}

		fd := 0.15 * heapCap
		qb := 0.10 * heapCap
		ib := 0.10 * heapCap
		avail := heapCap - fd - qb - ib
		if avail < 0 {
			avail = 0
		}

		out = append(out, NodeResourceView{
			NodeIndex: row.NodeIndex,

			RAMGiB:                ram,
			HeapCapGiB:            heapCap,
			OSPageCacheGiB:        osPage,
			DataGiB:               dataGiB,
			DiskCapGiB:            diskCap,
			DiskUsePct:            diskPct,
			FieldDataGiB:          fd,
			QueryBufferGiB:        qb,
			IndexBufferGiB:        ib,
			HeapAvailGiB:          avail,
			PageCacheCoversHotPct: coverPct,
			HeapOfRAMPct:          heapCap / ram * 100,
		})
	}
	return out
}
