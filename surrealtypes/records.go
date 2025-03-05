package surrealtypes

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"
)

var recordsTkemba = localKemba.Extend("Records[T]")

type Records[T any] struct {
	inner       []*Record[T]
	hasAnything bool
}

/*
var _ json.Marshaler = (*Records[T])(nil)
var _ json.Unmarshaler = (*Records[T])(nil)
var _ driver.Valuer = (*Records[T])(nil)
var _ sql.Scanner = (*Records[T])(nil)
*/

func NewRecords[T any](obj []T) Records[T] {
	k := recordsTkemba.Extend("NewRecords[T]")
	k.Printf("Making Record[T]s off of %T", obj)
	out := make([]*Record[T], len(obj))
	for _, o := range obj {
		r := NewRecord(o)
		out = append(out, &r)
	}
	return Records[T]{inner: out}
}

func (r *Records[T]) Get() ([]*Record[T], bool) {
	return r.inner, r.hasAnything
}

func (r *Records[T]) Len() int {
	return len(r.inner)
}

func (r *Records[T]) UnmarshalJSON(b []byte) error {
	k := recordsTkemba.Extend("UnmarshalJSON")
	data := gjson.ParseBytes(b)

	if !data.IsArray() {
		return errors.New("surrealtypes/records: got a non-array, needed an array")
	}

	var err error = nil

	if len(data.Array()) > 0 {
		k.Log("array has more than one object")
		r.hasAnything = true
		data.ForEach(func(key, value gjson.Result) bool {
			k := k.Extend("ForEach")
			k.Log("processing", key, value)
			one := new(Record[T])
			err = json.Unmarshal([]byte(value.Raw), one)
			if err != nil {
				k.Log("Error caused with this: " + value.Raw)
				return false
			}
			r.inner = append(r.inner, one)
			return true
		})
	} else {
		k.Log("array has no values")
		r.hasAnything = false
	}
	k.Log(err)
	return err
}

func (r *Records[T]) MarshalJSON() ([]byte, error) {
	k := recordsTkemba.Extend("MarshalJSON")
	if r.hasAnything {
		// BUG(IP): %T may result in pretty.formatter from github.com/kr/pretty ... fml.
		k.Log(fmt.Sprintf("Marshalling Records[T] -> %T", r.inner), r.inner)
		return json.Marshal(r.inner)
	} else {
		k.Log("Nothing to unmarshal, returning empty array")
		return []byte("[]"), nil
	}
}

func (r *Records[T]) Scan(src any) error {
	k := recordsTkemba.Extend("Scan")
	switch data := src.(type) {
	case []byte:
		k.Log("Found []byte")
		return r.UnmarshalJSON(data)
	default:
		return fmt.Errorf("input must be []byte, found %T", src)
	}
}

func (r *Records[T]) Value() (driver.Value, error) {
	k := recordsTkemba.Extend("Value")
	k.Log("-> MarshalJSON")
	return r.MarshalJSON()
}
