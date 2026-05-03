package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/util"
)

type NodeService interface {
	GetNodesInfo(ctx context.Context) ([]models.NodeInfo, error)
	GetNodeStats(ctx context.Context) ([]models.NodeStat, error)
	GetNodeBreakdown(ctx context.Context) (*models.NodeBreakdown, error)
	AnalyzeNodeBalance(ctx context.Context) (*models.BalanceAnalysis, error)
	GetNodeHealth(ctx context.Context) ([]models.NodeHealth, error)
}

type nodeService struct {
	client interfaces.ElasticClient
}

func NewNodeService(client interfaces.ElasticClient) NodeService {
	return &nodeService{
		client: client,
	}
}

func (s *nodeService) GetNodesInfo(ctx context.Context) ([]models.NodeInfo, error) {
	infoData, err := s.client.GetNodesInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrNodeInfoRequestFailed, err)
	}

	statsData, err := s.client.GetNodesStats(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrNodeStatsRequestFailed2, err)
	}

	if nodes, ok := infoData[constants.NodesField].(map[string]interface{}); ok {
		infos := make([]models.NodeInfo, 0)

		for nodeID, nodeData := range nodes {
			if node, ok := nodeData.(map[string]interface{}); ok {
				name, _ := node[constants.NameField].(string)
				ip, _ := node[constants.IPField].(string)

				var roles []string
				if rolesData, ok := node[constants.RolesField].([]interface{}); ok {
					for _, role := range rolesData {
						if roleStr, ok := role.(string); ok {
							roles = append(roles, roleStr)
						}
					}
				}

				var stats map[string]interface{}
				if statsNodes, ok := statsData[constants.NodesField].(map[string]interface{}); ok {
					if nodeStats, ok := statsNodes[nodeID].(map[string]interface{}); ok {
						stats = nodeStats
					}
				}

				nodeInfo := models.NodeInfo{
					Name:        name,
					IP:          ip,
					Roles:       roles,
					CPUPercent:  constants.ZeroPercentString,
					MemPercent:  constants.ZeroPercentString,
					HeapPercent: constants.ZeroPercentString,
					DiskAvail:   constants.DashString,
					DiskTotal:   constants.DashString,
					Documents:   0,
					HeapUsed:    constants.DashString,
					HeapMax:     constants.DashString,
				}

				if stats != nil {
					if jvm, ok := stats[constants.JVMField].(map[string]interface{}); ok {
						if mem, ok := jvm[constants.MemField].(map[string]interface{}); ok {
							if heapUsed, ok := mem[constants.HeapUsedInBytesField].(float64); ok {
								nodeInfo.HeapUsed = models.FormatBytes(int64(heapUsed))
							}
							if heapMax, ok := mem[constants.HeapMaxInBytesField].(float64); ok {
								nodeInfo.HeapMax = models.FormatBytes(int64(heapMax))
							}

							if heapPercent, ok := mem[constants.HeapUsedPercentField].(float64); ok {
								nodeInfo.HeapPercent = fmt.Sprintf(constants.PercentFormat, heapPercent)
							}
						}
					}

					if fs, ok := stats[constants.FSField].(map[string]interface{}); ok {
						if total, ok := fs[constants.TotalField].(map[string]interface{}); ok {
							if totalBytes, ok := total[constants.TotalInBytesField].(float64); ok {
								nodeInfo.DiskTotal = models.FormatBytes(int64(totalBytes))
							}
							if availBytes, ok := total[constants.AvailableInBytesField].(float64); ok {
								nodeInfo.DiskAvail = models.FormatBytes(int64(availBytes))

								// Calculate disk usage percentage
								if totalBytes, ok := total[constants.TotalInBytesField].(float64); ok {
									usedBytes := int64(totalBytes - availBytes)
									diskPercent := util.CalculatePercentage(usedBytes, int64(totalBytes))
									nodeInfo.DiskPercent = fmt.Sprintf(constants.PercentFormat, diskPercent)
								}
							}
						}
					}

					if process, ok := stats[constants.ProcessField].(map[string]interface{}); ok {
						if cpu, ok := process[constants.CPUField].(map[string]interface{}); ok {
							if percent, ok := cpu[constants.PercentField].(float64); ok {
								nodeInfo.CPUPercent = fmt.Sprintf(constants.PercentFormat, percent)
							}
						}
					}

					if os, ok := stats[constants.OSField].(map[string]interface{}); ok {
						if mem, ok := os[constants.MemField].(map[string]interface{}); ok {
							if usedPercent, ok := mem[constants.UsedPercentField].(float64); ok {
								nodeInfo.MemPercent = fmt.Sprintf(constants.PercentFormat, usedPercent)
							}
						}
					}

					if indices, ok := stats[constants.IndicesField].(map[string]interface{}); ok {
						if docs, ok := indices[constants.DocsField].(map[string]interface{}); ok {
							if count, ok := docs[constants.CountField].(float64); ok {
								nodeInfo.Documents = int64(count)
							}
						}
					}
				}

				infos = append(infos, nodeInfo)
			}
		}

		return infos, nil
	}

	return []models.NodeInfo{}, nil
}

