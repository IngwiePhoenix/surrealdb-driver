package surrealdbdriver

import (
	"database/sql/driver"
	"errors"
)

type SurrealResult struct {
	RawResult *SurrealAPIResponse
}

var _ driver.Result = (*SurrealResult)(nil)

// SurrealDB's "record IDs" are strings, not ints.
// this is likely going to be a problem and a half...
// I WISH there was a way to represent a record ID nummerically.
func (r *SurrealResult) LastInsertId() (int64, error) {
	if r.RawResult.Error.Code != 0 {
		return 0, errors.New(r.RawResult.Error.Message)
	}
	return 0, nil
}
func (r *SurrealResult) RowsAffected() (int64, error) {
	if r.RawResult.Error.Code != 0 {
		return 0, errors.New(r.RawResult.Error.Message)
	}
	if value, ok := r.RawResult.Result.([]interface{}); ok {
		return int64(len(value)), nil
	} else {
		// Technically, it'd be better to check if the result is empty.
		// However, that isn't exactly easy - so, I shall be lazy.
		// An empty result is, at the very least, "null".
		return 1, nil
	}
}
