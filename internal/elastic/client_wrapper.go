package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/util"
)

type ClientWrapper struct {
	client *elasticsearch.Client
}

func NewClientWrapper(client *elasticsearch.Client) interfaces.ElasticClient {
	return &ClientWrapper{client: client}
}

// decodeJSONResponse reads and decodes JSON from body, closing it when done.
func decodeJSONResponse(body io.ReadCloser, v interface{}) error {
	defer body.Close()
	return json.NewDecoder(body).Decode(v)
}

// checkElasticsearchError extracts error from ES response if present.
func checkElasticsearchError(result map[string]interface{}) error {
	if errorData, ok := result["error"].(map[string]interface{}); ok {
		if reason, ok := errorData["reason"].(string); ok {
			return fmt.Errorf("%s", reason)
		}
		return fmt.Errorf("elasticsearch error")
	}
	return nil
}

func (cw *ClientWrapper) GetClusterHealth(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cluster.Health(cw.client.Cluster.Health.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetClusterStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cluster.Stats(cw.client.Cluster.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetNodes(ctx context.Context) (map[string]interface{}, error) {
	return cw.GetNodesInfo(ctx)
}

func (cw *ClientWrapper) GetNodesInfo(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Nodes.Info(cw.client.Nodes.Info.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetNodesStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Nodes.Stats(cw.client.Nodes.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) fetchIndexAliases(ctx context.Context) (map[string]string, error) {
	res, err := cw.client.Cat.Aliases(
		cw.client.Cat.Aliases.WithContext(ctx),
		cw.client.Cat.Aliases.WithFormat("json"),
	)
	if err != nil {
		return nil, err
	}
	var aliases []map[string]interface{}
	if err := decodeJSONResponse(res.Body, &aliases); err != nil {
		return nil, err
	}
	indexAliases := make(map[string]string)
	for _, alias := range aliases {
		if index := util.GetStringField(alias, "index"); index != "" {
			if aliasName := util.GetStringField(alias, "alias"); aliasName != "" {
				if existing, exists := indexAliases[index]; exists {
					indexAliases[index] = existing + "," + aliasName
				} else {
					indexAliases[index] = aliasName
				}
			}
		}
	}
	return indexAliases, nil
}

func processIndicesWithAliases(indices []map[string]interface{}, indexAliases map[string]string) []map[string]interface{} {
	processed := make([]map[string]interface{}, 0, len(indices))
	for _, idx := range indices {
		indexName := util.GetStringField(idx, "index")
		alias := indexAliases[indexName]
		if alias == "" {
			alias = constants.DashString
		}
		processed = append(processed, map[string]interface{}{
			"health":     util.GetStringField(idx, "health"),
			"status":     util.GetStringField(idx, "status"),
			"index":      indexName,
			"docs.count": util.GetStringField(idx, "docs.count"),
			"store.size": util.GetStringField(idx, "store.size"),
			"pri":        util.GetStringField(idx, "pri"),
			"rep":        util.GetStringField(idx, "rep"),
			"alias":      alias,
		})
	}
	return processed
}

func (cw *ClientWrapper) GetIndices(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cat.Indices(
		cw.client.Cat.Indices.WithContext(ctx),
		cw.client.Cat.Indices.WithFormat("json"),
		cw.client.Cat.Indices.WithV(true),
	)
	if err != nil {
		return nil, err
	}
	var indices []map[string]interface{}
	if err := decodeJSONResponse(res.Body, &indices); err != nil {
		return nil, err
	}
	indexAliases, err := cw.fetchIndexAliases(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"": processIndicesWithAliases(indices, indexAliases)}, nil
}

func processShards(shards []map[string]interface{}) []map[string]interface{} {
	processed := make([]map[string]interface{}, 0, len(shards))
	for _, shard := range shards {
		p := map[string]interface{}{
			"index":  util.GetStringField(shard, "index"),
			"shard":  util.GetStringField(shard, "shard"),
			"prirep": util.GetStringField(shard, "prirep"),
			"state":  util.GetStringField(shard, "state"),
			"docs":   util.GetStringField(shard, "docs"),
			"store":  util.GetStringField(shard, "store"),
			"ip":     util.GetStringField(shard, "ip"),
			"node":   util.GetStringField(shard, "node"),
		}
		if nodeName := util.GetStringField(shard, "node_name"); nodeName != "" {
			p["node_name"] = nodeName
		}
		processed = append(processed, p)
	}
	return processed
}

func (cw *ClientWrapper) GetIndexStats(ctx context.Context, indexName string) (map[string]interface{}, error) {
	res, err := cw.client.Indices.Stats(cw.client.Indices.Stats.WithIndex(indexName), cw.client.Indices.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetShards(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cat.Shards(
		cw.client.Cat.Shards.WithContext(ctx),
		cw.client.Cat.Shards.WithFormat("json"),
		cw.client.Cat.Shards.WithV(true),
	)
	if err != nil {
		return nil, err
	}
	var shards []map[string]interface{}
	if err := decodeJSONResponse(res.Body, &shards); err != nil {
		return nil, err
	}
	return map[string]interface{}{"": processShards(shards)}, nil
}

func (cw *ClientWrapper) GetLuceneStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Indices.Stats(cw.client.Indices.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetSegments(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cat.Segments(
		cw.client.Cat.Segments.WithContext(ctx),
		cw.client.Cat.Segments.WithFormat("json"),
		cw.client.Cat.Segments.WithBytes("b"),
	)
	if err != nil {
		return nil, err
	}
	var segments []map[string]interface{}
	if err := decodeJSONResponse(res.Body, &segments); err != nil {
		return nil, err
	}
	processed := make([]map[string]interface{}, 0, len(segments))
	for _, s := range segments {
		processed = append(processed, map[string]interface{}{
			"index": util.GetStringField(s, "index"),
			"size":  util.GetStringField(s, "size"),
		})
	}
	return map[string]interface{}{"segments": processed}, nil
}

func (cw *ClientWrapper) Ping(ctx context.Context) error {
	res, err := cw.client.Ping(cw.client.Ping.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (cw *ClientWrapper) GetClient() *elasticsearch.Client {
	return cw.client
}

func (cw *ClientWrapper) GetTermvectors(ctx context.Context, indexName, documentID string, fields []string) (map[string]interface{}, error) {
	bodyBytes, err := json.Marshal(map[string]interface{}{"fields": fields})
	if err != nil {
		return nil, err
	}
	res, err := cw.client.Termvectors(
		indexName,
		cw.client.Termvectors.WithDocumentID(documentID),
		cw.client.Termvectors.WithBody(strings.NewReader(string(bodyBytes))),
		cw.client.Termvectors.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	if err := checkElasticsearchError(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetIndicesWithSort(ctx context.Context, sortBy, sortOrder string) ([]map[string]interface{}, error) {
	useClientSideSort := sortBy == "alias"
	sortParam := "index"
	if !useClientSideSort {
		sortParam = buildSortParam(sortBy, sortOrder)
	}

	indices, err := cw.makeIndicesRequest(ctx, sortParam)
	if err != nil {
		return nil, err
	}
	indexAliases, err := cw.fetchIndexAliases(ctx)
	if err != nil {
		return nil, err
	}
	processedIndices := processIndicesWithAliases(indices, indexAliases)

	if useClientSideSort {
		if sortOrder == "desc" {
			for i := 0; i < len(processedIndices)-1; i++ {
				for j := i + 1; j < len(processedIndices); j++ {
					aliasI := util.GetStringField(processedIndices[i], "alias")
					aliasJ := util.GetStringField(processedIndices[j], "alias")
					if aliasI < aliasJ {
						processedIndices[i], processedIndices[j] = processedIndices[j], processedIndices[i]
					}
				}
			}
		} else {
			for i := 0; i < len(processedIndices)-1; i++ {
				for j := i + 1; j < len(processedIndices); j++ {
					aliasI := util.GetStringField(processedIndices[i], "alias")
					aliasJ := util.GetStringField(processedIndices[j], "alias")
					if aliasI > aliasJ {
						processedIndices[i], processedIndices[j] = processedIndices[j], processedIndices[i]
					}
				}
			}
		}
	}

	return processedIndices, nil
}

func (cw *ClientWrapper) GetShardsWithSort(ctx context.Context, sortBy, sortOrder string) ([]map[string]interface{}, error) {
	shards, err := cw.makeShardsRequest(ctx, buildSortParam(sortBy, sortOrder))
	if err != nil {
		return nil, err
	}
	return processShards(shards), nil
}

func (cw *ClientWrapper) GetAnalyze(ctx context.Context, analyzerName, text string, analyzeType string) (map[string]interface{}, error) {
	var requestBody map[string]interface{}
	switch analyzeType {
	case "analyzer":
		requestBody = map[string]interface{}{"analyzer": analyzerName, "text": text}
	case "tokenizer":
		requestBody = map[string]interface{}{"tokenizer": analyzerName, "text": text}
	default:
		requestBody = map[string]interface{}{"text": text}
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}
	res, err := cw.client.Indices.Analyze(
		cw.client.Indices.Analyze.WithBody(strings.NewReader(string(bodyBytes))),
		cw.client.Indices.Analyze.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func buildSortParam(sortBy, sortOrder string) string {
	sortParam := sortBy
	if sortOrder != "" && sortOrder != "asc" {
		sortParam = sortBy + ":" + sortOrder
	}
	return sortParam
}

func (cw *ClientWrapper) GetIndexMapping(ctx context.Context, indexName string) (map[string]interface{}, error) {
	res, err := cw.client.Indices.GetMapping(
		cw.client.Indices.GetMapping.WithContext(ctx),
		cw.client.Indices.GetMapping.WithIndex(indexName),
	)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	if err := checkElasticsearchError(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) GetIndexSettings(ctx context.Context, indexName string) (map[string]interface{}, error) {
	res, err := cw.client.Indices.GetSettings(
		cw.client.Indices.GetSettings.WithContext(ctx),
		cw.client.Indices.GetSettings.WithIndex(indexName),
	)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	if err := checkElasticsearchError(result); err != nil {
		return nil, err
	}
	return result, nil
}

func (cw *ClientWrapper) CountWithBody(ctx context.Context, indexName string, body []byte) (int64, error) {
	res, err := cw.client.Count(
		cw.client.Count.WithContext(ctx),
		cw.client.Count.WithIndex(indexName),
		cw.client.Count.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return 0, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return 0, err
	}
	if res.IsError() {
		if err := checkElasticsearchError(result); err != nil {
			return 0, err
		}
		return 0, fmt.Errorf("count request failed: %s", res.Status())
	}
	if count, ok := result["count"].(float64); ok {
		return int64(count), nil
	}
	return 0, fmt.Errorf("unexpected count response")
}

func (cw *ClientWrapper) SearchWithBody(ctx context.Context, indexName string, body []byte) (map[string]interface{}, error) {
	res, err := cw.client.Search(
		cw.client.Search.WithContext(ctx),
		cw.client.Search.WithIndex(indexName),
		cw.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := decodeJSONResponse(res.Body, &result); err != nil {
		return nil, err
	}
	if res.IsError() {
		if err := checkElasticsearchError(result); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("search request failed: %s", res.Status())
	}
	return result, nil
}

func (cw *ClientWrapper) makeIndicesRequest(ctx context.Context, sortParam string) ([]map[string]interface{}, error) {
	res, err := cw.client.Cat.Indices(
		cw.client.Cat.Indices.WithContext(ctx),
		cw.client.Cat.Indices.WithFormat("json"),
		cw.client.Cat.Indices.WithS(sortParam),
		cw.client.Cat.Indices.WithV(true),
	)
	if err != nil {
		return nil, err
	}
	var data []map[string]interface{}
	if err := decodeJSONResponse(res.Body, &data); err != nil {
		return nil, err
	}
	return data, nil
}

func (cw *ClientWrapper) makeShardsRequest(ctx context.Context, sortParam string) ([]map[string]interface{}, error) {
	res, err := cw.client.Cat.Shards(
		cw.client.Cat.Shards.WithContext(ctx),
		cw.client.Cat.Shards.WithFormat("json"),
		cw.client.Cat.Shards.WithS(sortParam),
		cw.client.Cat.Shards.WithV(true),
	)
	if err != nil {
		return nil, err
	}
	var data []map[string]interface{}
	if err := decodeJSONResponse(res.Body, &data); err != nil {
		return nil, err
	}
	return data, nil
}
