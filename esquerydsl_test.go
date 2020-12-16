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