func (s *nodeService) GetNodeStats(ctx context.Context) ([]models.NodeStat, error) {
	shardsData, err := s.client.GetShards(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrShardInfoRequestFailed, err)
	}

	nodeStats := make(map[string]struct {
		PrimaryShards int
		ReplicaShards int
		TotalShards   int
		TotalSize     int64
		Indices       map[string]bool
	})

	processedCount := 0

	if shards, ok := shardsData[constants.EmptyString].([]map[string]interface{}); ok {
		for _, shard := range shards {
			index, _ := shard[constants.IndexField].(string)
			prirep, _ := shard[constants.PrirepField].(string)
			state, _ := shard[constants.StateField].(string)
			nodeIP, _ := shard[constants.ShardRowNodeField].(string)
			store, _ := shard[constants.StoreField].(string)

			if state != constants.ShardStateStarted || nodeIP == constants.DashString {
				continue
			}

			processedCount++

			if _, exists := nodeStats[nodeIP]; !exists {
				nodeStats[nodeIP] = struct {
					PrimaryShards int
					ReplicaShards int
					TotalShards   int
					TotalSize     int64
					Indices       map[string]bool
				}{
					Indices: make(map[string]bool),
				}
			}

			stats := nodeStats[nodeIP]
			if prirep == constants.PrimaryShortString {
				stats.PrimaryShards++
			} else {
				stats.ReplicaShards++
			}
			stats.TotalShards++
			stats.Indices[index] = true

			sizeBytes := models.ParseSize(store)
			stats.TotalSize += sizeBytes

			nodeStats[nodeIP] = stats
		}
	}

	var nodeStatsList []models.NodeStat
	for nodeIP, stats := range nodeStats {
		nodeStatsList = append(nodeStatsList, models.NodeStat{
			NodeIP:        nodeIP,
			PrimaryShards: stats.PrimaryShards,
			ReplicaShards: stats.ReplicaShards,
			TotalShards:   stats.TotalShards,
			TotalSize:     stats.TotalSize,
			IndexCount:    len(stats.Indices),
		})
	}

	sort.Slice(nodeStatsList, func(i, j int) bool {
		return nodeStatsList[i].TotalSize > nodeStatsList[j].TotalSize
	})

	return nodeStatsList, nil
}

