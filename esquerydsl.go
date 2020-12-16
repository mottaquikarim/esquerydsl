// Package esquerydsl exposes various structs and a json marshal-er that makes it easier
// to safely create complex ES Search Queries via the Query DSL

package esquerydsl

import (
	"encoding/json"
	"fmt"
	"strings"
)

type QueryDoc struct {
	Index       string
	Size        int
	Sort        []map[string]string
	SearchAfter []string
	And         []QueryItem
	Not         []QueryItem
	Or          []QueryItem
	Filter      []QueryItem
	PageSize    int
}

type QueryItem struct {
	Field string
	Value interface{}
	Type  string
}

func WrapQueryItems(itemType string, items ...QueryItem) QueryItem {
	queryDoc := QueryDoc{}
	switch strings.ToLower(itemType) {
	case "or":
		queryDoc.Or = items
	case "not":
		queryDoc.Not = items
	case "filter":
		queryDoc.Filter = items
	default:
		queryDoc.And = items
	}

	return QueryItem{
		Type:  "nested",
		Value: queryDoc,
	}
}

/*
   Builds a JSON string as follows:

   {
       "query": {
           "bool": {
               "must": [ ... ]
               "should": [ ... ]
               "filter": [ ... ]
           }
       }
   }
*/
type queryReqDoc struct {
	Query       queryWrap           `json:"query,omitempty"`
	Size        int                 `json:"size,omitempty"`
	Sort        []map[string]string `json:"sort,omitempty"`
	SearchAfter []string            `json:"search_after,omitempty"`
}

type queryWrap struct {
	Bool boolWrap `json:"bool"`
}

type boolWrap struct {
	AndList    []leafQuery `json:"must,omitempty"`
	NotList    []leafQuery `json:"must_not,omitempty"`
	OrList     []leafQuery `json:"should,omitempty"`
	FilterList []leafQuery `json:"filter,omitempty"`
}

type leafQuery struct {
	Type  string
	Name  string
	Value interface{}
}

func (q leafQuery) handleMarshalType() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		(q.Type): map[string]interface{}{
			(q.Name): q.Value,
		},
	})
}

func (q leafQuery) handleMarshalQueryString() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"query_string": map[string]interface{}{
			"fields":           []string{q.Name},
			"query":            q.Value,
			"analyze_wildcard": true,
		},
	})
}

func getWrappedQuery(query QueryDoc) queryWrap {
	boolDoc := boolWrap{}
	if len(query.And) > 0 {
		boolDoc.AndList = updateList(query.And)
	}
	if len(query.Not) > 0 {
		boolDoc.NotList = updateList(query.Not)
	}
	if len(query.Or) > 0 {
		boolDoc.OrList = updateList(query.Or)
	}
	if len(query.Filter) > 0 {
		boolDoc.FilterList = updateList(query.Filter)
	}
	return queryWrap{Bool: boolDoc}
}

func (q leafQuery) MarshalJSON() ([]byte, error) {
	if q.Type == "nested" {
		return json.Marshal(getWrappedQuery(q.Value.(QueryDoc)))
	}
	// TODO: make this logic more strict
	// that is to say, currently we rely
	// on the leafQuery values are properly
	// filled for the generated query to
	// be syntactically correct
	supportedTypes := map[string]bool{
		"match":        true,
		"term":         true,
		"terms":        true,
		"wildcard":     true,
		"range":        true,
		"exists":       true,
		"query_string": true,
	}

	if _, ok := supportedTypes[q.Type]; !ok {
		return []byte(""), fmt.Errorf("query.Type %s not supported", q.Type)
	}

	// lowercase wildcard queries
	if q.Type == "wildcard" {
		if s, ok := q.Value.(string); ok {
			q.Value = strings.ToLower(s)
		}
	}

	if q.Type == "query_string" {
		return q.handleMarshalQueryString()
	}

	return q.handleMarshalType()
}

func updateList(queryItems []QueryItem) []leafQuery {
	leafQueries := make([]leafQuery, 0)
	for _, item := range queryItems {
		leafQueries = append(leafQueries, leafQuery{
			Type:  item.Type,
			Name:  item.Field,
			Value: item.Value,
		})
	}
	return leafQueries
}

func GetQueryBlock(query QueryDoc) (string, string, error) {
	queryReq := queryReqDoc{
		Query:       getWrappedQuery(query),
		Size:        query.Size,
		Sort:        query.Sort,
		SearchAfter: query.SearchAfter,
	}

	requestBody, err := json.Marshal(queryReq)
	if err != nil {
		return "", "", nil
	}

	return fmt.Sprintf(`{"index":"%s"}`, query.Index), string(requestBody), nil
}

// Elasticsearch defines a set of "reserved keywords" that MUST be escaped
// in order to be queryable. More info can be found in the docs:
// BASE: https://www.elastic.co/guide/en/elasticsearch/reference/current ...
// /query-dsl-query-string-query.html#_reserved_characters
// This solution was implemented for BAK-3966
var reserved = []string{"\\", "+", "=", "&&", "||", "!", "(", ")", "{", "}", "[", "]", "^", "\"", "~", "*", "?", ":", "/"}

func SanitizeElasticQueryField(keyword string) string {
	sanitizedKeyword := keyword
	for _, char := range reserved {
		if strings.Contains(sanitizedKeyword, char) {
			replaceWith := `\` + char
			sanitizedKeyword = strings.ReplaceAll(sanitizedKeyword, char, replaceWith)
		}
	}
	return sanitizedKeyword
}
