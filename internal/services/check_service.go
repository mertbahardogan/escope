package services

import (
	"context"
	"fmt"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/util"
	"time"
)

type CheckService interface {
	GetClusterHealthCheck(ctx context.Context) (*models.ClusterInfo, error)
	GetNodeHealthCheck(ctx context.Context) ([]models.CheckNodeHealth, error)
	GetShardHealthCheck(ctx context.Context) (*models.ShardHealth, error)
	GetShardWarningsCheck(ctx context.Context) (*models.ShardWarnings, error)
	GetIndexHealthCheck(ctx context.Context) ([]models.IndexHealth, error)
	GetResourceUsageCheck(ctx context.Context) (*models.ResourceUsage, error)
	GetPerformanceCheck(ctx context.Context) (*models.Performance, error)
	GetNodeBreakdown(ctx context.Context) (*models.NodeBreakdown, error)
	GetSegmentWarningsCheck(ctx context.Context) (*models.SegmentWarnings, error)
}

type checkService struct {
	client          interfaces.ElasticClient
	nodeService     NodeService
	segmentsService SegmentsService
}

func NewCheckService(client interfaces.ElasticClient) CheckService {
	return &checkService{
		client:          client,
		nodeService:     NewNodeService(client),
		segmentsService: NewSegmentsService(client),
	}
}

func (s *checkService) GetClusterHealthCheck(ctx context.Context) (*models.ClusterInfo, error) {
	healthData, err := s.client.GetClusterHealth(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrClusterHealthRequestFailed, err)
	}

	health := &models.ClusterInfo{
		Timestamp: time.Now(),
	}

	if clusterName, ok := healthData[constants.ClusterNameField].(string); ok {
		health.ClusterName = clusterName
	}

	if status, ok := healthData[constants.StatusField].(string); ok {
		health.Status = status
	}

	if numberOfNodes, ok := healthData[constants.NumberOfNodesField].(float64); ok {
		health.NumberOfNodes = int(numberOfNodes)
	}

	if activePrimaryShards, ok := healthData[constants.ActivePrimaryShardsField].(float64); ok {
		health.ActivePrimaryShards = int(activePrimaryShards)
	}

	if activeShards, ok := healthData[constants.ActiveShardsField].(float64); ok {
		health.ActiveShards = int(activeShards)
	}

	if unassignedShards, ok := healthData[constants.UnassignedShardsField].(float64); ok {
		health.UnassignedShards = int(unassignedShards)
	}

	if relocatingShards, ok := healthData[constants.RelocatingShardsField].(float64); ok {
		health.RelocatingShards = int(relocatingShards)
	}

	if initializingShards, ok := healthData[constants.InitializingShardsField].(float64); ok {
		health.InitializingShards = int(initializingShards)
	}

	return health, nil
}

func (s *checkService) GetNodeHealthCheck(ctx context.Context) ([]models.CheckNodeHealth, error) {
	nodesData, err := s.client.GetNodesStats(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrNodesStatsRequestFailed, err)
	}

	var nodeHealths []models.CheckNodeHealth
	if nodes, ok := nodesData[constants.NodesField].(map[string]interface{}); ok {
		for nodeID, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				health := parseNodeHealth(nodeID, node)
				nodeHealths = append(nodeHealths, health)
			}
		}
	}

	return nodeHealths, nil
}

func (s *checkService) GetShardHealthCheck(ctx context.Context) (*models.ShardHealth, error) {
	shardsData, err := s.client.GetShards(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrShardsRequestFailed, err)
	}

	health := &models.ShardHealth{
		Timestamp: time.Now(),
	}

	if shardsList, ok := shardsData[constants.EmptyString].([]interface{}); ok {
		for _, shardData := range shardsList {
			if shard, ok := shardData.(map[string]interface{}); ok {
				state := util.GetStringField(shard, constants.StateField)
				switch state {
				case constants.ShardStateStarted:
					health.StartedShards++
				case constants.ShardStateInitializing:
					health.InitializingShards++
				case constants.ShardStateRelocating:
					health.RelocatingShards++
				case constants.ShardStateUnassigned:
					health.UnassignedShards++
				}
			}
		}
	}

	return health, nil
}

