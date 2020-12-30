package esquerydsl

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestBogusQueryType(t *testing.T) {
	_, err := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{map[string]string{"id": "asc"}},
		And: []QueryItem{
			QueryItem{
				Field: "some_index_id",
				Value: "some-long-key-id-value",
				Type:  100001,
			},
		},
	})

	var queryTypeErr *QueryTypeErr
	if !errors.As(err, &queryTypeErr) {
		t.Errorf("\nUnexpected error: %v", err)
	}
}

func TestQueryStringEsc(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			QueryItem{
				Field: "user.id",
				Value: "kimchy!",
				Type:  QueryString,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"query_string":{"analyze_wildcard":true,"fields":["user.id"],"query":"kimchy\\!"}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestAndQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{map[string]string{"id": "asc"}},
		And: []QueryItem{
			QueryItem{
				Field: "some_index_id",
				Value: "some-long-key-id-value",
				Type:  Match,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}},"sort":[{"id":"asc"}]}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestFilterQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			QueryItem{
				Field: "title",
				Value: "Search",
				Type:  Match,
			},
			QueryItem{
				Field: "content",
				Value: "Elasticsearch",
				Type:  Match,
			},
		},
		Filter: []QueryItem{
			QueryItem{
				Field: "status",
				Value: "published",
				Type:  Term,
			},
			QueryItem{
				Field: "publish_date",
				Value: map[string]string{
					"gte": "2015-01-01",
				},
				Type: Range,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"match":{"title":"Search"}},{"match":{"content":"Elasticsearch"}}],"filter":[{"term":{"status":"published"}},{"range":{"publish_date":{"gte":"2015-01-01"}}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}
