package services

import (
	"context"
	"fmt"
	"github.com/mertbahardogan/escope/internal/config"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/util"
	"math"
	"strconv"
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
	GetScaleWarningsCheck(ctx context.Context) (*models.ScaleWarnings, error)
}

type checkService struct {
	client          interfaces.ElasticClient
	nodeService     NodeService
	segmentsService SegmentsService
	indexService    IndexService
}

type indexTrafficRates struct {
	searchRate float64
	indexRate  float64
}

func NewCheckService(client interfaces.ElasticClient) CheckService {
	return &checkService{
		client:          client,
		nodeService:     NewNodeService(client),
		segmentsService: NewSegmentsService(client),
		indexService:    NewIndexService(client),
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

	type nodeMetric struct {
		cpuUsage  float64
		heapUsage float64
		nodeName  string
		nodeIP    string
	}

	var nodeMetrics []nodeMetric

	if nodes, ok := nodesData[constants.NodesField].(map[string]interface{}); ok {
		for _, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				isDataNode := false
				if rolesData, ok := node[constants.RolesField].([]interface{}); ok {
					for _, role := range rolesData {
						if roleStr, ok := role.(string); ok {
							if roleStr == constants.NodeRoleData {
								isDataNode = true
								break
							}
						}
					}
				}

				if !isDataNode {
					continue
				}

				metric := nodeMetric{
					nodeName: util.GetStringField(node, constants.NameField),
					nodeIP:   util.GetStringField(node, constants.IPField),
				}

				if os, ok := node[constants.OSField].(map[string]interface{}); ok {
					if cpu, ok := os[constants.CPUField].(map[string]interface{}); ok {
						if percent, ok := cpu[constants.CPUPercentField].(float64); ok {
							usage.CPUUsage += percent
							metric.cpuUsage = percent
						}
					}
				}

				if jvm, ok := node[constants.JVMField].(map[string]interface{}); ok {
					if mem, ok := jvm[constants.JVMMemField].(map[string]interface{}); ok {
						if heapUsed, ok := mem[constants.HeapUsedPctField].(float64); ok {
							usage.HeapUsage += heapUsed
							metric.heapUsage = heapUsed
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

				nodeMetrics = append(nodeMetrics, metric)
				usage.NodeCount = len(nodeMetrics) // Only count data nodes
			}
		}
	}

	dataNodeCount := len(nodeMetrics)
	if dataNodeCount > 0 {
		usage.CPUUsage /= float64(dataNodeCount)
		usage.HeapUsage /= float64(dataNodeCount)

		// Find min/max CPU nodes
		if len(nodeMetrics) > 0 {
			minCPUNode := nodeMetrics[0]
			maxCPUNode := nodeMetrics[0]
			minHeapNode := nodeMetrics[0]
			maxHeapNode := nodeMetrics[0]

			for _, metric := range nodeMetrics {
				if metric.cpuUsage < minCPUNode.cpuUsage {
					minCPUNode = metric
				}
				if metric.cpuUsage > maxCPUNode.cpuUsage {
					maxCPUNode = metric
				}
				if metric.heapUsage < minHeapNode.heapUsage {
					minHeapNode = metric
				}
				if metric.heapUsage > maxHeapNode.heapUsage {
					maxHeapNode = metric
				}
			}

			usage.CPUUsageMin = minCPUNode.cpuUsage
			usage.CPUUsageMax = maxCPUNode.cpuUsage
			usage.CPUUsageMinNode = fmt.Sprintf("%s - %s", minCPUNode.nodeName, minCPUNode.nodeIP)
			usage.CPUUsageMaxNode = fmt.Sprintf("%s - %s", maxCPUNode.nodeName, maxCPUNode.nodeIP)

			usage.HeapUsageMin = minHeapNode.heapUsage
			usage.HeapUsageMax = maxHeapNode.heapUsage
			usage.HeapUsageMinNode = fmt.Sprintf("%s - %s", minHeapNode.nodeName, minHeapNode.nodeIP)
			usage.HeapUsageMaxNode = fmt.Sprintf("%s - %s", maxHeapNode.nodeName, maxHeapNode.nodeIP)
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

func (s *checkService) GetScaleWarningsCheck(ctx context.Context) (*models.ScaleWarnings, error) {
	indicesData, err := s.client.GetIndices(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndicesRequestFailed, err)
	}

	warnings := &models.ScaleWarnings{
		OverScaledIndices: make([]models.OverScaledIndex, 0),
		WarningIssues:     make([]string, 0),
	}

	allStatsData, err := s.client.GetIndexStats(ctx, constants.EmptyString)
	if err != nil {
		allStatsData = nil
	}

	trafficRatesMap := s.parseAllIndexTrafficRates(allStatsData)

	nodeCount := s.getNodeCount(ctx)

	if indicesList, ok := indicesData[constants.EmptyString].([]map[string]interface{}); ok {
		for _, index := range indicesList {
			if overScaledIdx := s.processIndexForScale(index, trafficRatesMap, allStatsData, nodeCount); overScaledIdx != nil {
				warnings.OverScaledIndices = append(warnings.OverScaledIndices, *overScaledIdx)
				warnings.WarningIssues = append(warnings.WarningIssues, overScaledIdx.WarningMessage)
			}
		}
	} else if indicesList, ok := indicesData[constants.EmptyString].([]interface{}); ok {
		for _, indexData := range indicesList {
			if index, ok := indexData.(map[string]interface{}); ok {
				if overScaledIdx := s.processIndexForScale(index, trafficRatesMap, allStatsData, nodeCount); overScaledIdx != nil {
					warnings.OverScaledIndices = append(warnings.OverScaledIndices, *overScaledIdx)
					warnings.WarningIssues = append(warnings.WarningIssues, overScaledIdx.WarningMessage)
				}
			}
		}
	}

	return warnings, nil
}

// calculateSmartRecommendation uses hybrid scoring to recommend optimal shard count
func (s *checkService) calculateSmartRecommendation(
	indexSize int64,
	trafficRate float64,
	docCount int64,
	nodeCount int,
	currentShards int,
) models.ShardRecommendation {

	// 1. Size-based recommendation (weight: 0.5)
	sizeBasedShards := s.calculateShardsBySize(indexSize, nodeCount)

	// 2. Traffic-based recommendation (weight: 0.3)
	trafficBasedShards := s.calculateShardsByTraffic(trafficRate)

	// 3. Document-based recommendation (weight: 0.2)
	docBasedShards := s.calculateShardsByDocCount(docCount)

	// Check current per-shard size
	// If ≤60GB per shard, don't use size-based recommendation
	sizeWeight := constants.SizeWeight
	trafficWeight := constants.TrafficWeight
	docWeight := constants.DocCountWeight

	if currentShards > 0 && indexSize > 0 {
		indexSizeGB := float64(indexSize) / float64(constants.BytesInGB)
		perShardGB := indexSizeGB / float64(currentShards)

		// If per-shard size is 60GB or less, ignore size-based recommendation
		if perShardGB <= 60.0 {
			sizeWeight = 0.0
			// Redistribute weights: traffic becomes 0.6, doc count becomes 0.4
			trafficWeight = 0.6
			docWeight = 0.4
		}
	}

	// Calculate weighted average
	recommended := int(math.Round(
		float64(sizeBasedShards)*sizeWeight +
			float64(trafficBasedShards)*trafficWeight +
			float64(docBasedShards)*docWeight,
	))

	// Ensure minimum of 1 shard
	if recommended < 1 {
		recommended = 1
	}

	// Don't exceed node count (pointless to have more shards than nodes)
	if recommended > nodeCount && nodeCount > 0 {
		recommended = nodeCount
	}

	// Calculate acceptable range (±40% flexibility)
	minAcceptable := int(math.Ceil(float64(recommended) * (1.0 - constants.AcceptableRangeFlexibility)))
	maxAcceptable := int(math.Floor(float64(recommended) * (1.0 + constants.AcceptableRangeFlexibility)))

	if minAcceptable < 1 {
		minAcceptable = 1
	}

	// Calculate confidence based on data availability
	confidence := s.calculateConfidence(indexSize, trafficRate, docCount)

	// Build reasoning
	reasoning := s.buildRecommendationReasoning(
		sizeBasedShards,
		trafficBasedShards,
		docBasedShards,
		indexSize,
		trafficRate,
		docCount,
	)

	return models.ShardRecommendation{
		Recommended:   recommended,
		MinAcceptable: minAcceptable,
		MaxAcceptable: maxAcceptable,
		Confidence:    confidence,
		Reasoning:     reasoning,
	}
}

// calculateShardsBySize recommends shards based on index size
// Only recommends changes if per-shard size exceeds 60GB
func (s *checkService) calculateShardsBySize(indexSize int64, nodeCount int) int {
	if indexSize == 0 {
		return 1
	}

	indexSizeGB := float64(indexSize) / float64(constants.BytesInGB)

	// Target shard size: 60GB maximum per shard
	// Only flag as issue if current shards result in >60GB per shard
	targetShardSizeGB := 60.0
	optimalShards := int(math.Ceil(indexSizeGB / targetShardSizeGB))

	// Minimum 1 shard
	if optimalShards < 1 {
		optimalShards = 1
	}

	// Don't recommend more shards than nodes
	if nodeCount > 0 && optimalShards > nodeCount {
		optimalShards = nodeCount
	}

	return optimalShards
}

func (s *checkService) calculateShardsByTraffic(trafficRate float64) int {
	if trafficRate == 0 {
		return 1 // Default to 1 shard for no traffic
	}

	// Use existing traffic-based logic but more conservative
	if trafficRate < constants.LowRateThreshold {
		return 1
	} else if trafficRate < constants.MediumRateThreshold {
		return 2
	} else if trafficRate < constants.HighRateThreshold {
		return 4
	} else if trafficRate < constants.VeryHighRateThreshold {
		return 8
	}

	return 12
}

func (s *checkService) calculateShardsByDocCount(docCount int64) int {
	if docCount == 0 {
		return 1
	}

	// Calculate shards based on optimal docs per shard (10M docs)
	optimalShards := int(math.Ceil(float64(docCount) / float64(constants.OptimalDocsPerShard)))

	if optimalShards < 1 {
		optimalShards = 1
	}

	return optimalShards
}

func (s *checkService) calculateConfidence(indexSize int64, trafficRate float64, docCount int64) float64 {
	confidence := 0.0

	// Size data is most reliable
	if indexSize > 0 {
		confidence += 0.5
	}

	// Traffic data adds confidence
	if trafficRate > 0 {
		confidence += 0.3
	}

	// Document count adds confidence
	if docCount > 0 {
		confidence += 0.2
	}

	return confidence
}

func (s *checkService) buildRecommendationReasoning(
	sizeShards, trafficShards, docShards int,
	indexSize int64, trafficRate float64, docCount int64,
) string {
	var reasons []string

	if indexSize > 0 {
		sizeGB := float64(indexSize) / float64(constants.BytesInGB)
		reasons = append(reasons, fmt.Sprintf("Size: %.1fGB→%d shards", sizeGB, sizeShards))
	}

	if trafficRate > 0 {
		reasons = append(reasons, fmt.Sprintf("Traffic: %.1freq/s→%d shards", trafficRate, trafficShards))
	}

	if docCount > 0 {
		reasons = append(reasons, fmt.Sprintf("Docs: %dM→%d shards", docCount/1000000, docShards))
	}

	if len(reasons) == 0 {
		return "Minimal data available"
	}

	return "Based on " + fmt.Sprintf("%s", reasons[0])
}

func (s *checkService) getNodeCount(ctx context.Context) int {
	nodes, err := s.client.GetNodes(ctx)
	if err != nil || nodes == nil {
		return 0
	}

	if nodesMap, ok := nodes["nodes"].(map[string]interface{}); ok {
		return len(nodesMap)
	}

	return 0
}

func (s *checkService) processIndexForScale(
	index map[string]interface{},
	trafficRatesMap map[string]indexTrafficRates,
	allStatsData map[string]interface{},
	nodeCount int,
) *models.OverScaledIndex {
	indexName := util.GetStringField(index, constants.IndexField)
	if util.IsSystemIndex(indexName) {
		return nil
	}
	primaryShards := util.GetIntField(index, constants.PrimaryField)
	replicaShards := util.GetIntField(index, constants.ReplicaField)

	// Parse as string if int parsing failed
	if primaryShards == 0 {
		if priStr := util.GetStringField(index, constants.PrimaryField); priStr != "" {
			if pri, err := strconv.Atoi(priStr); err == nil {
				primaryShards = pri
			}
		}
	}
	if replicaShards == 0 {
		if repStr := util.GetStringField(index, constants.ReplicaField); repStr != "" {
			if rep, err := strconv.Atoi(repStr); err == nil {
				replicaShards = rep
			}
		}
	}

	totalShards := primaryShards * (replicaShards + 1)

	// Get index size and doc count from stats
	indexSize, docCount := s.getIndexSizeAndDocCount(indexName, allStatsData)

	// Skip very small indices (less than 1GB)
	if indexSize < constants.MinIndexSizeForCheck {
		return nil
	}

	// Check per-shard size: if ≤60GB, no scale warning needed (size is acceptable)
	if primaryShards > 0 && indexSize > 0 {
		indexSizeGB := float64(indexSize) / float64(constants.BytesInGB)
		perShardGB := indexSizeGB / float64(primaryShards)

		// If per-shard size is 60GB or less, skip all scale warnings
		if perShardGB <= 60.0 {
			return nil
		}
	}

	// Get traffic rates
	var searchRate, indexRate float64
	if rates, ok := trafficRatesMap[indexName]; ok {
		searchRate = rates.searchRate
		indexRate = rates.indexRate
	}
	combinedRate := searchRate + indexRate

	// Calculate smart recommendation using hybrid scoring
	recommendation := s.calculateSmartRecommendation(indexSize, combinedRate, docCount, nodeCount, primaryShards)

	// Check if within acceptable range - if so, no warning needed
	if primaryShards >= recommendation.MinAcceptable && primaryShards <= recommendation.MaxAcceptable {
		return nil
	}

	// All scale issues are treated as warnings
	severity := constants.SeverityWarning

	// Determine warning type based on current vs recommended
	var warningType string
	var warningMessage string

	if primaryShards > recommendation.MaxAcceptable {
		warningType = constants.WarningTypeOverScaled
		warningMessage = fmt.Sprintf(
			"%s has %d primary shards (recommended: %d, acceptable range: %d-%d) [%s confidence] - %s",
			indexName, primaryShards, recommendation.Recommended,
			recommendation.MinAcceptable, recommendation.MaxAcceptable,
			s.confidenceLabel(recommendation.Confidence), recommendation.Reasoning,
		)
	} else if primaryShards < recommendation.MinAcceptable {
		warningType = constants.WarningTypeUnderScaled
		warningMessage = fmt.Sprintf(
			"%s has %d primary shards (recommended: %d, acceptable range: %d-%d) [%s confidence] - %s",
			indexName, primaryShards, recommendation.Recommended,
			recommendation.MinAcceptable, recommendation.MaxAcceptable,
			s.confidenceLabel(recommendation.Confidence), recommendation.Reasoning,
		)
	}

	// Check replicas separately
	if replicaShards > constants.MaxAcceptableReplicaCount {
		if warningType == "" {
			warningType = constants.WarningTypeOverReplicated
			warningMessage = fmt.Sprintf("%s has %d replicas (optimal: %d)",
				indexName, replicaShards, constants.OptimalReplicaCount)
		}
	} else if replicaShards < 1 {
		if warningType == "" {
			warningType = constants.WarningTypeUnderReplicated
			warningMessage = fmt.Sprintf("%s has no replicas (optimal: %d)",
				indexName, constants.OptimalReplicaCount)
		}
	}

	// If no warning, return nil
	if warningType == "" {
		return nil
	}

	return &models.OverScaledIndex{
		Name:           indexName,
		PrimaryShards:  primaryShards,
		ReplicaShards:  replicaShards,
		TotalShards:    totalShards,
		IndexSize:      indexSize,
		DocCount:       docCount,
		SearchRate:     searchRate,
		IndexRate:      indexRate,
		WarningType:    warningType,
		WarningMessage: warningMessage,
		Recommendation: recommendation,
		Severity:       severity,
	}
}

func (s *checkService) getIndexSizeAndDocCount(indexName string, allStatsData map[string]interface{}) (int64, int64) {
	if allStatsData == nil {
		return 0, 0
	}

	indices, ok := allStatsData["indices"].(map[string]interface{})
	if !ok {
		return 0, 0
	}

	indexData, ok := indices[indexName].(map[string]interface{})
	if !ok {
		return 0, 0
	}

	var indexSize int64
	var docCount int64

	// Get size from primaries
	if primaries, ok := indexData["primaries"].(map[string]interface{}); ok {
		if store, ok := primaries["store"].(map[string]interface{}); ok {
			if sizeInBytes, ok := store["size_in_bytes"].(float64); ok {
				indexSize = int64(sizeInBytes)
			}
		}
		if docs, ok := primaries["docs"].(map[string]interface{}); ok {
			if count, ok := docs["count"].(float64); ok {
				docCount = int64(count)
			}
		}
	}

	return indexSize, docCount
}

func (s *checkService) confidenceLabel(confidence float64) string {
	if confidence >= constants.HighConfidence {
		return "HIGH"
	} else if confidence >= constants.MediumConfidence {
		return "MEDIUM"
	}
	return "LOW"
}

func (s *checkService) parseAllIndexTrafficRates(statsData map[string]interface{}) map[string]indexTrafficRates {
	result := make(map[string]indexTrafficRates)

	if statsData == nil {
		return result
	}

	if indices, ok := statsData["indices"].(map[string]interface{}); ok {
		for indexName, indexDataRaw := range indices {
			if indexData, ok := indexDataRaw.(map[string]interface{}); ok {
				if total, ok := indexData["total"].(map[string]interface{}); ok {
					var queryTotal, queryTimeMs, indexTotal, indexTimeMs float64

					if search, ok := total["search"].(map[string]interface{}); ok {
						if qt, ok := search["query_total"].(float64); ok {
							queryTotal = qt
						}
						if qtm, ok := search["query_time_in_millis"].(float64); ok {
							queryTimeMs = qtm
						}
					}

					if indexing, ok := total["indexing"].(map[string]interface{}); ok {
						if it, ok := indexing["index_total"].(float64); ok {
							indexTotal = it
						}
						if itm, ok := indexing["index_time_in_millis"].(float64); ok {
							indexTimeMs = itm
						}
					}

					var searchRate, indexRate float64

					if queryTotal > 0 && queryTimeMs > 0 {
						searchRate = queryTotal / (queryTimeMs / 1000)
						if searchRate > 1000000 {
							searchRate = queryTotal / 86400
						}
					} else if queryTotal > 0 {
						searchRate = queryTotal / 3600
					}

					// Calculate index rate
					if indexTotal > 0 && indexTimeMs > 0 {
						indexRate = indexTotal / (indexTimeMs / 1000)
						if indexRate > 1000000 {
							indexRate = indexTotal / 86400
						}
					} else if indexTotal > 0 {
						indexRate = indexTotal / 3600
					}

					result[indexName] = indexTrafficRates{
						searchRate: searchRate,
						indexRate:  indexRate,
					}
				}
			}
		}
	}

	return result
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
