package surrealdbdriver

import (
	"database/sql/driver"
	"fmt"
	"io"

	"github.com/senpro-it/dsb-tool/extras/surrealdb-driver/api"
	st "github.com/senpro-it/dsb-tool/extras/surrealdb-driver/surrealtypes"
)

// implements driver.Rows
type SurrealRows struct {
	conn      *SurrealConn
	rawResult any
	resultIdx int
	batchIdx  int // api.BatchResponse
}

func (rows *SurrealRows) Close() error {
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}

func (rows *SurrealRows) Columns() (cols []string) {
	getColsForOneResult := func(res api.QueryResult) (out []string) {
		if entries, ok := res.Result.([]st.Object); ok {
			// Sanity check: Is our current index in range?
			if rows.resultIdx >= len(entries) {
				panic("Columns() was called but resultIdx is out of range of entries")
			}

			// This result set is in fact an array.
			rows.conn.Driver.LogInfo("Rows:columns, is []st.Object: ", entries)
			seen := map[string]bool{}
			entry := entries[rows.resultIdx]
			rows.conn.Driver.LogInfo("Rows:columns, []st.Object: ", entry)
			for key := range entry {
				if !seen[key] { // avoid dupes
					rows.conn.Driver.LogInfo("Rows:columns []st.Object, saw key: ", key)
					seen[key] = true
					out = append(out, key)
				}
			}

			rows.conn.Driver.LogInfo("Rows:columns, finished iteration: ", cols)
			return out
		}

		if entry, ok := res.Result.(st.Object); ok {
			// Sanity check: Is our current index in range?
			if rows.resultIdx >= 1 {
				panic("Columns() was called target is st.Object and resultIdx >= 1")
			}

			// Single object
			rows.conn.Driver.LogInfo("Rows:columns, is st.Object: ", entry)
			for key := range entry {
				out = append(out, key)
			}
			return out
		} else if list, ok := res.Result.(st.Set); ok {
			if rows.resultIdx >= 1 {
				panic("Columns() was called while target is st.Set but resultIdx >= 1")
			}
			// We return a list of strings...that simple.
			for i := range list {
				out = append(out, string(i))
			}
			return out
		}

		// Nothing else matched - so this has to be a singular result.
		out = []string{"value"}
		return out
	}

	handleQueryResp := func(resp api.QueryResponse) []string {
		// THROW queries always come back in an array.
		if resp.Result.Status != "OK" {
			// No columns here
			return []string{}
		}

		// QueryResponse is just a single response.
		return getColsForOneResult(resp.Result)
	}

	switch rows.rawResult.(type) {
	case api.QueryResponse:
		q := rows.rawResult.(api.QueryResponse)
		return handleQueryResp(q)

	case api.BatchResponse:
		resList := rows.rawResult.(api.BatchResponse)
		if rows.batchIdx >= len(resList.Result) {
			panic("Column() called on BatchResponse but rowIdx is out of range")
		}
		//q := resList[].Result
		q := resList.Result[rows.batchIdx]
		cols = handleQueryResp(q)
		rows.batchIdx = rows.batchIdx + 1
		return cols

	// Technically not the output of a query,
	// but this might come in handy.
	case api.InfoResponse:
		return []string{"info"}

	case api.RelationResponse:
		r := rows.rawResult.(api.RelationResponse)
		cols = append(cols, []string{"id", "in", "out"}...)
		for k := range r.Result.Values {
			cols = append(cols, k)
		}
		return cols

	default:
		panic(fmt.Sprintf("tried to get columns for %T, which isn't supported", rows.rawResult))
	}
}

func (rows *SurrealRows) Next(dest []driver.Value) error {
	handleResults := func(res api.QueryResult) error {

		entryToDest := func(entry st.Object) error {
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

		if entries, ok := res.Result.([]st.Object); ok {
			// This result set is in fact an array.
			rows.conn.Driver.LogInfo("Rows:next, []interface{}: ", value)
			if rows.resultIdx >= len(entries) {
				return io.EOF
			}

			err := entryToDest(entries[rows.resultIdx])
			rows.resultIdx = rows.resultIdx + 1
			return err
		}

		if entry, ok := res.Result.(st.Object); ok {
			// Single k-v object
			if rows.resultIdx >= 1 {
				return io.EOF
			}
			err := entryToDest(entry)
			rows.resultIdx = rows.resultIdx + 1
			return err
		}

		// Not an object, not an array - it's _just_ a value.
		dest[0] = res.Result
	}

	switch rows.rawResult.(type) {
	case api.QueryResponse:
		qr := rows.rawResult.(api.QueryResponse).Result
		return handleResults(qr)

	case api.BatchResponse:
		resList := rows.rawResult.(api.BatchResponse)
		if rows.batchIdx >= len(resList) {
			panic("Column() called on BatchResponse but rowIdx is out of range")
		}
		qr := resList.Result[rows.batchIdx]
		err := handleResults(qr)
		rows.batchIdx = rows.batchIdx + 1
		return err

	case api.InfoResponse:
		r := rows.rawResult.(api.InfoResponse)
		dest[0] = r.Result
		return nil

	case api.RelationResponse:
		r := rows.rawResult.(api.RelationResponse)
		// We must fake column access...yay. o.o
		rr := r.Result
		cols := rows.Columns()
		for i, colName := range cols {
			switch colName {
			case "id":
				dest[i] = rr.ID
			case "in":
				dest[i] = rr.In
			case "out":
				dest[i] = rr.Out
			default:
				dest[i] = rr.Values[colName]
			}
		}
		// TODO: There's better ways to just ignore this...
		return nil

	default:
		panic(fmt.Sprintf("tried to call Next() for %T, which isn't supported", rows.rawResult))
	}

	panic("reached end of next() unexpectedly")
}

/*
func (r *SurrealRows) ColumnTypeLength(index int) (length int64, ok bool) {

}
func (r *SurrealRows) ColumnTypeDatabaseTypeName(index int) string {
}
*/
