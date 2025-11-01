package services

import (
	"context"
	"fmt"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
	"github.com/mertbahardogan/escope/internal/util"
	"time"
)

type IndexService interface {
	GetAllIndexInfos(ctx context.Context) ([]models.IndexInfo, error)
	GetLuceneStats(ctx context.Context) ([]models.LuceneStats, error)
	GetIndexDetailInfo(ctx context.Context, indexName string) (*models.IndexDetailInfo, error)
}

type indexService struct {
	client interfaces.ElasticClient
	cache  *models.IndexStatsCache
}

func NewIndexService(client interfaces.ElasticClient) IndexService {
	return &indexService{
		client: client,
		cache:  models.NewIndexStatsCache(),
	}
}

func (s *indexService) GetAllIndexInfos(ctx context.Context) ([]models.IndexInfo, error) {
	indicesData, err := s.client.GetIndices(ctx)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndicesRequestFailed2, err)
	}

	var indices []models.IndexInfo
	if indicesList, ok := indicesData[constants.EmptyString].([]map[string]interface{}); ok {
		for _, idx := range indicesList {
			index := models.IndexInfo{
				Alias:     util.GetStringField(idx, constants.AliasField),
				Name:      util.GetStringField(idx, constants.IndexField),
				Health:    util.GetStringField(idx, constants.HealthField),
				Status:    util.GetStringField(idx, constants.StatusField),
				DocsCount: util.GetStringField(idx, constants.DocsCountField),
				StoreSize: util.GetStringField(idx, constants.StoreSizeField),
				Primary:   util.GetStringField(idx, constants.PrimaryField),
				Replica:   util.GetStringField(idx, constants.ReplicaField),
			}
			indices = append(indices, index)
		}
	}

	return indices, nil
}

func (s *indexService) GetLuceneStats(ctx context.Context) ([]models.LuceneStats, error) {
	statsData, err := s.client.GetIndexStats(ctx, constants.EmptyString)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndexStatsRequestFailed, err)
	}

	var luceneStats []models.LuceneStats

	indexStats := parseIndexStatsData(statsData)

	for indexName, total := range indexStats {
		if segments, ok := getSegmentsData(total); ok {
			if indexing, ok := getIndexingData(total); ok {
				stats := parseLuceneStats(indexName, segments, indexing)
				luceneStats = append(luceneStats, stats)
			}
		}
	}

	return luceneStats, nil
}

func (s *indexService) GetIndexDetailInfo(ctx context.Context, indexName string) (*models.IndexDetailInfo, error) {
	statsData, err := s.client.GetIndexStats(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndexStatsRequestFailed, err)
	}

	var basicInfo models.IndexDetailInfo
	basicInfo.Name = indexName
	basicInfo.SearchRate = constants.DashString
	basicInfo.IndexRate = constants.DashString
	basicInfo.AvgQueryTime = constants.DashString
	basicInfo.AvgIndexTime = constants.DashString
	currentTime := time.Now()

	if indices, ok := statsData["indices"].(map[string]interface{}); ok {
		if indexData, ok := indices[indexName].(map[string]interface{}); ok {
			if total, ok := indexData["total"].(map[string]interface{}); ok {
				var currentQueryTotal, currentQueryTime, currentIndexTotal, currentIndexTime int64

				if search, ok := total["search"].(map[string]interface{}); ok {
					if queryTotal, ok := search["query_total"].(float64); ok {
						currentQueryTotal = int64(queryTotal)
					}
					if queryTime, ok := search["query_time_in_millis"].(float64); ok {
						currentQueryTime = int64(queryTime)
					}
				}
				if indexing, ok := total["indexing"].(map[string]interface{}); ok {
					if indexTotal, ok := indexing["index_total"].(float64); ok {
						currentIndexTotal = int64(indexTotal)
					}
					if indexTime, ok := indexing["index_time_in_millis"].(float64); ok {
						currentIndexTime = int64(indexTime)
					}
				}
				if prevSnapshot, exists := s.cache.GetSnapshot(indexName); exists {
					timeDelta := currentTime.Sub(prevSnapshot.Timestamp).Seconds()
					if timeDelta > 0 {
						queryDelta := currentQueryTotal - prevSnapshot.QueryTotal
						if queryDelta > 0 {
							searchRate := float64(queryDelta) / timeDelta
							basicInfo.SearchRate = s.formatRate(searchRate)

							queryTimeDelta := currentQueryTime - prevSnapshot.QueryTime
							avgQueryTime := float64(queryTimeDelta) / float64(queryDelta)
							basicInfo.AvgQueryTime = fmt.Sprintf(constants.TimeFormatMS, avgQueryTime)
						} else {
							basicInfo.SearchRate = constants.DashString
						}
						indexDelta := currentIndexTotal - prevSnapshot.IndexTotal
						if indexDelta > 0 {
							indexRate := float64(indexDelta) / timeDelta
							basicInfo.IndexRate = s.formatRate(indexRate)

							indexTimeDelta := currentIndexTime - prevSnapshot.IndexTime
							avgIndexTime := float64(indexTimeDelta) / float64(indexDelta)
							basicInfo.AvgIndexTime = fmt.Sprintf(constants.TimeFormatMS, avgIndexTime)
						} else {
							basicInfo.IndexRate = constants.DashString
						}
					}
				} else {
					if currentQueryTotal > 0 {
						basicInfo.AvgQueryTime = fmt.Sprintf(constants.TimeFormatMS, float64(currentQueryTime)/float64(currentQueryTotal))
					}
					if currentIndexTotal > 0 {
						basicInfo.AvgIndexTime = fmt.Sprintf(constants.TimeFormatMS, float64(currentIndexTime)/float64(currentIndexTotal))
					}
				}
				newSnapshot := &models.IndexStatsSnapshot{
					IndexName:  indexName,
					QueryTotal: currentQueryTotal,
					QueryTime:  currentQueryTime,
					IndexTotal: currentIndexTotal,
					IndexTime:  currentIndexTime,
					Timestamp:  currentTime,
				}
				s.cache.SetSnapshot(newSnapshot)
			}
		}
	}

	return &basicInfo, nil
}

