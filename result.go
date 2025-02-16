package surrealdbdriver

import (
	"database/sql/driver"
	"slices"

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

// SurrealDB's "record IDs" are strings, not ints.
// this is likely going to be a problem and a half...
// I WISH there was a way to represent a record ID nummerically.
// ...so I hotchpotch'd one. Yay. -.-
// Also! How the heck do I handle this? Just literally the last one...?
func (r *SurrealResult) LastInsertId() (int64, error) {
	k := r.k.Extend("LastInsertID")
	var rid string = ""
	if r.RawResult.Method == api.APIMethodQuery {
		k.Log("handling query")
		v := r.RawResult.Result.Array()
		l := len(v)
		k.Log("length", l)
		k.Log("v:", r.RawResult.Result)
		tid := v[l-1].Get("id")
		if tid.Exists() {
			rid = tid.String()
		} else {
			// INFO FOR...
			rid = "-2"
		}
		k.Log("rid", rid)
	} else if slices.Contains([]api.APIMethod{
		api.APIMethodCreate,
		api.APIMethodInsert,
		api.APIMethodUpdate,
		api.APIMethodUpsert,
		api.APIMethodRelate,
		api.APIMethodMerge,
	}, r.RawResult.Method) {
		k.Log("handling CRUDs")
		v := r.RawResult.Result
		if v.IsArray() {
			k.Log("CRUD is an array")
			va := v.Array()
			l := len(va)
			rid = va[l-1].Get("id").String()
		} else {
			k.Log("CRUD is not an array")
			if va := v.Get("id"); va.Exists() {
				rid = va.String()
			}
		}
	} else {
		panic("the API method <" + string(r.RawResult.Method) + "> isn't implemented yet")
	}

	srid, err := st.ParseID(rid)
	if err != nil {
		return -1, err
	}
	idhash := int64(stringToHash(srid.SurrealString()))
	return idhash, nil
}

func (r *SurrealResult) RowsAffected() (int64, error) {
	v := r.RawResult.Result
	if r.RawResult.Method == api.APIMethodQuery {
		return int64(len(v.Array())), nil
	} else if v.IsArray() {
		var objsFound int64 = 0
		v.ForEach(func(_, value gjson.Result) bool {
			if value.IsObject() && value.Get("id").Exists() {
				objsFound++
				return true
			}
			return false
		})
		return objsFound, nil
	} else if v.Type == gjson.Null {
		return 0, nil
	}

	// Everything else is one value, and one value only.
	return 1, nil
}
