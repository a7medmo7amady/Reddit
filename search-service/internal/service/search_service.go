package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"search-service/internal/model"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type SearchService interface {
	CreateIndex(ctx context.Context) error
	
	IndexPost(ctx context.Context, post model.Post) error
	IndexCommunity(ctx context.Context, community model.Community) error
	IndexUser(ctx context.Context, user model.User) error
	IndexComment(ctx context.Context, comment model.Comment) error
	
	SearchPosts(ctx context.Context, query string, filters map[string]interface{}, limit, page int) ([]model.Post, int64, error)
	SearchCommunities(ctx context.Context, query string, limit, page int) ([]model.Community, int64, error)
	SearchUsers(ctx context.Context, query string, limit, page int) ([]model.User, int64, error)
	
	SyncPosts(ctx context.Context, posts []model.Post) error
}

type searchService struct {
	esClient *elasticsearch.Client
}

func NewSearchService(esClient *elasticsearch.Client) SearchService {
	return &searchService{
		esClient: esClient,
	}
}

func (s *searchService) CreateIndex(ctx context.Context) error {
	indices := map[string]string{
		"posts": `{
			"mappings": {
				"properties": {
					"title": { "type": "text" },
					"content": { "type": "text" },
					"author_id": { "type": "keyword" },
					"author_name": { "type": "keyword" },
					"community_id": { "type": "keyword" },
					"community_name": { "type": "keyword" },
					"type": { "type": "keyword" },
					"flair": { "type": "keyword" },
					"nsfw": { "type": "boolean" },
					"spoiler": { "type": "boolean" },
					"oc": { "type": "boolean" },
					"created_at": { "type": "date" }
				}
			}
		}`,
		"communities": `{
			"mappings": {
				"properties": {
					"name": { "type": "text" },
					"description": { "type": "text" },
					"member_count": { "type": "integer" },
					"created_at": { "type": "date" }
				}
			}
		}`,
		"users": `{
			"mappings": {
				"properties": {
					"username": { "type": "text" },
					"karma": { "type": "integer" },
					"created_at": { "type": "date" }
				}
			}
		}`,
		"comments": `{
			"mappings": {
				"properties": {
					"content": { "type": "text" },
					"post_id": { "type": "keyword" },
					"author_name": { "type": "keyword" },
					"created_at": { "type": "date" }
				}
			}
		}`,
	}

	for name, mapping := range indices {
		req := esapi.IndicesCreateRequest{
			Index: name,
			Body:  bytes.NewReader([]byte(mapping)),
		}
		res, err := req.Do(ctx, s.esClient)
		if err != nil {
			log.Printf("Error creating index %s: %v", name, err)
			continue
		}
		res.Body.Close()
	}

	return nil
}

func (s *searchService) IndexPost(ctx context.Context, post model.Post) error {
	data, _ := json.Marshal(post)
	req := esapi.IndexRequest{Index: "posts", DocumentID: post.ID, Body: bytes.NewReader(data), Refresh: "true"}
	res, err := req.Do(ctx, s.esClient)
	if err != nil { return err }
	defer res.Body.Close()
	return nil
}

func (s *searchService) IndexCommunity(ctx context.Context, community model.Community) error {
	data, _ := json.Marshal(community)
	req := esapi.IndexRequest{Index: "communities", DocumentID: community.ID, Body: bytes.NewReader(data), Refresh: "true"}
	res, err := req.Do(ctx, s.esClient)
	if err != nil { return err }
	defer res.Body.Close()
	return nil
}

func (s *searchService) IndexUser(ctx context.Context, user model.User) error {
	data, _ := json.Marshal(user)
	req := esapi.IndexRequest{Index: "users", DocumentID: user.ID, Body: bytes.NewReader(data), Refresh: "true"}
	res, err := req.Do(ctx, s.esClient)
	if err != nil { return err }
	defer res.Body.Close()
	return nil
}

func (s *searchService) IndexComment(ctx context.Context, comment model.Comment) error {
	data, _ := json.Marshal(comment)
	req := esapi.IndexRequest{Index: "comments", DocumentID: comment.ID, Body: bytes.NewReader(data), Refresh: "true"}
	res, err := req.Do(ctx, s.esClient)
	if err != nil { return err }
	defer res.Body.Close()
	return nil
}

