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
	GetIndexMapping(ctx context.Context, indexName string) ([]models.FieldMapping, error)
	GetIndexSettings(ctx context.Context, indexName string) ([]models.IndexSettingInfo, error)
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
		segments, hasSegments := getSegmentsData(total)
		indexing, hasIndexing := getIndexingData(total)

		if hasSegments && hasIndexing {
			stats := parseLuceneStats(indexName, segments, indexing)
			luceneStats = append(luceneStats, stats)
		}
	}

	return luceneStats, nil
}

func (s *indexService) GetIndexDetailInfo(ctx context.Context, indexName string) (*models.IndexDetailInfo, error) {
	statsData, err := s.client.GetIndexStats(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndexStatsRequestFailed, err)
	}

	// Check if index exists
	indices, ok := statsData["indices"].(map[string]interface{})
	if !ok || len(indices) == 0 {
		return nil, fmt.Errorf("index '%s' not found", indexName)
	}

	var basicInfo models.IndexDetailInfo
	basicInfo.Name = indexName
	basicInfo.SearchRate = constants.DashString
	basicInfo.IndexRate = constants.DashString
	basicInfo.AvgQueryTime = constants.DashString
	basicInfo.AvgIndexTime = constants.DashString
	currentTime := time.Now()

	// Get first index from response (works with both alias and real index name)
	var indexData map[string]interface{}
	for _, data := range indices {
		if d, ok := data.(map[string]interface{}); ok {
			indexData = d
			break
		}
	}
	if indexData != nil {
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

func (s *indexService) GetIndexMapping(ctx context.Context, indexName string) ([]models.FieldMapping, error) {
	mappingData, err := s.client.GetIndexMapping(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf("mapping request failed: %w", err)
	}

	var fields []models.FieldMapping

	// Iterate through index mappings (response format: {indexName: {mappings: {...}}})
	for _, indexData := range mappingData {
		if indexMap, ok := indexData.(map[string]interface{}); ok {
			if mappings, ok := indexMap["mappings"].(map[string]interface{}); ok {
				if properties, ok := mappings["properties"].(map[string]interface{}); ok {
					fields = extractFields(properties, "", 0)
				}
			}
		}
	}

	return fields, nil
}

func extractFields(properties map[string]interface{}, prefix string, depth int) []models.FieldMapping {
	var fields []models.FieldMapping

	for fieldName, fieldData := range properties {
		path := fieldName
		if prefix != "" {
			path = prefix + "." + fieldName
		}

		if fieldMap, ok := fieldData.(map[string]interface{}); ok {
			field := models.FieldMapping{
				Path:           path,
				Name:           fieldName,
				Type:           getStringOrDefault(fieldMap, "type", "-"),
				Analyzer:       getStringOrDefault(fieldMap, "analyzer", "-"),
				SearchAnalyzer: getStringOrDefault(fieldMap, "search_analyzer", "-"),
				Normalizer:     getStringOrDefault(fieldMap, "normalizer", "-"),
				Index:          getBoolAsString(fieldMap, "index", "true"),
				Store:          getBoolAsString(fieldMap, "store", "false"),
				Depth:          depth,
			}

			// Handle nested properties
			if nestedProps, ok := fieldMap["properties"].(map[string]interface{}); ok {
				// Add parent field with type "object" if no type specified
				if field.Type == "-" {
					field.Type = "object"
				}
				fields = append(fields, field)
				// Recursively extract nested fields
				nestedFields := extractFields(nestedProps, path, depth+1)
				fields = append(fields, nestedFields...)
			} else {
				fields = append(fields, field)
			}
		}
	}

	return fields
}

func getStringOrDefault(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultVal
}

func getBoolAsString(m map[string]interface{}, key, defaultVal string) string {
	if val, ok := m[key].(bool); ok {
		if val {
			return "true"
		}
		return "false"
	}
	return defaultVal
}

func (s *indexService) GetIndexSettings(ctx context.Context, indexName string) ([]models.IndexSettingInfo, error) {
	settingsData, err := s.client.GetIndexSettings(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf("settings request failed: %w", err)
	}

	var settings []models.IndexSettingInfo

	// Iterate through index settings (response format: {indexName: {settings: {...}}})
	for _, indexData := range settingsData {
		if indexMap, ok := indexData.(map[string]interface{}); ok {
			if settingsMap, ok := indexMap["settings"].(map[string]interface{}); ok {
				settings = flattenSettings(settingsMap, "")
			}
		}
	}

	return settings, nil
}

func flattenSettings(data map[string]interface{}, prefix string) []models.IndexSettingInfo {
	var settings []models.IndexSettingInfo

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]interface{}:
			// Recursively flatten nested settings
			nestedSettings := flattenSettings(v, fullKey)
			settings = append(settings, nestedSettings...)
		case string:
			settings = append(settings, models.IndexSettingInfo{Key: fullKey, Value: v})
		case float64:
			settings = append(settings, models.IndexSettingInfo{Key: fullKey, Value: fmt.Sprintf("%v", v)})
		case bool:
			settings = append(settings, models.IndexSettingInfo{Key: fullKey, Value: fmt.Sprintf("%v", v)})
		default:
			settings = append(settings, models.IndexSettingInfo{Key: fullKey, Value: fmt.Sprintf("%v", v)})
		}
	}

	return settings
}
