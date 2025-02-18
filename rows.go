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
	conn          *SurrealConn     // Root driver
	RawResult     *api.Response    // Original API response
	resultIdx     int              // Current result to be iterated over.
	entryIdx      int              // Current sub-result (rows in a query) to be iterated over.
	hasMultiEntry bool             // mark that we have N responses of which at least one has M entries
	foundColumns  []string         // Cache for columns
	k             *kemba.Kemba     // Debug logger
	e             *Debugger        // Error logger
	realRows      []gjson.Result   // Gathered results to iterate over
	isNormalized  bool             // Has this been normalized yet?
	isColsIndexed bool             // have the columns been indexed yet?
	realCols      map[int][]string // Columns per realRows
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
	r.sanityCheck()
	result := r.RawResult.Result
	if !isQueryResponse(&result) {
		panic("received invalid query!")
	}

	result.ForEach(func(i, value gjson.Result) bool {
		k.Printf("Iterating through %d entry", i.Int())
		inResult := value.Get("result")
		if inResult.IsArray() {
			inResult.ForEach(func(j, value gjson.Result) bool {
				k.Printf("adding result entry: %v", j.Value())
				entry := gjson.Parse(value.Raw)
				r.realRows = append(r.realRows, entry)
				return true
			})
		} else {
			k.Println("'twas just one result, bop it in there")
			entry := gjson.Parse(inResult.Raw)
			r.realRows = append(r.realRows, entry)
		}
		return true
	})
	r.isNormalized = true
}

func (r *SurrealRows) Columns() (cols []string) {
	r.sanityCheck()
	if !isQueryResponse(&r.RawResult.Result) {
		panic("can not get columns from non-query response: " + r.RawResult.Method)
	}
	if !r.isNormalized {
		r.Normalize()
	}
	k := r.k.Extend("Columns")
	if r.resultIdx >= len(r.realRows) {
		k.Log("Should not be here (%v >= %v)", r.resultIdx, len(r.realRows))
		return []string{}
	}
	if r.realCols == nil {
		r.realCols = make(map[int][]string)
	}

	if xcols, ok := r.realCols[r.resultIdx]; ok {
		k.Log("Found cacheed, returning")
		return xcols
	}

	var grabKeys func(gjson.Result) []string
	grabKeys = func(o gjson.Result) []string {
		k := k.Extend("grabKeys")
		k.Log("<- here")
		out := []string{}
		root := gjson.Parse(o.Raw)
		o.ForEach(func(key, value gjson.Result) bool {
			k.Printf("iterate: %s = %s", key.String(), value.String())
			//fmt.Println(key, "-> ", value.Path(original.Raw))
			p := value.Path(root.Raw)
			k.Printf("appending: %s", p)
			out = append(out, p)
			if value.Type == gjson.JSON {
				k.Log("digging deeper") // dig baby dig /s
				out = append(out, grabKeys(value)...)
			}
			return true
		})
		return out
	}

	// Not in cache, so get, set and return.
	k.Log("Uncached, processing new")
	currRow := r.realRows[r.resultIdx]
	cols = grabKeys(currRow)
	sort.Strings(cols)
	cols = funk.UniqString(cols)
	//r.realCols[r.resultIdx] = []string{}
	r.realCols[r.resultIdx] = cols
	k.Log("stored:", cols)
	return cols
}

func (r *SurrealRows) Next(dest []driver.Value) error {
	k := r.k.Extend("Next")

	if r.resultIdx >= len(r.realRows) {
		k.Log("Reached end!")
		return io.EOF
	}

	k.Log("deferring the increment")
	defer func() {
		k.Log("incrementing r.resultIdx")
		r.resultIdx++
	}()

	// - foo
	// - baz.derp
	// - quix.0.name
	currRow := r.realRows[r.resultIdx]
	cols := r.Columns()
	k.Log(cols)
	for idx, path := range cols {
		k.Printf("reading path '%s' into '%d'", path, idx)
		v := currRow.Get(path)
		if v.Exists() && v.Type != gjson.JSON {
			vv := v.Value()
			k.Printf("PUT \"%s\" dest[%d] = %v", path, idx, vv)
			dest[idx] = vv
		}
	}
	return nil
}
