package services

import (
	"context"
	"fmt"

	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/util"
)

type ShardService interface {
	GetAllShardInfos(ctx context.Context) ([]models.ShardInfo, error)
	GetShardDistribution(ctx context.Context) (*models.ShardDistribution, error)
	GetShardWarnings(ctx context.Context) (*models.ShardWarnings, error)
}

type shardService struct {
	client interfaces.ElasticClient
}

func NewShardService(client interfaces.ElasticClient) ShardService {
	return &shardService{
		client: client,
	}
}

func (s *shardService) GetAllShardInfos(ctx context.Context) ([]models.ShardInfo, error) {
	shardsData, err := s.client.GetShards(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrShardsRequestFailed2, err)
	}

	var shards []models.ShardInfo
	if shardsList, ok := shardsData[constants.EmptyString].([]map[string]interface{}); ok {
		for _, shard := range shardsList {
			shardInfo := models.ShardInfo{
				Index:  util.GetStringField(shard, constants.IndexField),
				Shard:  util.GetStringField(shard, constants.ShardField),
				Prirep: util.GetStringField(shard, constants.PrirepField),
				State:  util.GetStringField(shard, constants.StateField),
				Docs:   util.GetStringField(shard, constants.DocsCountField),
				Store:  util.GetStringField(shard, constants.StoreField),
				IP:     util.GetStringField(shard, constants.IPField),
				Node:   util.GetStringField(shard, constants.ShardRowNodeField),
			}
			shards = append(shards, shardInfo)
		}
	}

	return shards, nil
}

func (s *shardService) GetShardDistribution(ctx context.Context) (*models.ShardDistribution, error) {
	shards, err := s.GetAllShardInfos(ctx)
	if err != nil {
		return nil, err
	}

	distribution := &models.ShardDistribution{
		NodeDistribution:  make(map[string]int),
		IndexDistribution: make(map[string]*models.ShardStat),
	}

	for _, shard := range shards {
		if shard.State != constants.ShardStateStarted {
			continue
		}

		if shard.IP != constants.DashString {
			distribution.NodeDistribution[shard.IP]++
		}

		if _, exists := distribution.IndexDistribution[shard.Index]; !exists {
			distribution.IndexDistribution[shard.Index] = &models.ShardStat{
				IndexName:     shard.Index,
				PrimaryShards: 0,
				ReplicaShards: 0,
				TotalShards:   0,
				Nodes:         make(map[string]bool),
			}
		}

		stats := distribution.IndexDistribution[shard.Index]
		if shard.Prirep == constants.PrimaryShortString {
			stats.PrimaryShards++
		} else {
			stats.ReplicaShards++
		}
		stats.TotalShards++
		if shard.IP != constants.DashString {
			stats.Nodes[shard.IP] = true
		}
	}

	return distribution, nil
}

func (s *shardService) GetShardWarnings(ctx context.Context) (*models.ShardWarnings, error) {
	shards, err := s.GetAllShardInfos(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrFailedToGetShardInfo, err)
	}

	warnings := &models.ShardWarnings{
		Recommendations: make([]string, 0),
		CriticalIssues:  make([]string, 0),
		WarningIssues:   make([]string, 0),
	}

	nodeShardCounts := make(map[string]int)
	for _, shard := range shards {
		switch shard.State {
		case constants.ShardStateUnassigned:
			warnings.UnassignedShards++
		case constants.ShardStateRelocating:
			warnings.RelocatingShards++
		case constants.ShardStateInitializing:
			warnings.InitializingShards++
		case constants.ShardStateStarted:
			if shard.IP != constants.DashString {
				nodeShardCounts[shard.IP]++
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
