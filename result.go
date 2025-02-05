package surrealdbdriver

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/senpro-it/dsb-tool/extras/surrealdb-driver/api"
	st "github.com/senpro-it/dsb-tool/extras/surrealdb-driver/surrealtypes"
)

type SurrealResult struct {
	RawResult any
}

var _ driver.Result = (*SurrealResult)(nil)

// SurrealDB's "record IDs" are strings, not ints.
// this is likely going to be a problem and a half...
// I WISH there was a way to represent a record ID nummerically.
func (r *SurrealResult) LastInsertId() (int64, error) {
	switch r.RawResult.(type) {
	case api.QueryResponse:
		qres := r.RawResult.(api.QueryResponse)
		i := 0
		if len(*qres.Result) > 1 {
			// TODO: Should be a warning.
			i = len(*qres.Result) - 1
		}
		res := (*qres.Result)[i]
		if res.Status != "OK" {
			message := res.Result.(string)
			return 0, errors.New(message)
		}

		if out, ok := res.Result.([]struct {
			Id st.RecordID `json:"id"`
		}); ok && len(out) > 0 {
			// - It's an array. (This should be the default for most queries.)
			// - Elements have an ID field
			// - There is at least 1 element
			idhash := stringToHash(out[0].Id.String())
			return int64(idhash), nil
		} else if out, ok := res.Result.(struct {
			Id st.RecordID `json:"id"`
		}); ok {
			// - It's a single object. (It was probably a custom query...)
			// - It has an id
			idhash := stringToHash(out.Id.String())
			return int64(idhash), nil
		}

	default:
		return 0, fmt.Errorf("can not grab ID from %T", r.RawResult)
	}

	// Nothing else matched - so, we got nothing.
	panic("unexpectedly reached end of LastInsertId()")
}

func (r *SurrealResult) RowsAffected() (int64, error) {
	switch r.RawResult.(type) {
	case api.QueryResponse:
		return countRows(r.RawResult.(api.QueryResponse).Result)

	case api.BatchResponse:
		res := r.RawResult.(api.BatchResponse).Result
		var total int64
		for _, qres := range res {
			t, err := countRows(qres)
			if err != nil {
				return 0, err
			}
			total = total + t
		}
		return total, nil

	default:
		return 0, fmt.Errorf("can not grab ID from %T", r.RawResult)
	}
}
