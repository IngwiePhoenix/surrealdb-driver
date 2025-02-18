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
	inner []Record[T]
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
	out := make([]Record[T], len(obj))
	for _, o := range obj {
		out = append(out, NewRecord(o))
	}
	return Records[T]{inner: out}
}

func (r *Records[T]) Get() []Record[T] {
	return r.inner
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
	data.ForEach(func(key, value gjson.Result) bool {
		k.Log(key, value)
		one := Record[T]{}
		err = json.Unmarshal([]byte(value.Raw), &one)
		if err != nil {
			return false
		}
		r.inner = append(r.inner, one)
		return true
	})
	k.Log(err)
	return err
}

func (r *Records[T]) MarshalJSON() ([]byte, error) {
	return json.MarshalNoEscape(r.inner)
}

func (r *Records[T]) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return r.UnmarshalJSON(data)
	default:
		return fmt.Errorf("input must be []byte, found %T", src)
	}
}

func (r *Records[T]) Value() (driver.Value, error) {
	return r.MarshalJSON()
}
