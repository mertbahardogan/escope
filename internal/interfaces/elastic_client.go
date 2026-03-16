package interfaces

import (
	"context"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticClient interface {
	GetClusterHealth(ctx context.Context) (map[string]interface{}, error)
	GetClusterStats(ctx context.Context) (map[string]interface{}, error)

	GetNodes(ctx context.Context) (map[string]interface{}, error)
	GetNodesInfo(ctx context.Context) (map[string]interface{}, error)
	GetNodesStats(ctx context.Context) (map[string]interface{}, error)

	GetIndices(ctx context.Context) (map[string]interface{}, error)
	GetIndicesWithSort(ctx context.Context, sortBy, sortOrder string) ([]map[string]interface{}, error)
	GetIndexStats(ctx context.Context, indexName string) (map[string]interface{}, error)

	GetShards(ctx context.Context) (map[string]interface{}, error)
	GetShardsWithSort(ctx context.Context, sortBy, sortOrder string) ([]map[string]interface{}, error)

	GetLuceneStats(ctx context.Context) (map[string]interface{}, error)
	GetSegments(ctx context.Context) (map[string]interface{}, error)

	GetTermvectors(ctx context.Context, indexName, documentID string, fields []string) (map[string]interface{}, error)

	GetAnalyze(ctx context.Context, analyzerName, text string, analyzeType string) (map[string]interface{}, error)

	GetIndexMapping(ctx context.Context, indexName string) (map[string]interface{}, error)
	GetIndexSettings(ctx context.Context, indexName string) (map[string]interface{}, error)

	Ping(ctx context.Context) error
	GetClient() *elasticsearch.Client
}
