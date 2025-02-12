package rel

import (
	"encoding/json"
	"strings"
)

type Quote struct{}

// IDs in SurrealDB are literally a string, same for the column.
// No idea what the MySQL driver is doing differently here?...
func (q Quote) ID(name string) string {
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
	b, err := json.Marshal(v)
	if err != nil {
		panic(err.Error())
	}
	return string(b)
}
