package config

import (
	"context"
	"github.com/mertbahardogan/escope/internal/interfaces"
)

type DynamicThresholds struct {
	HighSegmentThreshold  int
	SmallSegmentThreshold int64
	LargeSegmentThreshold int64
	HighCPUThreshold      float64
	HighMemoryThreshold   float64
	HighHeapThreshold     float64
	HighDiskThreshold     float64
}

func CalculateDynamicThresholds(ctx context.Context, client interfaces.ElasticClient) (*DynamicThresholds, error) {
	// Get cluster info to determine size
	clusterInfo, err := client.GetClusterHealth(ctx)
	if err != nil {
		return nil, err
	}

	// Get node count for scaling
	nodeCount := 1 // Default fallback
	if nodes, ok := clusterInfo["number_of_nodes"].(float64); ok {
		nodeCount = int(nodes)
	}
	if nodeCount == 0 {
		nodeCount = 1 // Fallback to prevent division by zero
	}

	// Calculate base thresholds with scaling factors
	thresholds := &DynamicThresholds{
		// Segment thresholds scale with cluster size
		HighSegmentThreshold:  calculateSegmentThreshold(nodeCount),
		SmallSegmentThreshold: 1024 * 1024,        // 1MB (static)
		LargeSegmentThreshold: 1024 * 1024 * 1024, // 1GB (static)

		// Memory thresholds become more lenient with larger clusters
		HighCPUThreshold:    calculateCPUThreshold(nodeCount),
		HighMemoryThreshold: calculateMemoryThreshold(nodeCount),
		HighHeapThreshold:   calculateHeapThreshold(nodeCount),
		HighDiskThreshold:   90.0, // Static for now
	}

	return thresholds, nil
}

// calculateSegmentThreshold scales segment threshold based on cluster size
func calculateSegmentThreshold(nodeCount int) int {
	// Base threshold: 1000 segments
	baseThreshold := 1000

	// Scale factor: more nodes = higher threshold
	// Formula: baseThreshold * (1 + (nodeCount-1) * 0.5)
	// 1 node = 1000, 3 nodes = 2000, 10 nodes = 5500
	scaleFactor := 1.0 + float64(nodeCount-1)*0.5

	return int(float64(baseThreshold) * scaleFactor)
}

// calculateCPUThreshold scales CPU threshold based on cluster size
func calculateCPUThreshold(nodeCount int) float64 {
	// Base threshold: 80%
	baseThreshold := 80.0

	// Larger clusters can handle higher CPU usage
	// Formula: baseThreshold + (nodeCount-1) * 2
	// 1 node = 80%, 3 nodes = 84%, 10 nodes = 98%
	scaledThreshold := baseThreshold + float64(nodeCount-1)*2.0

	// Cap at 95% to prevent unrealistic values
	if scaledThreshold > 95.0 {
		return 95.0
	}

	return scaledThreshold
}

// calculateMemoryThreshold scales memory threshold based on cluster size
func calculateMemoryThreshold(nodeCount int) float64 {
	// Base threshold: 90%
	baseThreshold := 90.0

	// Larger clusters can handle higher memory usage
	// Formula: baseThreshold + (nodeCount-1) * 1
	// 1 node = 90%, 3 nodes = 92%, 10 nodes = 99%
	scaledThreshold := baseThreshold + float64(nodeCount-1)*1.0

	// Cap at 98% to prevent unrealistic values
	if scaledThreshold > 98.0 {
		return 98.0
	}

	return scaledThreshold
}

// calculateHeapThreshold scales heap threshold based on cluster size
func calculateHeapThreshold(nodeCount int) float64 {
	// Base threshold: 85%
	baseThreshold := 85.0

	// Larger clusters can handle higher heap usage
	// Formula: baseThreshold + (nodeCount-1) * 1.5
	// 1 node = 85%, 3 nodes = 88%, 10 nodes = 98.5%
	scaledThreshold := baseThreshold + float64(nodeCount-1)*1.5

	// Cap at 95% to prevent unrealistic values
	if scaledThreshold > 95.0 {
		return 95.0
	}

	return scaledThreshold
}

// GetThresholdsForCluster returns appropriate thresholds for a given cluster
func GetThresholdsForCluster(ctx context.Context, client interfaces.ElasticClient) (*DynamicThresholds, error) {
	return CalculateDynamicThresholds(ctx, client)
}
