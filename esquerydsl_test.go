package esquerydsl

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestBogusQueryType(t *testing.T) {
	_, err := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{{"id": "asc"}},
		And: []QueryItem{
			{
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
			{
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

func TestMultiSearchDoc(t *testing.T) {
	doc, _ := MultiSearchDoc([]QueryDoc{
		{
			Index: "index1",
			And: []QueryItem{
				{
					Field: "user.id",
					Value: "kimchy!",
					Type:  QueryString,
				},
			},
		},
		{
			Index: "index2",
			And: []QueryItem{
				{
					Field: "some_index_id",
					Value: "some-long-key-id-value",
					Type:  Match,
				},
			},
		},
	})

	expected := `{"index":"index1"}
{"query":{"bool":{"must":[{"query_string":{"analyze_wildcard":true,"fields":["user.id"],"query":"kimchy\\!"}}]}}}
{"index":"index2"}
{"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}}}
`
	if string(doc) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(doc))
	}
}

func TestAndQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{{"id": "asc"}},
		And: []QueryItem{
			{
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

func TestNotQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{{"id": "desc"}},
		Not: []QueryItem{
			{
				Field: "some_index_id",
				Value: "some-not-value",
				Type:  Match,
			},
		},
	})

	expected := `{"query":{"bool":{"must_not":[{"match":{"some_index_id":"some-not-value"}}]}},"sort":[{"id":"desc"}]}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestOrQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{{"id": "desc"}},
		Or: []QueryItem{
			{
				Field: "some_index_id",
				Value: "some-option-one",
				Type:  Match,
			},
			{
				Field: "some_index_id",
				Value: "some-option-two",
				Type:  Match,
			},
		},
	})

	expected := `{"query":{"bool":{"should":[{"match":{"some_index_id":"some-option-one"}},{"match":{"some_index_id":"some-option-two"}}]}},"sort":[{"id":"desc"}]}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestFilterQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Field: "title",
				Value: "Search",
				Type:  Match,
			},
			{
				Field: "content",
				Value: "Elasticsearch",
				Type:  Match,
			},
		},
		Filter: []QueryItem{
			{
				Field: "status",
				Value: "published",
				Type:  Term,
			},
			{
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

func TestNestedQuery(t *testing.T) {
	body, _ := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Field: "nested_path",
				Value: NestedQueryItem{
					Filter: []QueryItem{WrapQueryItems("filter", QueryItem{
						Field: "id",
						Value: []string{"b4ab2c6e-93e3-40b9-8e66-9379f864186f"},
						Type:  Terms,
					})},
				},
				Type: NestedQuery,
			},
		},
	})

	expected := `{"query":{"bool":{"must":[{"nested":{"path":["nested_path"],"query":{"bool":{"filter":[{"bool":{"filter":[{"terms":{"id":["b4ab2c6e-93e3-40b9-8e66-9379f864186f"]}}]}}]}}}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestHasChildQuery(t *testing.T) {
	body, err := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Value: HasChildQueryItem{
					Query: WrapQueryItems("and",
						QueryItem{
							Field: "Field1",
							Value: "some-text",
							Type:  Match,
						},
						WrapQueryItems("or",
							QueryItem{
								Field: "Field2",
								Value: "some-text-2",
								Type:  Match,
							},
							QueryItem{
								Field: "Field3",
								Value: "some-text-3",
								Type:  Match,
							},
						),
					),
					Type: "childType",
				},
				Type: HasChild,
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	expected := `{"query":{"bool":{"must":[{"has_child":{"query":{"bool":{"must":[{"match":{"Field1":"some-text"}},{"bool":{"should":[{"match":{"Field2":"some-text-2"}},{"match":{"Field3":"some-text-3"}}]}}]}},"type":"childType"}}]}}}`
	if string(body) != expected {
		t.Errorf("\nWant: %q\nHave: %q", expected, string(body))
	}
}

func TestHasChildQueryInvalid(t *testing.T) {
	_, err := json.Marshal(QueryDoc{
		Index: "some_index",
		And: []QueryItem{
			{
				Value: QueryItem{
					Field: "Field1",
					Value: "some-text",
					Type:  Match,
				},
				Type: HasChild,
			},
		},
	})

	var queryTypeErr *QueryTypeErr
	if !errors.As(err, &queryTypeErr) {
		t.Errorf("\nUnexpected error: %v", err)
	}
}
