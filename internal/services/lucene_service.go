package services

import (
	"context"
	"fmt"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
)

type LuceneService interface {
	GetLuceneStats(ctx context.Context) ([]models.LuceneStats, error)
	GetLuceneStatsForIndex(ctx context.Context, indexName string) (*models.LuceneStats, error)
}

type luceneService struct {
	client interfaces.ElasticClient
}

func NewLuceneService(client interfaces.ElasticClient) LuceneService {
	return &luceneService{
		client: client,
	}
}

func (s *luceneService) GetLuceneStats(ctx context.Context) ([]models.LuceneStats, error) {
	statsData, err := s.client.GetIndexStats(ctx, constants.EmptyString)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndexStatsRequestFailed, err)
	}

	var luceneStats []models.LuceneStats

	indexStats := parseIndexStatsData(statsData)

	for indexName, total := range indexStats {
		if segments, ok := getSegmentsData(total); ok {
			if indexing, ok := getIndexingData(total); ok {
				stats := s.parseLuceneStats(indexName, segments, indexing)
				luceneStats = append(luceneStats, stats)
			}
		}
	}
	return luceneStats, nil
}

func (s *luceneService) GetLuceneStatsForIndex(ctx context.Context, indexName string) (*models.LuceneStats, error) {
	statsData, err := s.client.GetIndexStats(ctx, indexName)
	if err != nil {
		return nil, fmt.Errorf(constants.ErrIndexStatsRequestFailed, err)
	}

	indexStats := parseIndexStatsData(statsData)

	// Get first index from response (works with both alias and real index name)
	for resolvedIndexName, total := range indexStats {
		if segments, ok := getSegmentsData(total); ok {
			if indexing, ok := getIndexingData(total); ok {
				stats := s.parseLuceneStats(resolvedIndexName, segments, indexing)
				return &stats, nil
			}
		}
		break
	}

	return nil, fmt.Errorf("'%s' index not found", indexName)
}