func (s *checkService) GetShardWarningsCheck(ctx context.Context) (*models.ShardWarnings, error) {
	shardsData, err := s.client.GetShards(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFailedToGetShardInfo, err)
	}

	var shardsList []map[string]interface{}
	if shards, ok := shardsData[constants.EmptyString].([]map[string]interface{}); ok {
		shardsList = shards
	}

	warnings := &models.ShardWarnings{
		Recommendations: make([]string, 0),
		CriticalIssues:  make([]string, 0),
		WarningIssues:   make([]string, 0),
	}

	nodeShardCounts := make(map[string]int)
	for _, shard := range shardsList {
		state := util.GetStringField(shard, constants.StateField)
		node := util.GetStringField(shard, constants.NodeFieldKey)
		ip := util.GetStringField(shard, constants.IPFieldKey)

		switch state {
		case constants.ShardStateUnassigned:
			warnings.UnassignedShards++
		case constants.ShardStateRelocating:
			warnings.RelocatingShards++
		case constants.ShardStateInitializing:
			warnings.InitializingShards++
		case constants.ShardStateStarted:
			if node != constants.EmptyString && node != constants.DashString {
				nodeShardCounts[node]++
			} else if ip != constants.EmptyString && ip != constants.DashString {
				nodeShardCounts[ip]++
			}
		}
	}

	if warnings.UnassignedShards > 0 {
		warnings.CriticalIssues = append(warnings.CriticalIssues,
			fmt.Sprintf(constants.MsgUnassignedShards, warnings.UnassignedShards))
		warnings.Recommendations = append(warnings.Recommendations,
			fmt.Sprintf(constants.MsgInvestigateUnassigned, warnings.UnassignedShards))
	}

	if warnings.RelocatingShards > 0 {
		warnings.WarningIssues = append(warnings.WarningIssues,
			fmt.Sprintf(constants.MsgRelocatingShards, warnings.RelocatingShards))
	}

	if warnings.InitializingShards > 0 {
		warnings.WarningIssues = append(warnings.WarningIssues,
			fmt.Sprintf(constants.MsgInitializingShards, warnings.InitializingShards))
	}

	if len(nodeShardCounts) > 1 {
		var counts []int
		for _, count := range nodeShardCounts {
			counts = append(counts, count)
		}

		minShards := counts[0]
		maxShards := counts[0]
		for _, count := range counts {
			if count < minShards {
				minShards = count
			}
			if count > maxShards {
				maxShards = count
			}
		}

		if maxShards > 0 {
			warnings.UnbalancedRatio = float64(minShards) / float64(maxShards)

			if warnings.UnbalancedRatio < constants.BalanceRatioThreshold {
				warnings.UnbalancedShards = true
				warnings.WarningIssues = append(warnings.WarningIssues,
					fmt.Sprintf(constants.MsgShardUnbalanced, warnings.UnbalancedRatio))
				warnings.Recommendations = append(warnings.Recommendations,
					constants.MsgConsiderRebalancing)
			}
		}
	}

	if len(warnings.CriticalIssues) == 0 && len(warnings.WarningIssues) == 0 {
		warnings.Recommendations = append(warnings.Recommendations, constants.MsgShardHealthy)
	}

	return warnings, nil
}

func (s *checkService) GetIndexHealthCheck(ctx context.Context) ([]models.IndexHealth, error) {
	indicesData, err := s.client.GetIndices(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndicesRequestFailed, err)
	}

	var indexHealths []models.IndexHealth
	if indicesList, ok := indicesData[constants.EmptyString].([]interface{}); ok {
		for _, indexData := range indicesList {
			if index, ok := indexData.(map[string]interface{}); ok {
				health := models.IndexHealth{
					Timestamp: time.Now(),
					Name:      util.GetStringField(index, constants.IndexField),
					Health:    util.GetStringField(index, constants.HealthField),
					Status:    util.GetStringField(index, constants.StatusField),
					Docs:      util.GetStringField(index, constants.DocsCountField),
					Size:      util.GetStringField(index, constants.StoreSizeField),
				}
				indexHealths = append(indexHealths, health)
			}
		}
	}

	return indexHealths, nil
}

func (s *checkService) GetResourceUsageCheck(ctx context.Context) (*models.ResourceUsage, error) {
	nodesData, err := s.client.GetNodesStats(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrNodesStatsRequestFailed, err)
	}

	usage := &models.ResourceUsage{
		Timestamp: time.Now(),
	}

	var cpuValues []float64
	var heapValues []float64

	if nodes, ok := nodesData[constants.NodesField].(map[string]interface{}); ok {
		for _, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				if os, ok := node[constants.OSField].(map[string]interface{}); ok {
					if cpu, ok := os[constants.CPUField].(map[string]interface{}); ok {
						if percent, ok := cpu[constants.CPUPercentField].(float64); ok {
							usage.CPUUsage += percent
							cpuValues = append(cpuValues, percent)
						}
					}
				}

				if jvm, ok := node[constants.JVMField].(map[string]interface{}); ok {
					if mem, ok := jvm[constants.JVMMemField].(map[string]interface{}); ok {
						if heapUsed, ok := mem[constants.HeapUsedPctField].(float64); ok {
							usage.HeapUsage += heapUsed
							heapValues = append(heapValues, heapUsed)
						}
					}
				}

				if fs, ok := node[constants.FSField].(map[string]interface{}); ok {
					if total, ok := fs[constants.TotalField].(map[string]interface{}); ok {
						if totalBytes, ok := total[constants.TotalInBytesField].(float64); ok {
							usage.DiskTotal += int64(totalBytes)
						}
						if availableBytes, ok := total[constants.AvailableInBytesField].(float64); ok {
							usage.DiskAvailable += int64(availableBytes)
						}
					}
				}

				usage.NodeCount++
			}
		}
	}

	if usage.NodeCount > 0 {
		usage.CPUUsage /= float64(usage.NodeCount)
		usage.HeapUsage /= float64(usage.NodeCount)
		if len(cpuValues) > 0 {
			usage.CPUUsageMin = cpuValues[0]
			usage.CPUUsageMax = cpuValues[0]
			for _, val := range cpuValues {
				if val < usage.CPUUsageMin {
					usage.CPUUsageMin = val
				}
				if val > usage.CPUUsageMax {
					usage.CPUUsageMax = val
				}
			}
		}
		if len(heapValues) > 0 {
			usage.HeapUsageMin = heapValues[0]
			usage.HeapUsageMax = heapValues[0]
			for _, val := range heapValues {
				if val < usage.HeapUsageMin {
					usage.HeapUsageMin = val
				}
				if val > usage.HeapUsageMax {
					usage.HeapUsageMax = val
				}
			}
		}
	}

	return usage, nil
}

