package surrealdbdriver

import (
	"database/sql/driver"
	"fmt"
	"io"
	"sort"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	"github.com/clok/kemba"
	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

// implements driver.Rows
type SurrealRows struct {
	conn         *SurrealConn   // Root driver
	RawResult    *api.Response  // Original API response
	isNormalized bool           // Has this been normalized yet?
	realRows     []gjson.Result // Gathered results to iterate over
	realCols     []string       // Columns per realRows
	resultIdx    int            // Current result to be iterated over.
	k            *kemba.Kemba   // Debug logger
	e            *Debugger      // Error logger
}

var _ (driver.Rows) = (*SurrealRows)(nil)

//var _ driver.RowsColumnTypeScanType

func (rows *SurrealRows) sanityCheck() {
	if rows.RawResult.Method != api.APIMethodQuery {
		msg := "SurrealRows does not support anything but query responses." +
			" This one is of type: %s"
		panic(fmt.Sprint(msg, rows.RawResult.Method))
	}
}

func (rows *SurrealRows) Close() error {
	k := rows.k.Extend("Close")
	k.Log("bye!")
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}

func (r *SurrealRows) Normalize() {
	k := r.k.Extend("Normalize")
	if r.isNormalized {
		k.Log("early exit")
		return
	}

	// Check query
	r.sanityCheck()
	result := r.RawResult.Result
	if !isQueryResponse(&result) {
		panic("received invalid query!")
	}

	// Populate rows and columns
	cols := []string{}
	result.ForEach(func(i, value gjson.Result) bool {
		k.Printf("Iterating through %d entry", i.Int())
		inResult := value.Get("result")
		if inResult.IsArray() {
			inResult.ForEach(func(j, value gjson.Result) bool {
				k.Printf("adding result entry: %v", j.Value())
				entry := gjson.Parse(value.Raw)
				r.realRows = append(r.realRows, entry)
				k.Printf("grabbing columns: %v", j.Value())
				entryCols := r.grabKeys(entry, entry)
				cols = append(cols, entryCols...)
				return true
			})
		} else {
			k.Println("'twas just one result, bop it in there")
			entry := gjson.Parse(inResult.Raw)
			r.realRows = append(r.realRows, entry)
		}
		return true
	})

	k.Log("sorting cols")
	cols = funk.UniqString(cols)
	sort.Strings(cols)
	r.realCols = cols
	r.isNormalized = true
}

func (r *SurrealRows) grabKeys(root gjson.Result, o gjson.Result) []string {
	k := r.k.Extend("grabKeys")
	k.Log("<- here")
	out := []string{}
	o.ForEach(func(key, value gjson.Result) bool {
		k.Printf("iterate: %s = %s", key.String(), value.String())
		p := value.Path(root.Raw)
		//if !value.IsObject() { // !value.IsAttay() {
		k.Printf("appending: %s", p)
		out = append(out, p)
		//}
		/*if value.IsObject() { // value.IsArray() {
			k.Log("digging deeper") // dig baby dig /s
			out = append(out, r.grabKeys(root, value)...)
		}*/
		return true
	})
	return out
}

func (r *SurrealRows) Columns() (cols []string) {
	r.sanityCheck()
	r.Normalize()
	if !isQueryResponse(&r.RawResult.Result) {
		panic("can not get columns from non-query response: " + r.RawResult.Method)
	}
	k := r.k.Extend("Columns")
	k.Log("returning columns")
	return r.realCols
}

func (r *SurrealRows) Next(dest []driver.Value) error {
	k := r.k.Extend("Next")

	if r.resultIdx >= len(r.realRows) {
		k.Log("Reached end!")
		return io.EOF
	}

	// - foo
	// - baz.derp
	// - quix.0.name
	currRow := r.realRows[r.resultIdx]
	cols := r.Columns()
	k.Log(cols)
	for idx, path := range cols {
		v := currRow.Get(path)
		k.Printf("reading path '%s' into '%d' as %s", path, idx, v.Type)
		if v.Exists() {
			if v.IsArray() || v.IsObject() {
				k.Printf("PUT \"%s\" dest[%d] = []byte(%s)", path, idx, v.String())
				dest[idx] = []byte(v.Raw)
			} else {
				vv := v.Value()
				k.Printf("PUT \"%s\" dest[%d] = %v", path, idx, vv)
				vv, err := convertValue(v)
				if err != nil {
					return err
				}
				dest[idx] = vv
			}
		}
	}

	k.Log("incrementing r.resultIdx")
	r.resultIdx++
	return nil
}
