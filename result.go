package surrealdbdriver

import (
	"database/sql/driver"
	"errors"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
	"github.com/clok/kemba"
	"github.com/tidwall/gjson"
)

type SurrealResult struct {
	RawResult *api.Response
	k         *kemba.Kemba
	e         *Debugger
}

var _ driver.Result = (*SurrealResult)(nil)

func (r *SurrealResult) LastInsertId() (int64, error) {
	k := r.k.Extend("LastInsertID")
	if r.RawResult.Method != api.APIMethodQuery {
		return 0, errors.New("can only handle query results")
	}

	var v gjson.Result
	if r.RawResult.Result.IsArray() {
		k.Log("result is an array")
		a := r.RawResult.Result.Array()
		l := len(a)
		v = a[l-1]
	} else {
		k.Log("result is an object")
		v = r.RawResult.Result
	}

	var res gjson.Result
	var done bool = false
	if v.IsArray() {
		k.Log("target is an array")
		a := v.Array()
		l := len(a)
		o := a[l-1]
		if o.Get("id").Exists() {
			k.Log("...and last element has an ID - valid.")
			res = o
			done = true
		}
	} else if v.IsObject() {
		k.Log("target is an object")
		if v.Get("result").Get("id").Exists() {
			k.Log("...and has an ID, valid.")
			res = v.Get("result")
			done = true
		}
	}

	if !done {
		k.Log("Nothing found: " + r.RawResult.Method)
		// Nothing valid was found. Yeet.
		return 0, errors.New("could not determine ID; no record was returned")
	}

	srid, err := st.ParseID(res.Get("id").String())
	if err != nil {
		return 0, err
	}
	idhash := int64(stringToHash(srid.SurrealString()))
	return idhash, nil
}

func (r *SurrealResult) RowsAffected() (int64, error) {
	v := r.RawResult.Result
	if r.RawResult.Method != api.APIMethodQuery {
		return -1, errors.New("can only handle query results")
	}

	if v.IsArray() {
		var count int
		v.ForEach(func(_, value gjson.Result) bool {
			if value.IsObject() {
				count++
			}
			return true
		})
		return int64(count), nil
	} else {
		// Even {result: [NONE]} (i.e. from signin) is one affected row.
		return 1, nil
	}
}
