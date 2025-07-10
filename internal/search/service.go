package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pluggedin/mcp-analytics/internal/model"
)

const (
	serverIndexName = "mcp_servers"
	serverMapping = `{
		"mappings": {
			"properties": {
				"id": { "type": "keyword" },
				"name": { 
					"type": "text",
					"fields": {
						"keyword": { "type": "keyword" },
						"suggest": { "type": "completion" }
					}
				},
				"description": { "type": "text" },
				"author": { 
					"type": "text",
					"fields": {
						"keyword": { "type": "keyword" }
					}
				},
				"homepage": { "type": "keyword" },
				"source": { "type": "keyword" },
				"repository": { "type": "keyword" },
				"license": { "type": "keyword" },
				"categories": { "type": "keyword" },
				"packages": {
					"type": "nested",
					"properties": {
						"type": { "type": "keyword" },
						"name": { "type": "keyword" },
						"version": { "type": "keyword" }
					}
				},
				"version_detail": {
					"properties": {
						"version": { "type": "keyword" },
						"sdk_version": { "type": "keyword" },
						"protocol_version": { "type": "keyword" }
					}
				},
				"remotes": {
					"type": "nested",
					"properties": {
						"type": { "type": "keyword" },
						"transport": { "type": "keyword" },
						"command": { "type": "text" },
						"args": { "type": "text" },
						"url": { "type": "keyword" },
						"headers": { "type": "object" }
					}
				},
				"tools": {
					"type": "nested",
					"properties": {
						"name": { "type": "keyword" },
						"description": { "type": "text" }
					}
				},
				"prompts": {
					"type": "nested",
					"properties": {
						"name": { "type": "keyword" },
						"description": { "type": "text" }
					}
				},
				"templates": {
					"type": "nested",
					"properties": {
						"name": { "type": "keyword" },
						"description": { "type": "text" }
					}
				},
				"indexed_at": { "type": "date" },
				"last_updated": { "type": "date" },
				"install_count": { "type": "long" },
				"rating_average": { "type": "float" },
				"rating_count": { "type": "long" },
				"popularity_score": { "type": "float" },
				"trending_score": { "type": "float" },
				"quality_score": { "type": "float" }
			}
		}
	}`
)

// Service handles Elasticsearch operations
type Service struct {
	client *elasticsearch.Client
}

// NewService creates a new search service
func NewService(esURL string) (*Service, error) {
	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Elasticsearch client: %w", err)
	}

	// Test connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("Elasticsearch error: %s", res.String())
	}

	log.Println("Connected to Elasticsearch")

	service := &Service{client: client}

	// Initialize index
	if err := service.initializeIndex(); err != nil {
		return nil, fmt.Errorf("failed to initialize index: %w", err)
	}

	return service, nil
}

// initializeIndex creates the index with proper mappings if it doesn't exist
func (s *Service) initializeIndex() error {
	// Check if index exists
	res, err := s.client.Indices.Exists([]string{serverIndexName})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	// If index doesn't exist, create it
	if res.StatusCode == 404 {
		res, err := s.client.Indices.Create(
			serverIndexName,
			s.client.Indices.Create.WithBody(strings.NewReader(serverMapping)),
		)
		if err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
		defer res.Body.Close()

		if res.IsError() {
			return fmt.Errorf("failed to create index: %s", res.String())
		}

		log.Printf("Created index: %s", serverIndexName)
	} else {
		log.Printf("Index already exists: %s", serverIndexName)
	}

	return nil
}

// IndexServer indexes a server document
func (s *Service) IndexServer(ctx context.Context, server *model.ServerDetail) error {
	// Prepare document
	doc, err := json.Marshal(server)
	if err != nil {
		return fmt.Errorf("failed to marshal server: %w", err)
	}

	// Index document
	req := esapi.IndexRequest{
		Index:      serverIndexName,
		DocumentID: server.ID,
		Body:       bytes.NewReader(doc),
		Refresh:    "true",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index document: %s", res.String())
	}

	return nil
}

