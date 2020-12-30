// Package esquerydsl exposes various structs and a json marshal-er that makes it easier
// to safely create complex ES Search Queries via the Query DSL

package esquerydsl

import (
	"encoding/json"
	"fmt"
	"strings"
)

type QueryType int

const (
	Match QueryType = iota
	Term
	Terms
	Wildcard
	Range
	Exists
	QueryString
	Nested
)

type QueryTypeErr struct {
	typeVal QueryType
}

func (e *QueryTypeErr) Error() string {
	return fmt.Sprintf("Type %d is not supported", e.typeVal)
}

func (qt QueryType) String() (string, error) {
	convs := [...]string{
		"match",
		"term",
		"terms",
		"wildcard",
		"range",
		"exists",
		"query_string",
		"nested",
	}
	if int(qt) > len(convs) {
		return "", &QueryTypeErr{typeVal: qt}
	}

	return convs[qt], nil
}

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
	Type  QueryType
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
		Type:  Nested,
		Value: queryDoc,
	}
}

// Builds a JSON string as follows:
// {
//     "query": {
//         "bool": {
//             "must": [ ... ]
//             "should": [ ... ]
//             "filter": [ ... ]
//         }
//     }
// }
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
	Type  QueryType
	Name  string
	Value interface{}
}

func (q leafQuery) handleMarshalType(queryType string) ([]byte, error) {
	// lowercase wildcard queries
	if q.Type == Wildcard {
		if s, ok := q.Value.(string); ok {
			q.Value = strings.ToLower(s)
		}
	}

	if q.Type == QueryString {
		return q.handleMarshalQueryString(queryType)
	}

	return json.Marshal(map[string]interface{}{
		(queryType): map[string]interface{}{
			(q.Name): q.Value,
		},
	})
}

func (q leafQuery) handleMarshalQueryString(queryType string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		queryType: map[string]interface{}{
			"fields":           []string{q.Name},
			"query":            SanitizeElasticQueryField(q.Value.(string)),
			"analyze_wildcard": true, // TODO: make this configurable
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
	if q.Type == Nested {
		return json.Marshal(getWrappedQuery(q.Value.(QueryDoc)))
	}

	if queryType, err := q.Type.String(); err != nil {
		return []byte(""), err
	} else {
		return q.handleMarshalType(queryType)
	}
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

func (query QueryDoc) MarshalJSON() ([]byte, error) {
	queryReq := queryReqDoc{
		Query:       getWrappedQuery(query),
		Size:        query.Size,
		Sort:        query.Sort,
		SearchAfter: query.SearchAfter,
	}

	requestBody, err := json.Marshal(queryReq)
	if err != nil {
		return nil, err
	}

	return requestBody, nil
}

// Elasticsearch defines a set of "reserved keywords" that MUST be escaped
// in order to be queryable. More info can be found in the docs:
// BASE: https://www.elastic.co/guide/en/elasticsearch/reference/current ...
// /query-dsl-query-string-query.html#_reserved_characters
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
