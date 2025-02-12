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
	conn      *SurrealConn
	rawResult any
	resultIdx int
}

func (rows *SurrealRows) Close() error {
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}

func (rows *SurrealRows) Columns() (cols []string) {
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
			// Should we panic?
			return []string{}
		}
		currRow := (*res.Result)[currId]
		if r, ok := currRow.Result.(st.Object); ok {
			return handleSingleQueryObj(r)
		} else if r, ok := currRow.Result.([]interface{}); ok {
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
		return handleSingleQueryObj(*obj)

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
		return cols

	// TODO: Every other response kind...

	default:
		panic(fmt.Sprintf("tried to get columns for %T, which isn't supported", rows.rawResult))
	}
}

func (rows *SurrealRows) Next(dest []driver.Value) error {
	handleResult := func(entry st.Object) error {
		rows.conn.Driver.LogInfo("Rows:next, current row: ", rows.resultIdx, entry)
		cols := rows.Columns()
		for i, v := range cols {
			rows.conn.Driver.LogInfo("Rows:next, 2nd level iteration: ", i, v)
			dest[i] = entry[v]
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
			return handleResult(r)
		} else if r, ok := obj.([]interface{}); ok {
			// .Columns() has returned "valies", so do we.
			// Each column is just the index number, so we return the values.
			for i, v := range r {
				// TODO: Can we add more type info...?
				dest[i] = v
			}
			return nil
		} else {
			// .Columns() has returned "value"
			dest[0] = obj
			return nil
		}

	case api.InfoResponse:
		r := rows.rawResult.(api.InfoResponse)
		dest[0] = r.Result
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
