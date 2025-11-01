package elastic

import (
	"context"
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mertbahardogan/escope/internal/constants"
	"github.com/mertbahardogan/escope/internal/interfaces"
	"github.com/mertbahardogan/escope/internal/util"
	"io"
	"strings"
)

type ClientWrapper struct {
	client *elasticsearch.Client
}

func NewClientWrapper(client *elasticsearch.Client) interfaces.ElasticClient {
	return &ClientWrapper{client: client}
}

func (cw *ClientWrapper) GetClusterHealth(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cluster.Health(cw.client.Cluster.Health.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetClusterStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cluster.Stats(cw.client.Cluster.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetNodes(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Nodes.Info(cw.client.Nodes.Info.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetNodesInfo(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Nodes.Info(cw.client.Nodes.Info.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetNodesStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Nodes.Stats(cw.client.Nodes.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
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
	defer res.Body.Close()

	var indices []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&indices); err != nil {
		return nil, err
	}

	aliasesRes, err := cw.client.Cat.Aliases(
		cw.client.Cat.Aliases.WithContext(ctx),
		cw.client.Cat.Aliases.WithFormat("json"),
	)
	if err != nil {
		return nil, err
	}
	defer aliasesRes.Body.Close()

	var aliases []map[string]interface{}
	if err := json.NewDecoder(aliasesRes.Body).Decode(&aliases); err != nil {
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

	var processedIndices []map[string]interface{}
	for _, idx := range indices {
		indexName := util.GetStringField(idx, "index")
		alias := indexAliases[indexName]
		if alias == "" {
			alias = constants.DashString
		}

		processedIndex := map[string]interface{}{
			"health":     util.GetStringField(idx, "health"),
			"status":     util.GetStringField(idx, "status"),
			"index":      indexName,
			"docs.count": util.GetStringField(idx, "docs.count"),
			"store.size": util.GetStringField(idx, "store.size"),
			"pri":        util.GetStringField(idx, "pri"),
			"rep":        util.GetStringField(idx, "rep"),
			"alias":      alias,
		}
		processedIndices = append(processedIndices, processedIndex)
	}

	return map[string]interface{}{
		"": processedIndices,
	}, nil
}

func (cw *ClientWrapper) GetIndexStats(ctx context.Context, indexName string) (map[string]interface{}, error) {
	res, err := cw.client.Indices.Stats(cw.client.Indices.Stats.WithIndex(indexName), cw.client.Indices.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetShards(ctx context.Context) (map[string]interface{}, error) {
	// Use format=json to get structured data and avoid language issues
	res, err := cw.client.Cat.Shards(
		cw.client.Cat.Shards.WithContext(ctx),
		cw.client.Cat.Shards.WithFormat("json"),
		cw.client.Cat.Shards.WithV(true), // verbose mode to get all fields
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var shards []map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&shards); err != nil {
		return nil, err
	}

	var processedShards []map[string]interface{}
	for _, shard := range shards {
		processedShard := map[string]interface{}{
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
			processedShard["node_name"] = nodeName
		}

		processedShards = append(processedShards, processedShard)
	}

	return map[string]interface{}{
		"": processedShards,
	}, nil
}

func (cw *ClientWrapper) GetLuceneStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Indices.Stats(cw.client.Indices.Stats.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetSegments(ctx context.Context) (map[string]interface{}, error) {
	res, err := cw.client.Cat.Segments(
		cw.client.Cat.Segments.WithContext(ctx),
		cw.client.Cat.Segments.WithFormat("json"),
		cw.client.Cat.Segments.WithBytes("b"), // get sizes in bytes
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var segments []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &segments); err != nil {
		return nil, err
	}

	var processedSegments []map[string]interface{}
	for _, segment := range segments {
		processedSegment := map[string]interface{}{
			"index": util.GetStringField(segment, "index"),
			"size":  util.GetStringField(segment, "size"),
		}
		processedSegments = append(processedSegments, processedSegment)
	}

	return map[string]interface{}{
		"segments": processedSegments,
	}, nil
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
	requestBody := map[string]interface{}{
		"fields": fields,
	}

	bodyBytes, err := json.Marshal(requestBody)
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
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func (cw *ClientWrapper) GetIndicesWithSort(ctx context.Context, sortBy, sortOrder string) ([]map[string]interface{}, error) {
	useClientSideSort := sortBy == "alias"

	var sortParam string
	if useClientSideSort {
		sortParam = "index"
	} else {
		sortParam = buildSortParam(sortBy, sortOrder)
	}

	indices, err := cw.makeIndicesRequest(ctx, sortParam)
	if err != nil {
		return nil, err
	}

	// Fetch aliases like in the regular GetIndices method
	aliasesRes, err := cw.client.Cat.Aliases(
		cw.client.Cat.Aliases.WithContext(ctx),
		cw.client.Cat.Aliases.WithFormat("json"),
	)
	if err != nil {
		return nil, err
	}
	defer aliasesRes.Body.Close()

	var aliases []map[string]interface{}
	if err := json.NewDecoder(aliasesRes.Body).Decode(&aliases); err != nil {
		return nil, err
	}

	// Build index to aliases mapping
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

	var processedIndices []map[string]interface{}
	for _, idx := range indices {
		indexName := util.GetStringField(idx, "index")
		alias := indexAliases[indexName]
		if alias == "" {
			alias = constants.DashString
		}

		processedIndex := map[string]interface{}{
			"health":     util.GetStringField(idx, "health"),
			"status":     util.GetStringField(idx, "status"),
			"index":      indexName,
			"docs.count": util.GetStringField(idx, "docs.count"),
			"store.size": util.GetStringField(idx, "store.size"),
			"pri":        util.GetStringField(idx, "pri"),
			"rep":        util.GetStringField(idx, "rep"),
			"alias":      alias,
		}
		processedIndices = append(processedIndices, processedIndex)
	}

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
	sortParam := buildSortParam(sortBy, sortOrder)

	shards, err := cw.makeShardsRequest(ctx, sortParam)
	if err != nil {
		return nil, err
	}

	var processedShards []map[string]interface{}
	for _, shard := range shards {
		processedShard := map[string]interface{}{
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
			processedShard["node_name"] = nodeName
		}

		processedShards = append(processedShards, processedShard)
	}
	return processedShards, nil
}

func (cw *ClientWrapper) GetAnalyze(ctx context.Context, analyzerName, text string, analyzeType string) (map[string]interface{}, error) {
	var requestBody map[string]interface{}

	if analyzeType == "analyzer" {
		requestBody = map[string]interface{}{
			"analyzer": analyzerName,
			"text":     text,
		}
	} else if analyzeType == "tokenizer" {
		requestBody = map[string]interface{}{
			"tokenizer": analyzerName,
			"text":      text,
		}
	} else {
		requestBody = map[string]interface{}{
			"text": text,
		}
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
	defer res.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
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
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
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
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}
