package surrealdbdriver

import (
	"database/sql/driver"
	"errors"
	"io"
	"slices"
	"sort"
	"strconv"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	"github.com/clok/kemba"
	"github.com/tidwall/gjson"
)

// implements driver.Rows
type SurrealRows struct {
	conn          *SurrealConn
	RawResult     *api.Response
	resultIdx     int
	entryIdx      int
	hasMultiEntry bool
	foundColumns  []string
	k             *kemba.Kemba
	e             *Debugger
}

var _ (driver.Rows) = (*SurrealRows)(nil)

//var _ driver.RowsColumnTypeScanType

func (rows *SurrealRows) Close() error {
	k := rows.k.Extend("Close")
	k.Log("bye!")
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}

func (r *SurrealRows) Columns() (cols []string) {
	k := r.k.Extend("Columns")

	getKeyFromObjs := func(o gjson.Result) []string {
		out := []string{}
		o.ForEach(func(key, _ gjson.Result) bool {
			out = append(out, key.String())
			return true
		})
		return out
	}

	dedupeKeys := func(keys []string) []string {
		seen := map[string]bool{}
		out := []string{}
		for _, key := range keys {
			if !seen[key] {
				seen[key] = true
			}
		}
		for key := range seen {
			out = append(out, key)
		}
		return out
	}

	if r.RawResult.Method == api.APIMethodQuery {
		k.Log("handling query")
		v := r.RawResult.Result
		out := []string{}

		// Case: The result isn't an object or array.
		if v.Type != gjson.JSON {
			k.Log("result not an array, early skip")
			// {result: "string", status, time}
			return []string{"value"}
		}

		if v.IsArray() {
			// [{result: [{...},{...},{...}], status, time}
			k := k.Extend("isArray")

			// Valid query ("SELECT * FROM x" etc.)
			keylists := v.Get("@this.#(status==\"OK\")#.result.@join.@keys")
			k.Log("keylists:", keylists) // [[keys...], [keys...]]
			seen := map[string]bool{}
			keylists.ForEach(func(_, value gjson.Result) bool {
				value.ForEach(func(key, value gjson.Result) bool {
					sv := value.String()
					if !seen[sv] {
						seen[sv] = true
					}
					return true
				})
				return true
			})

			for key := range seen {
				out = append(out, key)
			}

			if len(out) == 1 && out[0] == "" {
				k.Log("must've been a result set of primitives?")
				out = []string{"value"}
			}

			k.Log("got:", out)
		} else if v.IsObject() {
			k := k.Extend("isObject")
			k.Log("attempting to get keys")
			for _, k := range v.Get("@keys").Array() {
				out = append(out, k.String())
			}
		} else {
			k.Log("neither object nor array, assuming primitive", v)
			out = []string{"value"}
		}
		sort.Strings(out)
		k.Log("returning", out)
		return out
	} else if slices.Contains([]api.APIMethod{
		api.APIMethodCreate,
		api.APIMethodInsert,
		api.APIMethodUpdate,
		api.APIMethodUpsert,
		api.APIMethodRelate,
		api.APIMethodMerge,
	}, r.RawResult.Method) {
		k.Log("handling CRUD")
		fullOut := []string{}

		if r.RawResult.Result.IsObject() {
			k.Log("single object")
			fullOut = getKeyFromObjs(r.RawResult.Result)
			sort.Strings(fullOut)
			return fullOut
		} else if r.RawResult.Result.IsArray() {
			k.Log("array...of objects?")
			out := []string{}
			shouldSkip := false

			r.RawResult.Result.ForEach(func(_, value gjson.Result) bool {
				if !value.IsObject() {
					k.Log("not everything is an object, skip")
					shouldSkip = true
					return false
				}
				out = append(out, getKeyFromObjs(r.RawResult.Result)...)
				return true
			})

			if !shouldSkip {
				k.Log("did not skip array, it's legit")
				out = dedupeKeys(out)
				sort.Strings(fullOut)
				return fullOut
			}
		}
	} else if r.RawResult.Result.IsArray() {
		k.Log("handling array")
		out := []string{}
		for i := range r.RawResult.Result.Array() {
			out = append(out, strconv.Itoa(i))
		}
		return out
	}

	// fallthrough
	return []string{"value"}

}

