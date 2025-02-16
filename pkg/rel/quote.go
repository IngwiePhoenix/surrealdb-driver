package rel

import (
	"github.com/goccy/go-json"
	"fmt"
	"strings"

	"github.com/go-rel/sql/builder"
)

type Quote struct{}

var _ (builder.Quoter) = (*Quote)(nil)

// IDs in SurrealDB are literally a string, same for the column.
// No idea what the MySQL driver is doing differently here?...
func (q Quote) ID(name string) string {
	fmt.Println("!! QUOTER: ID", name)
	bytes, err := json.Marshal(name)
	if err != nil {
		panic(err.Error())
	}
	return strings.Trim(string(bytes), "\"")
}

// TODO: I might have to do something specific here... maybe. perhaps.
// I wonder if I can just marshall into just an escaped string?
// Wouldn't surprise me if there was a JSON quoter... but on the other hand,
// I don't really need anything else. o.o
func (q Quote) Value(v interface{}) string {
	fmt.Println("!! QUOTER: Value", v)
	return fmt.Sprintf("%v", v)
	/*if s, ok := v.(interface{ String() string }); ok {
		return s.String()
	}
	b, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return string(b)*/
}
