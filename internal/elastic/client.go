package elastic

import (
	"log"
	"net/http"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/mertbahardogan/escope/internal/constants"
)

func NewClient(host, username, password string, secure bool) *elasticsearch.Client {
	if host == "" {
		log.Fatalf("Failed to create Elasticsearch client: host is required")
	}

	cfg := elasticsearch.Config{
		Addresses: []string{host},
	}

	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}
	return client
}

func NewRecordSamplerClient(host, username, password string, secure bool) *elasticsearch.Client {
	if host == "" {
		log.Fatalf("Failed to create Elasticsearch client: host is required")
	}
	tr, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		log.Fatalf("Failed to clone default HTTP transport")
	}
	tp := tr.Clone()
	tp.ResponseHeaderTimeout = time.Duration(constants.RecordHTTPResponseHeaderTimeoutSecs) * time.Second
	cfg := elasticsearch.Config{
		Addresses: []string{host},
		Transport: tp,
	}
	if username != "" && password != "" {
		cfg.Username = username
		cfg.Password = password
	}
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Elasticsearch client: %v", err)
	}
	return client
}
