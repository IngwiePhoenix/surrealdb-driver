package surrealdbdriver

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
)

// implements driver.Rows
type SurrealRows struct {
	conn         *SurrealConn
	rawResult    any
	resultIdx    int
	entryIdx     int
	foundColumns []string
}

var _ (driver.Rows) = (*SurrealRows)(nil)

//var _ driver.RowsColumnTypeScanType

func (rows *SurrealRows) Close() error {
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}

func (rows *SurrealRows) Columns() (cols []string) {
	// Short-circuit
	if len(rows.foundColumns) > 0 {
		return rows.foundColumns
	}

	handleSingleQueryObj := func(r st.Object) []string {
		out := []string{}
		for k := range r {
			out = append(out, k)
		}
		sort.Strings(out)
		return out
	}

	switch rows.rawResult.(type) {
	case api.QueryResponse:
		res := rows.rawResult.(api.QueryResponse)
		currId := rows.resultIdx

		if currId >= len(*res.Result) {
			// Should we panic? -- Yes, we actually probably should.
			rows.conn.Driver.LogInfo("Rows:columns, Early exit?!")
			return []string{}
		}

		currRow := (*res.Result)[currId]
		if r, ok := currRow.Result.(st.Object); ok {
			rows.conn.Driver.LogInfo("Rows:columns, Handling st.Object")
			cols = handleSingleQueryObj(r)
			rows.foundColumns = cols
			return cols
		} else if r, ok := currRow.Result.([]st.Object); ok {
			rows.conn.Driver.LogInfo("Rows:columns, Handling []st.Object")
			seen := map[string]bool{}
			for _, e := range r {
				rows.conn.Driver.LogInfo("Rows:columns, Handling st.Object in []interface{}")
				// We are dealing with a list of objects
				for eKey := range e {
					if !seen[eKey] {
						rows.conn.Driver.LogInfo("Rows:columns, Saw: ", eKey)
						seen[eKey] = true
						cols = append(cols, eKey)
					}
				}
			}
			rows.conn.Driver.LogInfo("Rows:columns, Collected: ", cols)
			sort.Strings(cols)
			rows.foundColumns = cols
			return cols
		} else if r, ok := currRow.Result.([]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:columns, Handling any in []interface{}")
			// Query that has a non-object array response.
			// Possibly something like "return [1, 2];"
			// Best to tread it basic
			for k := range r {
				cols = append(cols, strconv.Itoa(k))
			}
			return cols
		} else {
			// Assume a primitive
			return []string{"value"}
		}

	// Technically not the output of a query,
	// but this might come in handy.
	case api.InfoResponse:
		// Contains an object, so grab keys.
		// Also prevent overspin
		if rows.resultIdx > 0 {
			panic("attempting to over-index an InfoResponse in Columns")
		}
		res := rows.rawResult.(api.InfoResponse)
		obj := res.Result
		cols = handleSingleQueryObj(*obj)
		rows.foundColumns = cols
		return cols

	case api.RelationResponse:
		if rows.resultIdx > 0 {
			panic("attempting to over-index a RelationResponse in Columns")
		}
		res := rows.rawResult.(api.RelationResponse)
		cols = append(cols, []string{"id", "in", "out"}...)
		obj := *res.Result
		for k := range obj.Values {
			cols = append(cols, k)
		}
		sort.Strings(cols)
		rows.foundColumns = cols
		return cols

	// TODO: Every other response kind...

	default:
		panic(fmt.Sprintf("tried to get columns for %T, which isn't supported", rows.rawResult))
	}
	panic("reached end of Columns() unexpectedly")
}

func (rows *SurrealRows) Next(dest []driver.Value) error {
	handleResult := func(entry st.Object) error {
		rows.conn.Driver.LogInfo("Rows:next, current row: ", rows.resultIdx, entry)
		cols := rows.Columns()
		for i, v := range cols {
			rows.conn.Driver.LogInfo("Rows:next, 2nd level iteration: ", i, v)
			dv, err := convertValue(v)
			if err != nil {
				return err
			}
			dest[i] = dv
		}

		// Technically an error won't really happen here but, just in case.
		// I should probably consider using recover()...?
		return nil
	}

	switch rows.rawResult.(type) {
	case api.QueryResponse:
		res := rows.rawResult.(api.QueryResponse)
		objs := *res.Result
		rows.conn.Driver.LogInfo("Rows:next, grabbing: ", rows.resultIdx, len(objs))
		if rows.resultIdx >= len(objs) {
			rows.conn.Driver.LogInfo("Rows:next, Done reading: ", rows.resultIdx, len(objs))
			return io.EOF
		} else {
			rows.conn.Driver.LogInfo("Rows:next, STILL reading")
		}

		qres := objs[rows.resultIdx]
		obj := qres.Result
		defer func() {
			rows.resultIdx = rows.resultIdx + 1
		}()

		if qres.Status != "OK" {
			msg := obj.(string)
			return errors.New(msg)
		}

		if r, ok := obj.(st.Object); ok {
			rows.conn.Driver.LogInfo("Rows:next, Handle st.Object")
			return handleResult(r)
		} else if r, ok := obj.([]st.Object); ok {
			rows.conn.Driver.LogInfo("Rows:next, Handle []st.Object")
			// Check if we are on a good entry
			if rows.entryIdx >= len(r) {
				return io.EOF
			}
			// Increment the entry index
			defer func() {
				rows.entryIdx++
			}()
			entry := r[rows.entryIdx]
			return handleResult(entry)
		} else if r, ok := obj.([]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:next, Handle []interface{} (values)")
			// .Columns() has returned "valies", so do we.
			// Each column is just the index number, so we return the values.
			for i, v := range r {
				ev, err := convertValue(v)
				if err != nil {
					return err
				}
				dest[i] = ev
			}
			// failsafe
			return nil
		} else {
			rows.conn.Driver.LogInfo("Rows:next, Handle anything else (value)")
			// .Columns() has returned "value"
			ev, err := convertValue(obj)
			if err != nil {
				return err
			}
			dest[0] = ev
			return nil
		}

	case api.InfoResponse:
		r := rows.rawResult.(api.InfoResponse)
		re, err := convertValue(r.Result)
		if err != nil {
			return err
		}
		dest[0] = re
		return nil

	case api.RelationResponse:
		if rows.resultIdx > 0 {
			panic("attempting to over-index an InfoResponse in Columns")
		}
		res := rows.rawResult.(api.RelationResponse)
		r := *res.Result
		cols := rows.Columns()
		for i, colName := range cols {
			switch colName {
			case "id":
				dest[i] = r.ID
			case "in":
				dest[i] = r.In
			case "out":
				dest[i] = r.Out
			default:
				dest[i] = r.Values[colName]
			}
		}
		// Increment to trigger the other short-circuits
		rows.resultIdx = rows.resultIdx + 1
		return nil

	// TODO: All the other response types...

	default:
		panic(fmt.Sprintf("tried to call Next() for %T, which isn't supported", rows.rawResult))
	}

	//panic("reached end of next() unexpectedly")
}

/*
func (r *SurrealRows) ColumnTypeLength(index int) (length int64, ok bool) {

}
func (r *SurrealRows) ColumnTypeDatabaseTypeName(index int) string {
}
*/
