package services

import (
	"context"
	"fmt"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/models"
)

type TermvectorsService interface {
	GetDocumentTermvectors(ctx context.Context, indexName, documentID string, fields []string) ([]models.TermInfo, error)
}

type termvectorsService struct {
	client interfaces.ElasticClient
}

func NewTermvectorsService(client interfaces.ElasticClient) TermvectorsService {
	return &termvectorsService{
		client: client,
	}
}

func (s *termvectorsService) GetDocumentTermvectors(ctx context.Context, indexName, documentID string, fields []string) ([]models.TermInfo, error) {
	result, err := s.client.GetTermvectors(ctx, indexName, documentID, fields)
	if err != nil {
		return nil, fmt.Errorf("termvectors request failed: %w", err)
	}

	if found, ok := result["found"].(bool); ok && !found {
		return nil, fmt.Errorf("document '%s' not found in index '%s'", documentID, indexName)
	}

	var termInfos []models.TermInfo
	var termVectors map[string]interface{}

	if docs, ok := result["docs"].(map[string]interface{}); ok {
		if tv, ok := docs["term_vectors"].(map[string]interface{}); ok {
			termVectors = tv
		}
	} else if tv, ok := result["term_vectors"].(map[string]interface{}); ok {
		termVectors = tv
	}

	if termVectors != nil {
		for fieldName, fieldData := range termVectors {
			if field, ok := fieldData.(map[string]interface{}); ok {
				if terms, ok := field["terms"].(map[string]interface{}); ok {
					for termName, termData := range terms {
						if term, ok := termData.(map[string]interface{}); ok {
							termInfo := s.parseTermInfo(fieldName, termName, term)
							termInfos = append(termInfos, termInfo)
						}
					}
				}
			}
		}
	}

	return termInfos, nil
}

func (s *termvectorsService) parseTermInfo(fieldName, termName string, termData map[string]interface{}) models.TermInfo {
	termInfo := models.TermInfo{
		Field: fieldName,
		Term:  termName,
	}

	if termFreq, ok := termData["term_freq"].(float64); ok {
		termInfo.TermFreq = int(termFreq)
	}

	return termInfo
}
