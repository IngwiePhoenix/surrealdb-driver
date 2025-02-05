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
// ...so I hotchpotch'd one. Yay. -.-
// Also! How the heck do I handle this? Just literally the last one...?
func (r *SurrealResult) LastInsertId() (int64, error) {
	switch r.RawResult.(type) {
	case api.QueryResponse:
		qres := (*r.RawResult.(api.QueryResponse).Result)
		i := 0
		if len(qres) > 1 {
			// TODO: Should be a warning.
			i = len(qres) - 1
		}
		res := qres[i]

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
			// Pick the literally last entry
			idhash := stringToHash(out[len(out)-1].Id.String())
			return int64(idhash), nil
		} else if out, ok := res.Result.(struct {
			Id st.RecordID `json:"id"`
		}); ok {
			// - It's a single object. (It was probably a custom query...)
			// - It has an id
			idhash := stringToHash(out.Id.String())
			return int64(idhash), nil
		}

	case api.SingleNoSQLResponse:
		res := r.RawResult.(api.SingleNoSQLResponse)
		if idValue, ok := (*res.Result)["id"]; ok {
			id, err := idValue.(string)
			if !err {
				return 0, errors.New("could not get ID from NoSQL object")
			}
			idhash := stringToHash(id)
			return int64(idhash), nil
		}
		return 0, errors.New("tried to index a NoSQL result's id, but there was none")

	case api.MultiNoSQLResponse:
		res := (*r.RawResult.(api.MultiNoSQLResponse).Result)
		if len(res) <= 0 {
			return 0, errors.New("tried to index a multi-NoSQL result, that had no results")
		}
		innerRes := res[len(res)-1]
		if idValue, ok := innerRes["id"]; ok {
			id, err := idValue.(string)
			if !err {
				return 0, errors.New("could not get ID from NoSQL object")
			}
			idhash := stringToHash(id)
			return int64(idhash), nil
		}
		return 0, errors.New("tried to index a Multi-NoSQL result's id, but there was none")

	default:
		return 0, fmt.Errorf("can not grab ID from %T", r.RawResult)
	}

	// Nothing else matched - so, we got nothing.
	panic("unexpectedly reached end of LastInsertId()")
}

func (r *SurrealResult) RowsAffected() (int64, error) {
	switch r.RawResult.(type) {
	case api.QueryResponse:
		res := r.RawResult.(api.QueryResponse).Result
		return int64(len(*res)), nil

	case api.SingleNoSQLResponse:
		res := r.RawResult.(api.SingleNoSQLResponse)
		if res.Result != nil {
			return 1, nil
		}
		return -1, errors.New("Single NoSQL response had no object")

	case api.MultiNoSQLResponse:
		res := r.RawResult.(api.MultiNoSQLResponse)
		if res.Result != nil {
			return int64(len(*res.Result)), nil
		}
		return -1, errors.New("Multi NoSQL response had no objects")

	default:
		return 0, fmt.Errorf("can not grab ID from %T", r.RawResult)
	}
}
