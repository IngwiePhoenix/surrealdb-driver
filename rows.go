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
	conn          *SurrealConn
	rawResult     any
	resultIdx     int
	entryIdx      int
	hasMultiEntry bool
	foundColumns  []string
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
		if r, ok := currRow.Result.(map[string]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:columns, Handling st.Object")
			cols = handleSingleQueryObj(r)
			rows.foundColumns = cols
			return cols
		} else if r, ok := currRow.Result.([]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:columns, Handling any in []interface{}")
			seen := map[string]bool{}
			for k, v := range r {
				if r, ok := v.(map[string]interface{}); ok {
					rows.conn.Driver.LogInfo("Rows:columns, Handling st.Object, in array")
					for _, c := range handleSingleQueryObj(r) {
						if !seen[c] {
							seen[c] = true
							cols = append(cols, c)
						}
					}
				} else {
					c := strconv.Itoa(k)
					if !seen[c] {
						seen[c] = true
						cols = append(cols, c)
					}
				}
			}
			rows.foundColumns = cols
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
	//panic("reached end of Columns() unexpectedly")
}

func (rows *SurrealRows) Next(dest []driver.Value) error {
	handleResult := func(entry st.Object) error {
		rows.conn.Driver.LogInfo("Rows:next, current row: ", rows.resultIdx, entry)
		cols := rows.Columns()
		for i, v := range cols {
			rows.conn.Driver.LogInfo("Rows:next, 2nd level iteration: ", i, v)
			dv, err := convertValue(entry[v])
			if err != nil {
				rows.conn.Driver.LogInfo("Rows:next, Saw error: ", err.Error())
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
		if rows.resultIdx >= len(objs) && !rows.hasMultiEntry {
			rows.conn.Driver.LogInfo("Rows:next, Done reading: ", rows.resultIdx, len(objs))
			return io.EOF
		} else {
			rows.conn.Driver.LogInfo("Rows:next, STILL reading")
		}

		qres := objs[rows.resultIdx]
		obj := qres.Result
		defer func() {
			if !rows.hasMultiEntry {
				rows.resultIdx = rows.resultIdx + 1
			}
		}()

		if qres.Status != "OK" {
			msg := obj.(string)
			return errors.New(msg)
		}

		if r, ok := obj.(map[string]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:next, Handle st.Object")
			return handleResult(r)
		} else if r, ok := obj.([]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:next, Handle []st.Object? ", len(r), rows.entryIdx)
			// Check if we are on a good entry
			if rows.entryIdx >= len(r) {
				rows.resultIdx = rows.resultIdx + 1
				return io.EOF
			}

			// Increment the entry index
			entry := r[rows.entryIdx]
			defer func() {
				rows.entryIdx++
			}()

			if rx, ok := entry.(map[string]interface{}); ok {
				rows.hasMultiEntry = true
				rows.conn.Driver.LogInfo("Rows:next, Handle []st.Object, indeed!")
				return handleResult(rx)
			} else if rx, ok := entry.([]interface{}); ok {
				rows.hasMultiEntry = false
				rows.conn.Driver.LogInfo("Rows:next, Handle []interface{} (values)")
				// .Columns() has returned "valies", so do we.
				// Each column is just the index number, so we return the values.
				for i, v := range rx {
					ev, err := convertValue(v)
					if err != nil {
						return err
					}
					dest[i] = ev
				}
				return nil
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

var _ (driver.RowsColumnTypeScanType) = (*SurrealRows)(nil)

func (rows *SurrealRows) ColumnTypeScanType(index int) reflect.Type {
	switch rows.rawResult.(type) {
	case api.QueryResponse:
		res := rows.rawResult.(api.QueryResponse)
		currId := rows.resultIdx

		if currId >= len(*res.Result) {
			panic("tried to read column past index")
		}

		currRow := (*res.Result)[currId]
		if r, ok := currRow.Result.(map[string]interface{}); ok {
			cols := rows.Columns()
			colVal := r[cols[index]]
			val, err := convertValue(colVal)
			if err != nil {
				panic("could not convert value: " + err.Error())
			}
			return reflect.TypeOf(val)
		} else if r, ok := currRow.Result.([]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:columns, Handling any in []interface{}")
			cols := rows.Columns()
			currRow := r[rows.resultIdx]
			if rr, ok := currRow.(map[string]interface{}); ok {
				colVal := rr[cols[index]]
				val, err := convertValue(colVal)
				if err != nil {
					panic("could not convert value: " + err.Error())
				}
				return reflect.TypeOf(val)
			} else {
				return reflect.TypeOf(currRow)
			}
		} else {
			// Assume a primitive
			return reflect.TypeOf(currRow.Result)
		}
	}
	return nil
}

*/