func (s *nodeService) AnalyzeNodeBalance(ctx context.Context) (*models.BalanceAnalysis, error) {
	nodeStats, err := s.GetNodeStats(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFailedToGetNodeStats, err)
	}

	if len(nodeStats) == 0 {
		return &models.BalanceAnalysis{
			IsBalanced:     true,
			Recommendation: constants.MsgNoNodesFound,
		}, nil
	}

	sort.Slice(nodeStats, func(i, j int) bool {
		return nodeStats[i].TotalShards > nodeStats[j].TotalShards
	})

	maxShards := nodeStats[0].TotalShards
	minShards := nodeStats[len(nodeStats)-1].TotalShards
	balanceRatio := float64(minShards) / float64(maxShards)

	analysis := &models.BalanceAnalysis{
		MostLoadedNode:  nodeStats[0].NodeIP,
		LeastLoadedNode: nodeStats[len(nodeStats)-1].NodeIP,
		MaxShards:       maxShards,
		MinShards:       minShards,
		BalanceRatio:    balanceRatio,
		IsBalanced:      balanceRatio >= constants.BalanceRatioThreshold,
	}

	if !analysis.IsBalanced {
		analysis.Recommendation = constants.MsgConsiderRebalancing
	} else {
		analysis.Recommendation = constants.MsgNodeBalanceGood
	}

	return analysis, nil
}

func (s *nodeService) GetNodeBreakdown(ctx context.Context) (*models.NodeBreakdown, error) {
	nodesInfo, err := s.GetNodesInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFailedToGetNodeInfo, err)
	}

	breakdown := &models.NodeBreakdown{}

	for _, node := range nodesInfo {
		hasDataRole := false
		hasMasterRole := false
		hasIngestRole := false

		for _, role := range node.Roles {
			switch role {
			case constants.NodeRoleData:
				hasDataRole = true
				breakdown.DataNodes++
			case constants.NodeRoleMaster:
				hasMasterRole = true
				breakdown.MasterNodes++
			case constants.NodeRoleIngest:
				hasIngestRole = true
				breakdown.IngestNodes++
			}
		}

		if !hasDataRole && !hasMasterRole && !hasIngestRole {
			breakdown.CoordinatingNodes++
		}
	}

	return breakdown, nil
}

func (s *nodeService) GetNodeHealth(ctx context.Context) ([]models.NodeHealth, error) {
	nodes, err := s.GetNodesInfo(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFailedToGetNodeInfo, err)
	}

	var healthList []models.NodeHealth
	for _, node := range nodes {
		health := models.NodeHealth{
			NodeName:  node.Name,
			Status:    constants.HealthyString,
			IsHealthy: true,
		}

		if node.CPUPercent != constants.EmptyString && node.CPUPercent != constants.ZeroPercentString {
			if cpu, err := util.ParsePercentString(node.CPUPercent); err == nil {
				health.CPUUsage = cpu
				if cpu > constants.HighCPUThreshold {
					health.IsHealthy = false
					health.Issues = append(health.Issues, constants.MsgHighCPUUsage)
				}
			}
		}

		if node.MemPercent != constants.EmptyString && node.MemPercent != constants.ZeroPercentString {
			if mem, err := util.ParsePercentString(node.MemPercent); err == nil {
				health.MemoryUsage = mem
				if mem > constants.HighMemoryThreshold {
					health.IsHealthy = false
					health.Issues = append(health.Issues, constants.MsgHighMemoryUsage)
				}
			}
		}

		if node.HeapPercent != constants.EmptyString && node.HeapPercent != constants.ZeroPercentString {
			if heap, err := util.ParsePercentString(node.HeapPercent); err == nil {
				health.HeapUsage = heap
				if heap > constants.HighHeapThreshold {
					health.IsHealthy = false
					health.Issues = append(health.Issues, constants.MsgHighHeapUsage)
				}
			}
		}

		if node.DiskAvail != constants.EmptyString && node.DiskTotal != constants.EmptyString && node.DiskAvail != constants.DashString && node.DiskTotal != constants.DashString {
			avail := models.ParseSize(node.DiskAvail)
			total := models.ParseSize(node.DiskTotal)
			if total > 0 {
				used := total - avail
				diskUsage := float64(used) / float64(total) * constants.HundredMultiplier
				health.DiskUsage = diskUsage
				if diskUsage > constants.HighDiskThreshold {
					health.IsHealthy = false
					health.Issues = append(health.Issues, constants.MsgHighDiskUsage)
				}
			}
		}

		if !health.IsHealthy {
			health.Status = constants.WarningString
		}

		healthList = append(healthList, health)
	}

	return healthList, nil
}
