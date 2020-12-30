![build](https://github.com/mottaquikarim/esquerydsl/workflows/Build%20Status/badge.svg) [![GoDoc](https://godoc.org/github.com/mottaquikarim/esquerydsl?status.svg)](https://godoc.org/github.com/mottaquikarim/esquerydsl)
# [ES Query DSL](https://godoc.org/github.com/mottaquikarim/esquerydsl)
Structs and marshal-ers that help wrangle writing elastic search queries using the [ES query DSL](https://www.elastic.co/guide/en/elasticsearch/reference/current/query-dsl.html) spec

## Installation

```bash
go get github.com/mottaquikarim/esquerydsl
```

## Usage

(You can copy paste the snippet below into [playground](https://play.golang.org/) and see for yourself)

```go
package main

import (
	"fmt"

	"github.com/mottaquikarim/esquerydsl"
)

func main() {
	body, _ := json.Marshal(esquerydsl.QueryDoc{
		Index: "some_index",
		Sort:  []map[string]string{map[string]string{"id": "asc"}},
		And: []esquerydsl.QueryItem{
			esquerydsl.QueryItem{
				Field: "some_index_id",
				Value: "some-long-key-id-value",
				Type:  esquerydsl.Match,
			},
		},
	})
	fmt.Println(string(body))
	// {"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}},"sort":[{"id":"asc"}]}
}

```

(Please find additional examples in the unit tests)

## MultiSearch Support

```go
package main

import (
	"fmt"

	"github.com/mottaquikarim/esquerydsl"
)

func main() {
	doc, _ := MultiSearchDoc([]QueryDoc{
		QueryDoc{
			Index: "index1",
			And: []QueryItem{
				QueryItem{
					Field: "user.id",
					Value: "kimchy!",
					Type:  QueryString,
				},
			},
		},
		QueryDoc{
			Index: "index2",
			And: []QueryItem{
				QueryItem{
					Field: "some_index_id",
					Value: "some-long-key-id-value",
					Type:  Match,
				},
			},
		},
	})
	fmt.Println(doc)
	// OUTPUT:

	// {"index":"index1"}
	// {"query":{"bool":{"must":[{"query_string":{"analyze_wildcard":true,"fields":["user.id"],"query":"kimchy\\!"}}]}}}
	// {"index":"index2"}
	// {"query":{"bool":{"must":[{"match":{"some_index_id":"some-long-key-id-value"}}]}}}

	// ...
	// eventually, pass this in like so:
	
	// request := esapi.MsearchRequest{
	//	Body: strings.NewReader(doc),
	// }

}

```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

### Run Tests

```bash
make test
```

### Format

```bash
make fmt
```

### Lint

```bash
make lint
```

## License
[MIT](https://choosealicense.com/licenses/mit/)
