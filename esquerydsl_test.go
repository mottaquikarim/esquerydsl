package esquerydsl

import (
	"testing"
)

func TestAndQuery(t *testing.T) {
	_, body, _ := GetQueryBlock(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{map[string]string{"id": "asc"}},
		And: []QueryItem{
			QueryItem{
				Field: "some_index_id",
				Value: "some-long-key-id-value",
				Type:  "match",
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}},"sort":[{"id":"asc"}]}`
	if body != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, body)
	}
}

func TestFilterQuery(t *testing.T) {
	_, body, _ := GetQueryBlock(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			QueryItem{
				Field: "title",
				Value: "Search",
				Type:  "match",
			},
			QueryItem{
				Field: "content",
				Value: "Elasticsearch",
				Type:  "match",
			},
		},
		Filter: []QueryItem{
			QueryItem{
				Field: "status",
				Value: "published",
				Type:  "term",
			},
			QueryItem{
				Field: "publish_date",
				Value: map[string]string{
					"gte": "2015-01-01",
				},
				Type: "range",
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"match":{"title":"Search"}},{"match":{"content":"Elasticsearch"}}],"filter":[{"term":{"status":"published"}},{"range":{"publish_date":{"gte":"2015-01-01"}}}]}}}`
	if body != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, body)
	}
}