func (s *searchService) SearchPosts(ctx context.Context, query string, filters map[string]interface{}, limit, page int) ([]model.Post, int64, error) {
	if limit <= 0 { limit = 20 }
	from := (page - 1) * limit

	boolQuery := map[string]interface{}{
		"must": []map[string]interface{}{
			{
				"multi_match": map[string]interface{}{
					"query":  query,
					"fields": []string{"title", "content", "community_name", "author_name"},
				},
			},
		},
	}

	filterClauses := []map[string]interface{}{}
	if contentType, ok := filters["type"]; ok && contentType != "" {
		filterClauses = append(filterClauses, map[string]interface{}{"term": map[string]interface{}{"type": contentType}})
	}
	if communityID, ok := filters["community_id"]; ok && communityID != "" {
		filterClauses = append(filterClauses, map[string]interface{}{"term": map[string]interface{}{"community_id": communityID}})
	}
	if dateRange, ok := filters["date_range"]; ok && dateRange != "" {
		now := time.Now()
		var start time.Time
		switch dateRange {
		case "week":
			start = now.AddDate(0, 0, -7)
		case "month":
			start = now.AddDate(0, -1, 0)
		}
		if !start.IsZero() {
			filterClauses = append(filterClauses, map[string]interface{}{
				"range": map[string]interface{}{
					"created_at": map[string]interface{}{"gte": start},
				},
			})
		}
	}
	boolQuery["filter"] = filterClauses

	searchQuery := map[string]interface{}{
		"from": from,
		"size": limit,
		"query": map[string]interface{}{"bool": boolQuery},
		"sort": []map[string]interface{}{{"created_at": map[string]interface{}{"order": "desc"}}},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(searchQuery)

	res, err := s.esClient.Search(s.esClient.Search.WithIndex("posts"), s.esClient.Search.WithBody(&buf))
	if err != nil { return nil, 0, err }
	defer res.Body.Close()

	var r map[string]interface{}
	json.NewDecoder(res.Body).Decode(&r)

	hitsMap := r["hits"].(map[string]interface{})
	total := int64(hitsMap["total"].(map[string]interface{})["value"].(float64))
	hits := hitsMap["hits"].([]interface{})

	var posts []model.Post
	for _, hit := range hits {
		var post model.Post
		b, _ := json.Marshal(hit.(map[string]interface{})["_source"])
		json.Unmarshal(b, &post)
		posts = append(posts, post)
	}

	return posts, total, nil
}

func (s *searchService) SearchCommunities(ctx context.Context, query string, limit, page int) ([]model.Community, int64, error) {
	if limit <= 0 { limit = 20 }
	from := (page - 1) * limit

	searchQuery := map[string]interface{}{
		"from": from,
		"size": limit,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  query,
				"fields": []string{"name", "description"},
			},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(searchQuery)
	res, _ := s.esClient.Search(s.esClient.Search.WithIndex("communities"), s.esClient.Search.WithBody(&buf))
	defer res.Body.Close()
	var r map[string]interface{}
	json.NewDecoder(res.Body).Decode(&r)
	hitsMap := r["hits"].(map[string]interface{})
	total := int64(hitsMap["total"].(map[string]interface{})["value"].(float64))
	hits := hitsMap["hits"].([]interface{})
	var items []model.Community
	for _, hit := range hits {
		var item model.Community
		b, _ := json.Marshal(hit.(map[string]interface{})["_source"])
		json.Unmarshal(b, &item)
		items = append(items, item)
	}
	return items, total, nil
}

func (s *searchService) SearchUsers(ctx context.Context, query string, limit, page int) ([]model.User, int64, error) {
	if limit <= 0 { limit = 20 }
	from := (page - 1) * limit

	searchQuery := map[string]interface{}{
		"from": from,
		"size": limit,
		"query": map[string]interface{}{
			"match": map[string]interface{}{"username": query},
		},
	}

	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(searchQuery)
	res, _ := s.esClient.Search(s.esClient.Search.WithIndex("users"), s.esClient.Search.WithBody(&buf))
	defer res.Body.Close()
	var r map[string]interface{}
	json.NewDecoder(res.Body).Decode(&r)
	hitsMap := r["hits"].(map[string]interface{})
	total := int64(hitsMap["total"].(map[string]interface{})["value"].(float64))
	hits := hitsMap["hits"].([]interface{})
	var items []model.User
	for _, hit := range hits {
		var item model.User
		b, _ := json.Marshal(hit.(map[string]interface{})["_source"])
		json.Unmarshal(b, &item)
		items = append(items, item)
	}
	return items, total, nil
}

func (s *searchService) SyncPosts(ctx context.Context, posts []model.Post) error {
	for _, post := range posts {
		s.IndexPost(ctx, post)
	}
	return nil
}
