package surrealdbdriver

import (
	"database/sql/driver"
	"hash/fnv"
	"os"
	"strconv"
	"time"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
	"github.com/clok/kemba"
	"github.com/tidwall/gjson"
	"github.com/ztrue/tracerr"
)

type Debugger struct {
	k *kemba.Kemba
}

func (d *Debugger) Debug(e error) bool {
	var skip = true
	if _, ok := os.LookupEnv("KEMBA"); ok {
		skip = false
	} else if _, ok := os.LookupEnv("DEBUG"); ok {
		skip = false
	}

	if skip {
		return e != nil
	}

	if e != nil {
		d.k.Log("ERROR : " + tracerr.SprintSource(e))
		return true
	}
	return false
}

func makeErrorLogger(k *kemba.Kemba) *Debugger {
	return &Debugger{k: k}
}

// Basically a ripoff from: https://github.com/go-sql-driver/mysql/blob/341a5a5246835b2ac4b8d36bb12a9dfad70663f4/statement.go#L143
// Only the variable names were slightly changed but...that's that.
// Purpose of this method is to convert the value to something sensible, and error out
// when the value would technically not be compatible anymore.
// Further, the following is not respected yet:
// > If CheckNamedValue returns ErrRemoveArgument, the NamedValue will not be included
// > in the final query arguments. This may be used to pass special options to the query itself.
// >
// > If ErrSkip is returned the column converter error checking path is used for the argument.
// > Drivers may wish to return ErrSkip after they have exhausted their own special cases.
// (via: https://pkg.go.dev/database/sql/driver#NamedValueChecker)
func checkNamedValue(value any) (driver.Value, error) {
	if s, ok := value.(st.SurrealDBRecordID); ok {
		return s.SurrealString(), nil
	} else if t, ok := value.(time.Time); ok {
		return `d'` + t.Format(time.RFC3339) + `'`, nil
	}
	return value, nil
}

func validateResponse(method api.APIMethod, data []byte) (*api.Response, error) {
	// Valid JSON?
	if !gjson.ValidBytes(data) {
		return nil, strconv.ErrSyntax
	}

	// SurrealDB error?
	r := gjson.ParseBytes(data)
	if err := r.Get("error"); err.Exists() {
		return nil, &api.APIError{
			Code:    int(err.Get("code").Int()),
			Message: err.Get("message").String(),
		}
	}

	result := r.Get("result")
	return &api.Response{
		Method: method,
		Result: result,
	}, nil
}

// HACK: Need to decode it too...
func stringToHash(str string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}

func isQueryResponse(o *gjson.Result) bool {
	isValid := false
	// 1. Is the result an array?
	if o.IsArray() {
		// 2. Is each element in the result a valid query response?
		//    as in, does it have .result .time and .status ?
		o.ForEach(func(_, v gjson.Result) bool {
			if v.IsObject() {
				for _, k := range v.Get("@keys").Array() {
					k := k.String()
					isValid = k == "id" || k == "result" || k == "time"
				}
			}
			return isValid
		})
	}
	return isValid
}