func (s *checkService) GetPerformanceCheck(ctx context.Context) (*models.Performance, error) {
	clusterData, err := s.client.GetClusterStats(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrClusterStatsRequestFailed, err)
	}

	performance := &models.Performance{
		Timestamp: time.Now(),
	}

	if indices, ok := clusterData[constants.IndicesField].(map[string]interface{}); ok {
		if indexing, ok := indices[constants.IndexingField].(map[string]interface{}); ok {
			if indexTotal, ok := indexing[constants.IndexTotalField].(float64); ok {
				performance.IndexTotal = int64(indexTotal)
			}
			if indexTime, ok := indexing[constants.IndexTimeInMillisField].(float64); ok {
				performance.IndexTimeInMillis = int64(indexTime)
			}
		}

		if search, ok := indices[constants.SearchField].(map[string]interface{}); ok {
			if queryTotal, ok := search[constants.QueryTotalField].(float64); ok {
				performance.QueryTotal = int64(queryTotal)
			}
			if queryTime, ok := search[constants.QueryTimeInMillisField].(float64); ok {
				performance.QueryTimeInMillis = int64(queryTime)
			}
		}
	}

	return performance, nil
}

func parseNodeHealth(nodeID string, node map[string]interface{}) models.CheckNodeHealth {
	health := models.CheckNodeHealth{
		Timestamp: time.Now(),
		NodeID:    nodeID,
	}

	if name, ok := node[constants.NameField].(string); ok {
		health.Name = name
	}

	if os, ok := node[constants.OSField].(map[string]interface{}); ok {
		if cpu, ok := os[constants.CPUField].(map[string]interface{}); ok {
			if percent, ok := cpu[constants.PercentField].(float64); ok {
				health.CPUUsage = percent
			}
		}
	}

	if jvm, ok := node[constants.JVMField].(map[string]interface{}); ok {
		if mem, ok := jvm[constants.MemField].(map[string]interface{}); ok {
			if heapUsed, ok := mem[constants.HeapUsedPercentField].(float64); ok {
				health.HeapUsage = heapUsed
			}
		}
	}

	return health
}

func (s *checkService) GetNodeBreakdown(ctx context.Context) (*models.NodeBreakdown, error) {
	return s.nodeService.GetNodeBreakdown(ctx)
}

func (s *checkService) GetSegmentWarningsCheck(ctx context.Context) (*models.SegmentWarnings, error) {
	segments, err := s.segmentsService.GetSegmentsInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFailedToGetSegmentsInfo, err)
	}

	thresholds, err := config.GetThresholdsForCluster(ctx, s.client)
	if err != nil {
		thresholds = &config.DynamicThresholds{
			HighSegmentThreshold:  constants.HighSegmentThreshold,
			SmallSegmentThreshold: constants.SmallSegmentThreshold,
			LargeSegmentThreshold: constants.LargeSegmentThreshold,
		}
	}

	warnings := &models.SegmentWarnings{}

	for _, seg := range segments {
		if util.IsSystemIndex(seg.Index) {
			continue
		}

		if seg.SegmentCount > thresholds.HighSegmentThreshold {
			warnings.HighSegmentIndices++
		}

		avgMemPerSeg := int64(0)
		if seg.SegmentCount > 0 {
			avgMemPerSeg = seg.SizeBytes / int64(seg.SegmentCount)
		}

		if avgMemPerSeg < thresholds.SmallSegmentThreshold {
			warnings.SmallSegmentIndices++
		}
		if avgMemPerSeg > thresholds.LargeSegmentThreshold {
			warnings.LargeSegmentIndices++
		}
	}

	return warnings, nil
}
