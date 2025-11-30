package services

import "github.com/mertbahardogan/escope/internal/constants"

func parseIndexStatsData(statsData map[string]interface{}) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	indices, ok := statsData[constants.IndicesField].(map[string]interface{})
	if !ok {
		return result
	}

	for indexName, indexData := range indices {
		if index, ok := indexData.(map[string]interface{}); ok {
			if total, ok := index[constants.TotalField].(map[string]interface{}); ok {
				result[indexName] = total
			}
		}
	}

	return result
}

func getSegmentsData(total map[string]interface{}) (map[string]interface{}, bool) {
	segments, ok := total[constants.SegmentsField].(map[string]interface{})
	return segments, ok
}

func getIndexingData(total map[string]interface{}) (map[string]interface{}, bool) {
	indexing, ok := total[constants.IndexingField].(map[string]interface{})
	return indexing, ok
}

func getSearchData(total map[string]interface{}) (map[string]interface{}, bool) {
	search, ok := total[constants.SearchField].(map[string]interface{})
	return search, ok
}

func extractTrafficMetrics(total map[string]interface{}) (queryTotal, queryTime, indexTotal, indexTime float64) {
	if search, ok := getSearchData(total); ok {
		if qt, ok := search[constants.QueryTotalField].(float64); ok {
			queryTotal = qt
		}
		if qtm, ok := search[constants.QueryTimeInMillisField].(float64); ok {
			queryTime = qtm
		}
	}

	if indexing, ok := getIndexingData(total); ok {
		if it, ok := indexing[constants.IndexTotalField].(float64); ok {
			indexTotal = it
		}
		if itm, ok := indexing[constants.IndexTimeInMillisField].(float64); ok {
			indexTime = itm
		}
	}

	return
}

func calculateTrafficRate(total, timeMs float64) float64 {
	if total > 0 && timeMs > 0 {
		rate := total / (timeMs / 1000)
		// If rate seems unrealistic (> 1M ops/sec), use daily average
		if rate > 1000000 {
			return total / 86400
		}
		return rate
	} else if total > 0 {
		// Fallback to hourly average if time data is missing
		return total / 3600
	}
	return 0
}

// getIndexSizeAndDocCount extracts index size and document count from stats data
func getIndexSizeAndDocCount(indexName string, allStatsData map[string]interface{}) (int64, int64) {
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

func getNodeCount(nodesData map[string]interface{}) int {
	if nodesData == nil {
		return 0
	}

	if nodesMap, ok := nodesData["nodes"].(map[string]interface{}); ok {
		return len(nodesMap)
	}

	return 0
}