func (s *luceneService) parseLuceneStats(indexName string, segments, indexing map[string]interface{}) models.LuceneStats {
	stats := models.LuceneStats{
		IndexName: indexName,
	}
	totalCalculatedMemory := int64(0)

	if count, ok := segments[constants.CountField].(float64); ok {
		stats.SegmentCount = int(count)
	}

	if totalBytes, ok := segments[constants.MemoryInBytesField].(float64); ok {
		stats.SegmentMemoryBytes = int64(totalBytes)
		stats.SegmentMemory = models.FormatBytes(int64(totalBytes))
	}

	if termsBytes, ok := segments[constants.TermsMemoryInBytesField].(float64); ok {
		stats.TermsMemoryBytes = int64(termsBytes)
		stats.TermsMemory = models.FormatBytes(int64(termsBytes))
		totalCalculatedMemory += int64(termsBytes)
	} else if termsBytes, ok := segments[constants.TermsField].(map[string]interface{}); ok {
		if memoryBytes, ok := termsBytes[constants.MemoryInBytesField].(float64); ok {
			stats.TermsMemoryBytes = int64(memoryBytes)
			stats.TermsMemory = models.FormatBytes(int64(memoryBytes))
			totalCalculatedMemory += int64(memoryBytes)
		}
	}
	if storedBytes, ok := segments[constants.StoredFieldsMemoryInBytesField].(float64); ok {
		stats.StoredMemoryBytes = int64(storedBytes)
		stats.StoredMemory = models.FormatBytes(int64(storedBytes))
		totalCalculatedMemory += int64(storedBytes)
	} else if storedBytes, ok := segments[constants.StoredFieldsField].(map[string]interface{}); ok {
		if memoryBytes, ok := storedBytes[constants.MemoryInBytesField].(float64); ok {
			stats.StoredMemoryBytes = int64(memoryBytes)
			stats.StoredMemory = models.FormatBytes(int64(memoryBytes))
			totalCalculatedMemory += int64(memoryBytes)
		}
	}

	if docValuesBytes, ok := segments[constants.DocValuesMemoryInBytesField].(float64); ok {
		stats.DocValuesMemoryBytes = int64(docValuesBytes)
		stats.DocValuesMemory = models.FormatBytes(int64(docValuesBytes))
		totalCalculatedMemory += int64(docValuesBytes)
	} else if docValuesBytes, ok := segments[constants.DocValuesField].(map[string]interface{}); ok {
		if memoryBytes, ok := docValuesBytes[constants.MemoryInBytesField].(float64); ok {
			stats.DocValuesMemoryBytes = int64(memoryBytes)
			stats.DocValuesMemory = models.FormatBytes(int64(memoryBytes))
			totalCalculatedMemory += int64(memoryBytes)
		}
	}

	if pointsBytes, ok := segments[constants.PointsMemoryInBytesField].(float64); ok {
		stats.PointsMemoryBytes = int64(pointsBytes)
		stats.PointsMemory = models.FormatBytes(int64(pointsBytes))
		totalCalculatedMemory += int64(pointsBytes)
	} else if pointsBytes, ok := segments[constants.PointsField].(map[string]interface{}); ok {
		if memoryBytes, ok := pointsBytes[constants.MemoryInBytesField].(float64); ok {
			stats.PointsMemoryBytes = int64(memoryBytes)
			stats.PointsMemory = models.FormatBytes(int64(memoryBytes))
			totalCalculatedMemory += int64(memoryBytes)
		}
	}

	if normsBytes, ok := segments[constants.NormsMemoryInBytesField].(float64); ok {
		stats.NormsMemoryBytes = int64(normsBytes)
		stats.NormsMemory = models.FormatBytes(int64(normsBytes))
		totalCalculatedMemory += int64(normsBytes)
	} else if normsBytes, ok := segments[constants.NormsField].(map[string]interface{}); ok {
		if memoryBytes, ok := normsBytes[constants.MemoryInBytesField].(float64); ok {
			stats.NormsMemoryBytes = int64(memoryBytes)
			stats.NormsMemory = models.FormatBytes(int64(memoryBytes))
			totalCalculatedMemory += int64(memoryBytes)
		}
	}

	if fixedBitSetBytes, ok := segments[constants.FixedBitSetMemoryInBytesField].(float64); ok {
		stats.FixedBitSetMemoryBytes = int64(fixedBitSetBytes)
		stats.FixedBitSetMemory = models.FormatBytes(int64(fixedBitSetBytes))
		totalCalculatedMemory += int64(fixedBitSetBytes)
	}

	if versionMapBytes, ok := segments[constants.VersionMapMemoryInBytesField].(float64); ok {
		stats.VersionMapMemoryBytes = int64(versionMapBytes)
		stats.VersionMapMemory = models.FormatBytes(int64(versionMapBytes))
		totalCalculatedMemory += int64(versionMapBytes)
	}

	if maxUnsafeAutoIDTimestamp, ok := segments[constants.MaxUnsafeAutoIDTimestampField].(float64); ok {
		stats.MaxUnsafeAutoIDTimestamp = int64(maxUnsafeAutoIDTimestamp)
	}

	if indexMemory, ok := indexing[constants.IndexMemoryField].(map[string]interface{}); ok {
		if totalBytes, ok := indexMemory[constants.TotalInBytesField].(float64); ok {
			stats.IndexMemoryBytes = int64(totalBytes)
			stats.IndexMemory = models.FormatBytes(int64(totalBytes))
		}
	}

	if stats.TermsMemory == constants.EmptyString {
		stats.TermsMemory = constants.ZeroByteString
	}
	if stats.StoredMemory == constants.EmptyString {
		stats.StoredMemory = constants.ZeroByteString
	}
	if stats.DocValuesMemory == constants.EmptyString {
		stats.DocValuesMemory = constants.ZeroByteString
	}
	if stats.PointsMemory == constants.EmptyString {
		stats.PointsMemory = constants.ZeroByteString
	}
	if stats.NormsMemory == constants.EmptyString {
		stats.NormsMemory = constants.ZeroByteString
	}
	if stats.FixedBitSetMemory == constants.EmptyString {
		stats.FixedBitSetMemory = constants.ZeroByteString
	}
	if stats.VersionMapMemory == constants.EmptyString {
		stats.VersionMapMemory = constants.ZeroByteString
	}

	if stats.SegmentMemoryBytes == 0 && totalCalculatedMemory > 0 {
		stats.SegmentMemoryBytes = totalCalculatedMemory
		stats.SegmentMemory = models.FormatBytes(totalCalculatedMemory)
	}

	return stats
}
