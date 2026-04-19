// Package calculator ports the sizing logic from gbaptista/elastic-calculator (Vue).
package calculator

import (
	"math"
	"strconv"
)

// Inputs mirrors elastic-calculator home.js / Cluster.vue props.
type Inputs struct {
	Nodes            int
	DedicatedMasters int
	Shards           int
	ReplicasPerShard int // R in "number_of_replicas"
	GBSize           int // total primary data size across all primary shards (GiB scale per original UI)
	Documents        int64
	ReadRPM          float64
	WriteRPM         float64
	Clusters         int // display / scenario count; storage summary can multiply when >1
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
	SizeWarningHighDanger SizeWarning = "high-danger" // > 32 GiB / primary
	SizeWarningHighWarn   SizeWarning = "high-warning"
	SizeWarningLowDanger  SizeWarning = "low-danger"
	SizeWarningLowWarn    SizeWarning = "low-warning"
)

// Result holds derived values for the TUI and tests.
type Result struct {
	GBPerPrimaryShard float64
	ClusterBytes      int64 // total stored size including replica copies (same formula as Vue clusterSize)

	ReadPerPiece  float64 // rpm per primary or replica piece
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

// Compute runs the same formulas as the reference Vue app.
func Compute(in Inputs) Result {
	if in.Clusters < 1 {
		in.Clusters = 1
	}

	out := Result{
		Allocation: map[int][]Piece{},
		Messages:   nil,
	}

	dataNodes := in.Nodes - in.DedicatedMasters
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

	// Cluster.vue clusterSize: gbPerShard * (shards + shards*replicas) * 1000 * 1000000
	clusterGiB := gbPer * float64(in.Shards+in.Shards*in.ReplicasPerShard)
	out.ClusterBytes = int64(clusterGiB * 1000 * 1000000)

	out.ReadPerPiece = 0
	denom := float64(in.Shards*in.ReplicasPerShard + in.Shards)
	if denom > 0 && in.Shards > 0 {
		out.ReadPerPiece = in.ReadRPM / denom
	}
	if in.Shards > 0 {
		out.WritePerShard = in.WriteRPM / float64(in.Shards)
	}

	out.SizeWarning = sizeWarning(gbPer, in.Shards)

	out.Allocation, out.PrimarySlotsFilled, out.ReplicaSlotsFilled = allocatePieces(
		dataNodes, in.Shards, in.ReplicasPerShard,
	)

	expectedReplicas := in.Shards * in.ReplicasPerShard
	replicasOK := out.ReplicaSlotsFilled == expectedReplicas && out.PrimarySlotsFilled == in.Shards
	nodesOK := in.Nodes >= out.ExpectedMinNodes && replicasOK
	out.HasExpectedNodes = nodesOK

	// Optional: multiple independent clusters (home.js `clusters`) multiply total stored size.
	if in.Clusters > 1 {
		out.ClusterBytes *= int64(in.Clusters)
	}

	out.Messages = buildMessages(in, out)

	return out
}

func sizeWarning(gbPerShard float64, shards int) SizeWarning {
	if gbPerShard > 32 {
		return SizeWarningHighDanger
	}
	if gbPerShard > 28 {
		return SizeWarningHighWarn
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
	if !out.HasExpectedNodes && in.Shards > 0 {
		msgs = append(msgs, "The cluster needs more nodes!")
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
		readPerPiece = in.ReadRPM / denom
	}
	writePerShard := in.WriteRPM / float64(in.Shards)
	docsPerShard := float64(in.Documents) / float64(in.Shards)
	gbPerShard := float64(in.GBSize) / float64(in.Shards)

	var rows []NodeSummary
	maxN := in.Nodes - in.DedicatedMasters
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
			ReadRPM:   float64(pieceCount) * readPerPiece,
			WriteRPM:  float64(prim) * writePerShard,
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
	ReadRPM   float64
	WriteRPM  float64
	Docs      float64
	BytesAll  int64
	BytesPrim int64
}

// ShardSummaries returns per-primary-shard metrics (Shard.vue).
func ShardSummaries(in Inputs) []ShardSummary {
	if in.Shards == 0 {
		return nil
	}
	r := in.ReplicasPerShard
	denom := float64(in.Shards*r + in.Shards)
	readEach := 0.0
	if denom > 0 {
		readEach = in.ReadRPM / denom
	}
	writeEach := in.WriteRPM / float64(in.Shards)
	writeEach = math.Ceil(writeEach)
	if readEach > 0 {
		readEach = math.Ceil(readEach)
	}
	gbPer := float64(in.GBSize) / float64(in.Shards)
	docEach := float64(in.Documents) / float64(in.Shards)

	out := make([]ShardSummary, in.Shards)
	for i := 0; i < in.Shards; i++ {
		out[i] = ShardSummary{
			Index:     i + 1,
			ReadRPM:   readEach,
			WriteRPM:  writeEach,
			Docs:      docEach,
			Bytes:     int64(gbPer * 1000 * 1000000),
			GBPerPrim: gbPer,
			Warning:   sizeWarning(gbPer, in.Shards),
		}
	}
	return out
}

// ShardSummary is one primary shard row (Shard.vue).
type ShardSummary struct {
	Index     int
	ReadRPM   float64
	WriteRPM  float64
	Docs      float64
	Bytes     int64
	GBPerPrim float64
	Warning   SizeWarning
}
