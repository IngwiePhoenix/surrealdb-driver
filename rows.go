package surrealdbdriver

import (
	"database/sql/driver"
	"errors"
	"io"
)

// implements driver.Rows
type SurrealRows struct {
	conn      *SurrealConn
	rawResult *SurrealAPIResponse
	resultIdx int
}

func (rows *SurrealRows) Columns() (cols []string) {
	if value, ok := rows.rawResult.Result.(map[string]interface{}); ok {
		// Response contains key-value pairs
		for k, _ := range value {
			cols = append(cols, k)
		}
		return cols
	}
	if value, ok := rows.rawResult.Result.([]map[string]interface{}); ok {
		// Response contains an array of k-v pairs
		seen := map[string]bool{}
		for _, v := range value {
			for k, _ := range v {
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
}
func (rows *SurrealRows) Close() error {
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}
func (rows *SurrealRows) Next(dest []driver.Value) error {
	// SurrealDB returns all results, at all time, with no paging.
	// That means we have to write the result back one by one.
	// This, however, only works if the result IS an array.
	// If it is not, then we kinda can't index it.
	// So we have to run multiple strats.
	if value, ok := rows.rawResult.Result.(map[string]interface{}); ok {
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
	if value, ok := rows.rawResult.Result.([]map[string]interface{}); ok {
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
		// Single string response, column is "value"
		if rows.resultIdx == 1 {
			return io.EOF
		}
		dest[0] = value
		rows.resultIdx = rows.resultIdx + 1
		return nil
	}
	return errors.New("Reached end of next() unexpectedly")
}
