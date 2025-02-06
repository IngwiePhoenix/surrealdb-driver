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
		if r, ok := res.Result.(st.Object); ok {
			// Do we even _have_ an ID?
			idStr, ok := r["id"].(string)
			if !ok {
				// No id, bail.
				return -1, nil
			}
			rid, err := st.NewRecordIDFromString(idStr)
			if err != nil {
				return 0, err
			}
			idhash := int64(stringToHash(rid.String()))
			return idhash, nil
		} else {
			// Ok, so this is likely a random query.
			// What the fuck am I supposed to do here?
			return -1, nil
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
	panic("unexpectedly reached end of LastInsertId() with " + fmt.Sprintf("%T", r.RawResult))
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
