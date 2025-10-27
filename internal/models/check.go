package models

import "time"

type CheckNodeHealth struct {
	Timestamp time.Time
	NodeID    string
	Name      string
	CPUUsage  float64
	HeapUsage float64
}

type ShardHealth struct {
	Timestamp          time.Time
	StartedShards      int
	InitializingShards int
	RelocatingShards   int
	UnassignedShards   int
}

type IndexHealth struct {
	Timestamp time.Time
	Name      string
	Health    string
	Status    string
	Docs      string
	Size      string
}

type ResourceUsage struct {
	Timestamp        time.Time
	NodeCount        int
	CPUUsage         float64
	CPUUsageMin      float64
	CPUUsageMax      float64
	CPUUsageMinNode  string
	CPUUsageMaxNode  string
	HeapUsage        float64
	HeapUsageMin     float64
	HeapUsageMax     float64
	HeapUsageMinNode string
	HeapUsageMaxNode string
	DiskTotal        int64
	DiskAvailable    int64
}

type Performance struct {
	Timestamp         time.Time
	IndexTotal        int64
	IndexTimeInMillis int64
	QueryTotal        int64
	QueryTimeInMillis int64
}

type SegmentWarnings struct {
	HighSegmentIndices  int
	SmallSegmentIndices int
	LargeSegmentIndices int
}