func (s *indexService) formatRate(rate float64) string {
	if rate >= constants.ThousandDivisor {
		return fmt.Sprintf(constants.RateFormatK, rate/constants.ThousandDivisor)
	} else if rate >= 1 {
		return fmt.Sprintf(constants.RateFormat, rate)
	} else {
		return fmt.Sprintf(constants.RateFormat2, rate)
	}
}

func parseLuceneStats(indexName string, segments, indexing map[string]interface{}) models.LuceneStats {
	stats := models.LuceneStats{
		IndexName: indexName,
	}

	if count, ok := segments["count"].(float64); ok {
		stats.SegmentCount = int(count)
	}

	if memory, ok := segments["memory"].(map[string]interface{}); ok {
		if totalBytes, ok := memory["total_in_bytes"].(float64); ok {
			stats.SegmentMemoryBytes = int64(totalBytes)
			stats.SegmentMemory = models.FormatBytes(int64(totalBytes))
		}
	}

	if terms, ok := segments["terms"].(map[string]interface{}); ok {
		if memoryBytes, ok := terms["memory_in_bytes"].(float64); ok {
			stats.TermsMemoryBytes = int64(memoryBytes)
			stats.TermsMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if stored, ok := segments["stored"].(map[string]interface{}); ok {
		if memoryBytes, ok := stored["memory_in_bytes"].(float64); ok {
			stats.StoredMemoryBytes = int64(memoryBytes)
			stats.StoredMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if docValues, ok := segments["doc_values"].(map[string]interface{}); ok {
		if memoryBytes, ok := docValues["memory_in_bytes"].(float64); ok {
			stats.DocValuesMemoryBytes = int64(memoryBytes)
			stats.DocValuesMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if points, ok := segments["points"].(map[string]interface{}); ok {
		if memoryBytes, ok := points["memory_in_bytes"].(float64); ok {
			stats.PointsMemoryBytes = int64(memoryBytes)
			stats.PointsMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if norms, ok := segments["norms"].(map[string]interface{}); ok {
		if memoryBytes, ok := norms["memory_in_bytes"].(float64); ok {
			stats.NormsMemoryBytes = int64(memoryBytes)
			stats.NormsMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if fixedBitSet, ok := segments["fixed_bit_set"].(map[string]interface{}); ok {
		if memoryBytes, ok := fixedBitSet["memory_in_bytes"].(float64); ok {
			stats.FixedBitSetMemoryBytes = int64(memoryBytes)
			stats.FixedBitSetMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if versionMap, ok := segments["version_map"].(map[string]interface{}); ok {
		if memoryBytes, ok := versionMap["memory_in_bytes"].(float64); ok {
			stats.VersionMapMemoryBytes = int64(memoryBytes)
			stats.VersionMapMemory = models.FormatBytes(int64(memoryBytes))
		}
	}

	if maxUnsafeAutoID, ok := segments["max_unsafe_auto_id_timestamp"].(float64); ok {
		stats.MaxUnsafeAutoIDTimestamp = int64(maxUnsafeAutoID)
	}

	if indexMemory, ok := indexing["index_memory"].(map[string]interface{}); ok {
		if totalBytes, ok := indexMemory["total_in_bytes"].(float64); ok {
			stats.IndexMemoryBytes = int64(totalBytes)
			stats.IndexMemory = models.FormatBytes(int64(totalBytes))
		}
	}

	return stats
}