// GetServer retrieves a server by ID
func (s *Service) GetServer(ctx context.Context, id string) (*model.ServerDetail, error) {
	req := esapi.GetRequest{
		Index:      serverIndexName,
		DocumentID: id,
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			return nil, fmt.Errorf("server not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get document: %s", res.String())
	}

	// Parse response
	var result struct {
		Source model.ServerDetail `json:"_source"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Source, nil
}

// DeleteServer deletes a server from the index
func (s *Service) DeleteServer(ctx context.Context, id string) error {
	req := esapi.DeleteRequest{
		Index:      serverIndexName,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(ctx, s.client)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		if res.StatusCode == 404 {
			// Already deleted, not an error
			return nil
		}
		return fmt.Errorf("failed to delete document: %s", res.String())
	}

	return nil
}

// Search performs a search query
func (s *Service) Search(ctx context.Context, query SearchQuery) (*SearchResult, error) {
	// Build Elasticsearch query
	esQuery := s.buildESQuery(query)

	// Prepare request body
	body, err := json.Marshal(esQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// Execute search
	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(serverIndexName),
		s.client.Search.WithBody(bytes.NewReader(body)),
		s.client.Search.WithFrom(query.Offset),
		s.client.Search.WithSize(query.Limit),
		s.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("search error: %s", res.String())
	}

	// Parse response
	var esResult struct {
		Hits struct {
			Total struct {
				Value int `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source model.ServerDetail `json:"_source"`
				Score  float64           `json:"_score"`
			} `json:"hits"`
		} `json:"hits"`
		Aggregations map[string]interface{} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Build result
	result := &SearchResult{
		Total:   esResult.Hits.Total.Value,
		Servers: make([]model.ServerDetail, len(esResult.Hits.Hits)),
		Facets:  s.parseFacets(esResult.Aggregations),
	}

	for i, hit := range esResult.Hits.Hits {
		result.Servers[i] = hit.Source
		result.Servers[i].Score = hit.Score
	}

	return result, nil
}

// buildESQuery builds an Elasticsearch query from search parameters
func (s *Service) buildESQuery(query SearchQuery) map[string]interface{} {
	// Base query structure
	esQuery := map[string]interface{}{
		"query": map[string]interface{}{},
		"aggs":  map[string]interface{}{},
		"sort":  []interface{}{},
	}

	// Build bool query
	boolQuery := map[string]interface{}{
		"must":   []interface{}{},
		"filter": []interface{}{},
	}

	// Add text search
	if query.Query != "" {
		boolQuery["must"] = append(boolQuery["must"].([]interface{}), map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query.Query,
				"fields": []string{"name^3", "description^2", "author", "categories"},
				"type":   "best_fields",
			},
		})
	}

	// Add filters
	for field, value := range query.Filters {
		switch field {
		case "package_type":
			boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
				"nested": map[string]interface{}{
					"path": "packages",
					"query": map[string]interface{}{
						"term": map[string]interface{}{
							"packages.type": value,
						},
					},
				},
			})
		case "transport":
			boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
				"nested": map[string]interface{}{
					"path": "remotes",
					"query": map[string]interface{}{
						"term": map[string]interface{}{
							"remotes.transport": value,
						},
					},
				},
			})
		default:
			boolQuery["filter"] = append(boolQuery["filter"].([]interface{}), map[string]interface{}{
				"term": map[string]interface{}{
					field: value,
				},
			})
		}
	}

	// Set query
	if len(boolQuery["must"].([]interface{})) > 0 || len(boolQuery["filter"].([]interface{})) > 0 {
		esQuery["query"] = map[string]interface{}{
			"bool": boolQuery,
		}
	} else {
		esQuery["query"] = map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	// Add sorting
	switch query.Sort {
	case "popularity":
		esQuery["sort"] = []interface{}{
			map[string]interface{}{"popularity_score": map[string]interface{}{"order": "desc"}},
		}
	case "trending":
		esQuery["sort"] = []interface{}{
			map[string]interface{}{"trending_score": map[string]interface{}{"order": "desc"}},
		}
	case "rating":
		esQuery["sort"] = []interface{}{
			map[string]interface{}{"rating_average": map[string]interface{}{"order": "desc"}},
		}
	case "recent":
		esQuery["sort"] = []interface{}{
			map[string]interface{}{"last_updated": map[string]interface{}{"order": "desc"}},
		}
	default:
		// Default to relevance (no explicit sort)
	}

	// Add aggregations for facets
	esQuery["aggs"] = map[string]interface{}{
		"categories": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "categories",
				"size":  20,
			},
		},
		"package_types": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "packages",
			},
			"aggs": map[string]interface{}{
				"types": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "packages.type",
						"size":  10,
					},
				},
			},
		},
		"transports": map[string]interface{}{
			"nested": map[string]interface{}{
				"path": "remotes",
			},
			"aggs": map[string]interface{}{
				"types": map[string]interface{}{
					"terms": map[string]interface{}{
						"field": "remotes.transport",
						"size":  10,
					},
				},
			},
		},
	}

	return esQuery
}

