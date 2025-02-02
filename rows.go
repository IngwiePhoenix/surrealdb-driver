package surrealdbdriver

import (
	"database/sql/driver"
	"fmt"
	"io"
)

// implements driver.Rows
type SurrealRows struct {
	conn      *SurrealConn
	rawResult *SurrealAPIResponse
	resultIdx int
}

func (rows *SurrealRows) Close() error {
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}

func (rows *SurrealRows) Columns() (cols []string) {
	rows.conn.Driver.LogInfo("Rows:columns, start")
	if value, ok := rows.rawResult.Result.([]interface{}); ok {
		// This result set is in fact an array.
		rows.conn.Driver.LogInfo("Rows:columns, []interface{}: ", value)
		seen := map[string]bool{}
		for _, entry := range value {
			rows.conn.Driver.LogInfo("Rows:columns, iterating: ", entry)
			realValue := entry.(map[string]interface{})
			for k := range realValue["result"].(map[string]interface{}) {
				rows.conn.Driver.LogInfo("Rows:columns, 2nd level iteration: ", k)
				if !seen[k] { // avoid dupes
					seen[k] = true
					cols = append(cols, k)
				}
			}
		}
		rows.conn.Driver.LogInfo("Rows:columns, finished iteration: ", cols)
		return cols
	}
	/*
		if value, ok := rows.rawResult.Result.(map[string]interface{}); ok {
			// Response contains key-value pairs
			for k := range value {
				cols = append(cols, k)
			}
			return cols
		}
		if value, ok := rows.rawResult.Result.([]map[string]interface{}); ok {
			// Response contains an array of k-v pairs
			seen := map[string]bool{}
			for _, v := range value {
				for k := range v {
					if !seen[k] { // avoid dupes
						seen[k] = true
						cols = append(cols, k)
					}
				}
			}
			return cols
		}
		if _, ok := rows.rawResult.Result.(string); ok {
			// Single string response
			cols = []string{"value"}
			return cols
		}
		if _, ok := rows.rawResult.Result.([]string); ok {
			// Array-of-string response
			cols = []string{"values"}
			return cols
		}
		return cols
	*/
	panic("reached columns() unexpectedly")
}

func (rows *SurrealRows) Next(dest []driver.Value) error {
	rows.conn.Driver.LogInfo("Rows:next start: ", fmt.Sprintf("%T", rows.rawResult.Result), rows.resultIdx)
	defer rows.conn.Driver.LogInfo("Rows:finished: ", dest)
	// SurrealDB returns all results, at all time, with no paging.
	// That means we have to write the result back one by one.
	// This, however, only works if the result IS an array.
	// If it is not, then we kinda can't index it.
	// So we have to run multiple strats.
	if value, ok := rows.rawResult.Result.([]interface{}); ok {
		// This result set is in fact an array.
		rows.conn.Driver.LogInfo("Rows:next, []interface{}: ", value)
		if rows.resultIdx >= len(value) {
			return io.EOF
		}
		entry := value[rows.resultIdx]
		rows.conn.Driver.LogInfo("Rows:next, current row: ", rows.resultIdx, entry)
		realValue := entry.(map[string]interface{})
		var i int = 0
		for _, v := range realValue["result"].(map[string]interface{}) {
			rows.conn.Driver.LogInfo("Rows:next, 2nd level iteration: ", i, v)
			dest[i] = v
			i = i + 1
		}
		rows.conn.Driver.LogInfo("Rows:next, finished iteration: ", dest)
		rows.resultIdx = rows.resultIdx + 1
		return nil
	}
	/*
		if value, ok := rows.rawResult.Result.([]SurrealQueryResponse); ok {
			rows.conn.Driver.LogInfo("Rows:next, SurrealQueryResponse")
			// Can be either a single or a multi...
			if rows.resultIdx > len(value) {
				return io.EOF
			}
			resultEntry := value[rows.resultIdx].Result.([]interface{})
			var i int = 0
			for _, v := range resultEntry {
				dest[i] = v
				i = i + 1
			}
			rows.resultIdx = rows.resultIdx + 1
			return nil
		}
		if value, ok := rows.rawResult.Result.(map[string]interface{}); ok {
			rows.conn.Driver.LogInfo("Rows:next, map[string]interface{}")
			// Single k-v response
			if rows.resultIdx == 1 {
				return io.EOF
			}
			var i int = 0
			for _, v := range value {
				dest[i] = v
				i = i + 1
			}
			rows.resultIdx = rows.resultIdx + 1
			return nil
		}
		if value, ok := rows.rawResult.Result.([]map[string]any); ok {
			rows.conn.Driver.LogInfo("Rows:next, []map[string]interface{}")
			// List of k-v responses
			//if res, ok := value[rows.resultIdx]; ok {
			if idx := rows.resultIdx; idx < len(value) && value[idx] != nil {
				var i int = 0
				for _, v := range value[idx] {
					dest[i] = v
					i = i + 1
				}
				rows.resultIdx = rows.resultIdx + 1
				return nil
			} else {
				return io.EOF
			}
		}
		if value, ok := rows.rawResult.Result.([]string); ok {
			rows.conn.Driver.LogInfo("Rows:next, []string")
			// Multi-string response. Column is "values", so we just
			// put all of them in there, immediately.
			if rows.resultIdx == 1 {
				return io.EOF
			}
			dest[0] = value
			rows.resultIdx = rows.resultIdx + 1
			return nil
		}
		if value, ok := rows.rawResult.Result.(string); ok {
			rows.conn.Driver.LogInfo("Rows:next, string")
			// Single string response, column is "value"
			if rows.resultIdx == 1 {
				return io.EOF
			}
			dest[0] = value
			rows.resultIdx = rows.resultIdx + 1
			return nil
		}
	*/
	rows.conn.Driver.LogInfo("Rows:next, wtf: ", rows.rawResult.Result)
	panic("reached end of next() unexpectedly")
}
