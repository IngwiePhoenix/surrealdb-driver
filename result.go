package surrealdbdriver

import "errors"

type SurrealResult struct {
	RawResult *SurrealAPIResponse
}

// SurrealDB's "record IDs" are strings, not ints.
// this is likely going to be a problem and a half...
// I WISH there was a way to represent a record ID nummerically.
func (r *SurrealResult) LastInsertId() (int64, error) {
	if r.RawResult.Error != nil {
		return 0, errors.New(r.RawResult.Error.Message)
	}
	return 0, nil
}
func (r *SurrealResult) RowsAffected() (int64, error) {
	if value, ok := r.RawResult.Result.([]interface{}); ok {
		return int64(len(value)), nil
	}
	if _, ok := r.RawResult.Result.(interface{}); ok {
		return 1, nil
	}
	if r.RawResult.Error != nil {
		return 0, errors.New(r.RawResult.Error.Message)
	}
	return 0, errors.New("RowsAffected() fell through")
}