// MigrateServerSources updates existing servers to populate the source field
func (s *Service) MigrateServerSources(ctx context.Context) error {
	// Search for all servers without source field or with empty source
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []interface{}{
					map[string]interface{}{
						"bool": map[string]interface{}{
							"must_not": map[string]interface{}{
								"exists": map[string]interface{}{
									"field": "source",
								},
							},
						},
					},
					map[string]interface{}{
						"term": map[string]interface{}{
							"source": "",
						},
					},
				},
			},
		},
		"size": 1000,
	}

	body, err := json.Marshal(query)
	if err != nil {
		return fmt.Errorf("failed to marshal query: %w", err)
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(ctx),
		s.client.Search.WithIndex(serverIndexName),
		s.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return fmt.Errorf("failed to search servers: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("search error: %s", res.String())
	}

	var result struct {
		Hits struct {
			Hits []struct {
				ID     string              `json:"_id"`
				Source model.ServerDetail `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	// Update each server with the source field
	for _, hit := range result.Hits.Hits {
		server := hit.Source
		
		// Determine source based on name pattern
		if strings.HasPrefix(server.Name, "io.github.") {
			server.Source = "github"
		} else if strings.Contains(server.Name, "community") || strings.Contains(server.ID, "community") {
			server.Source = "community"
		} else if strings.Contains(server.Name, "private") || strings.Contains(server.ID, "private") {
			server.Source = "private"
		} else {
			server.Source = "github" // Default for io.github.* servers
		}

		// Update the server in Elasticsearch
		if err := s.IndexServer(ctx, &server); err != nil {
			log.Printf("Failed to update server %s: %v", server.ID, err)
			continue
		}
		
		log.Printf("Updated server %s with source: %s", server.ID, server.Source)
	}

	return nil
}

// parseFacets extracts facet data from aggregations
func (s *Service) parseFacets(aggs map[string]interface{}) []Facet {
	facets := []Facet{}

	// Parse categories
	if catAgg, ok := aggs["categories"].(map[string]interface{}); ok {
		if buckets, ok := catAgg["buckets"].([]interface{}); ok {
			catFacet := Facet{
				Field:  "categories",
				Values: make([]FacetValue, 0, len(buckets)),
			}
			for _, bucket := range buckets {
				if b, ok := bucket.(map[string]interface{}); ok {
					catFacet.Values = append(catFacet.Values, FacetValue{
						Value: b["key"].(string),
						Count: int(b["doc_count"].(float64)),
					})
				}
			}
			if len(catFacet.Values) > 0 {
				facets = append(facets, catFacet)
			}
		}
	}

	// Parse package types
	if pkgAgg, ok := aggs["package_types"].(map[string]interface{}); ok {
		if typesAgg, ok := pkgAgg["types"].(map[string]interface{}); ok {
			if buckets, ok := typesAgg["buckets"].([]interface{}); ok {
				pkgFacet := Facet{
					Field:  "package_type",
					Values: make([]FacetValue, 0, len(buckets)),
				}
				for _, bucket := range buckets {
					if b, ok := bucket.(map[string]interface{}); ok {
						pkgFacet.Values = append(pkgFacet.Values, FacetValue{
							Value: b["key"].(string),
							Count: int(b["doc_count"].(float64)),
						})
					}
				}
				if len(pkgFacet.Values) > 0 {
					facets = append(facets, pkgFacet)
				}
			}
		}
	}

	// Parse transports
	if transAgg, ok := aggs["transports"].(map[string]interface{}); ok {
		if typesAgg, ok := transAgg["types"].(map[string]interface{}); ok {
			if buckets, ok := typesAgg["buckets"].([]interface{}); ok {
				transFacet := Facet{
					Field:  "transport",
					Values: make([]FacetValue, 0, len(buckets)),
				}
				for _, bucket := range buckets {
					if b, ok := bucket.(map[string]interface{}); ok {
						transFacet.Values = append(transFacet.Values, FacetValue{
							Value: b["key"].(string),
							Count: int(b["doc_count"].(float64)),
						})
					}
				}
				if len(transFacet.Values) > 0 {
					facets = append(facets, transFacet)
				}
			}
		}
	}

	return facets
}

// SearchQuery represents search parameters
type SearchQuery struct {
	Query   string                 `json:"query"`
	Filters map[string]interface{} `json:"filters"`
	Sort    string                 `json:"sort"`
	Offset  int                    `json:"offset"`
	Limit   int                    `json:"limit"`
}

// SearchResult represents search results
type SearchResult struct {
	Total   int                  `json:"total"`
	Servers []model.ServerDetail `json:"servers"`
	Facets  []Facet             `json:"facets"`
}

// Facet represents a search facet
type Facet struct {
	Field  string       `json:"field"`
	Values []FacetValue `json:"values"`
}

// FacetValue represents a facet value
type FacetValue struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}