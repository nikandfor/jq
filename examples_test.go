package jq_test

import (
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"nikand.dev/go/cbor"
	"nikand.dev/go/jq"
	"nikand.dev/go/jq/jqbase64"
	"nikand.dev/go/jq/jqcsv"
	"nikand.dev/go/jq/jqjson"
)

func Example() {
	// Suppose some API returns this object where among tons of other useless keys we have this:
	input := `{
		    "data": {
		        "entries": [
		            {"data": {"the actual interesting object": "here"}},
		            {"data": {"the actual interesting object": "here"}},
		            {"data": {"the actual interesting object": "here"}},
		        ],
		        "pagination": {
		            "next_cursor": "some_cursor"
		        }
		    }
		}`

	/*
		And we want:

		{
		    "results": [
		        {"the actual interesting object": "here"},
		        {"the actual interesting object": "here"},
		        {"the actual interesting object": "here"},
		    ],
		    "cursor": "some_cursor"
		}
	*/

	// Prepare filter, decoder, encoder, and buffer.
	// Better to do it once and reuse, but in a single threaded manner.

	// This is equivalent to jq program:
	// {results: [.data.entries[].data], cursor: .data.pagination.next_cursor}
	f := jq.NewObject(
		"results", jq.NewArray(
			jq.NewQuery("data", "entries", jq.NewIter(), "data"),
		),
		"cursor", jq.NewQuery("data", "pagination", "next_cursor"),
	)

	s := jq.NewSandwich(
		jqjson.NewDecoder(),
		jqjson.NewEncoder(),
	)
	s.Reset() // call if s is reused. Process* do it implicitly.

	// read data
	r := strings.NewReader(input) // net.Conn or something

	buf, err := io.ReadAll(r)
	_ = err // if err != nil ...

	var res []byte

	res, err = s.ProcessAll(f, res[:0], buf)
	// if err != nil ...

	_ = res // res now contains what we wanted
	fmt.Printf("%s\n", res)

	//// Output:
}

func Example_2() {
	// Suppose we requested data with the query and got this results in csv format.
	query := "james bond"
	input := `id,graphql_id,name,url,followers
124,ArWqv41,"James Bond","https://page-url.com/profile/124",100500
140,RefqWet,"James Bond Jr.","https://page-url.com/profile/140",1030
`

	// This is equivalent to jq program:
	// {query: "query"} + .
	f := jq.NewPlus(jq.NewObject("query", jq.NewLiteral(query)), jq.Dot{})

	d := jqcsv.NewDecoder()
	d.Tag = cbor.Map
	d.Header = true

	e := jqjson.NewEncoder()
	e.Separator = []byte{'\n'}

	s := jq.NewSandwich(d, e)
	s.Reset() // call if s is reused. Process* do it implicitly.

	buf := []byte(input)
	res, _ := s.ProcessAll(f, nil, buf)

	_ = res // res is a new line separated list of json objects with `"query": "james bond"` added to each object
	fmt.Printf("%s\n", res)

	//// Output:
}

func Example_3() {
	// generated by command
	// jq -nc '{key3: "value"} | {key2: (. | tojson)} | @base64 | {key1: .}'
	data := []byte(`{"key1":"eyJrZXkyIjoie1wia2V5M1wiOlwidmFsdWVcIn0ifQ=="}`)

	f := jq.NewQuery(
		"key1",
		jqbase64.NewDecoder(base64.StdEncoding),
		jqjson.NewDecoder(),
		"key2",
		jqjson.NewDecoder(),
		"key3",
	)

	s := jq.NewSandwich(jqjson.NewDecoder(), nil)
	s.Reset() // call if s is reused. Process* do it implicitly.

	res, _, _, err := s.DecodeApply(f, data, 0)
	_ = err // if err != nil {

	value, err := s.Buffer.Reader().BytesChecked(res)
	_ = err // if err != nil { // not a string (or bytes)

	fmt.Printf("%s\n", value)

	//// Output:
}