func (r *SurrealRows) Next(dest []driver.Value) error {
	k := r.k.Extend("Next")

	makeErr := func(o gjson.Result) error {
		if o.Get("status").String() != "OK" {
			k.Log("saw error", o.Get("result"))
			return errors.New(o.Get("result").String())
		}
		return nil
	}

	putValueInDest := func(i int, col string, v gjson.Result) error {
		k := k.Extend("putValueInDest")
		x, err := convertValue(v)
		if err != nil {
			k.Printf("%s: error: %s", err.Error())
			return err
		}
		k.Printf("%s: dest[%d] = %s", col, i, v.Type.String())
		dest[i] = x
		return nil
	}

	setDest := func(o gjson.Result) error {
		k := k.Extend("setDest")

		for i, col := range r.Columns() {
			k.Log("col:", col)
			v := o.Get(col)
			k.Log("col val:", o, v)
			if err := putValueInDest(i, col, v); err != nil {
				return err
			}
		}
		return nil
	}

	if r.RawResult.Method == api.APIMethodQuery {
		k.Log("handling query")

		v := r.RawResult.Result
		idx := r.resultIdx
		edx := r.entryIdx

		if v.IsArray() {
			/* r.RawResult.Result =
			[
				{
					"result": {
						"life": 42,
						"testWords": ["foo","bar","baz"]
					},
					"status": "OK",
					"time":"39.8µs"
				}
			]
			*/
			k.Log("array of results, using result", idx)
			l := len(v.Array())
			if idx >= l {
				k.Log("end")
				return io.EOF
			}
			defer func() {
				k.Log("trigger resultIdx++ (query)")
				r.resultIdx++
			}()
			v := v.Array()[idx] // -> object{ result: object|[]object, status, time }

			if err := makeErr(v); err != nil {
				return err
			}

			if v.Get("result").IsArray() {
				k.Log("array of results, with an array of results, using entry", edx)
				if err := makeErr(v); err != nil {
					return err
				}
				// Drop down into result array
				v := v.Get("result")
				l := len(v.Array())
				if edx >= l {
					// from previous iteration; wrap around
					edx = 0
					r.entryIdx = 0
				}
				// Drop down to current index
				v = v.Array()[edx]
				return setDest(v)
			} else if v.Get("result").IsObject() {
				k.Log("array of results with a single object")
				if err := makeErr(v); err != nil {
					return err
				}
				return setDest(v.Get("result"))
			} else {
				k.Log("result[].result != object|[]object")
				return putValueInDest(0, "value", v.Get("result"))
			}
		} else if v.IsObject() {
			return setDest(v)
		}
	} else if slices.Contains([]api.APIMethod{
		api.APIMethodCreate,
		api.APIMethodInsert,
		api.APIMethodUpdate,
		api.APIMethodUpsert,
		api.APIMethodRelate,
		api.APIMethodMerge,
	}, r.RawResult.Method) {
		k.Log("handling CRUD")

		v := r.RawResult.Result
		if v.IsObject() {
			if r.resultIdx > 0 {
				return io.EOF
			}
			return setDest(v)
		} else if v.IsArray() {
			l := len(v.Array())
			if r.resultIdx >= l {
				return io.EOF
			}
			defer func() {
				k.Log("trigger resultIdx++ (CRUD)")
				r.resultIdx++
			}()
			return setDest(v.Array()[r.resultIdx])
		}
	} else if r.RawResult.Result.IsArray() {
		k.Log("handling plain array")

		v := r.RawResult.Result
		l := len(v.Array())
		if r.resultIdx >= l {
			return io.EOF
		}
		defer func() {
			k.Log("trigger resultIdx++ (array...?)")
			r.resultIdx++
		}()
		return setDest(v.Array()[r.resultIdx])
	}

	if r.resultIdx > 0 {
		return io.EOF
	}
	v := r.RawResult.Result
	x, err := convertValue(v)
	if err != nil {
		return err
	}
	dest[0] = x
	return nil
}
