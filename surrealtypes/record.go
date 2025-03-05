package surrealtypes

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/tidwall/gjson"
)

var recordTKemba = localKemba.Extend("Record[T]")

type Record[T any] struct {
	inner   T
	id      SurrealDBRecordID
	hasData bool
	hasId   bool
}

/*
var _ json.Marshaler = (*Record[T])(nil)
var _ json.Unmarshaler = (*Record[T])(nil)
var _ driver.Valuer = (*Record[T])(nil)
var _ sql.Scanner = (*Record[T])(nil)
*/

func NewRecord[T any](obj T) Record[T] {
	return Record[T]{inner: obj}
}

func (r *Record[T]) HasData() bool {
	return r.hasData
}

func (r *Record[T]) HasID() bool {
	return r.hasId
}

func (r *Record[T]) innerIsSlice() bool {
	t := reflect.TypeOf(r.inner)
	return t.Kind() == reflect.Slice
}

func (r *Record[T]) Get() (*T, bool) {
	if r.hasData {
		return &r.inner, true
	} else {
		return nil, false
	}
}

func (r *Record[T]) GetID() (SurrealDBRecordID, bool) {
	if r.hasId {
		return r.id, true
	} else {
		return nil, false
	}
}

func (r *Record[T]) UnmarshalJSON(b []byte) error {
	k := recordTKemba.Extend("UnmarshalJSON")
	data := gjson.ParseBytes(b)
	k.Printf("in: %s", data.Raw)

	if r.innerIsSlice() {
		return errors.New("surrealtypes/record: T is a slice, expected a single type (ment st.Records[T]?)")
	} else if data.IsArray() {
		return errors.New("surrealtypes/record: got array, needed single value")
	}

	// A string is an ID
	if data.Type == gjson.String {
		k.Log("Only a string was given; treating it as an ID") // TODO: "VerifyID(str)"?
		id, err := ParseID(data.String())
		if err != nil {
			k.Printf("Saw error: %s", err.Error())
			return err
		}
		k.Log(id, err)
		r.id = id
		r.hasData = false
		r.hasId = true
		return nil
	}

	// This is mainly for safety: A record should be an object.
	k.Log("It's a normal object, deserialize it")
	id, err := ParseID(data.Get("id").String())
	k.Log("done id", id.SurrealString(), err)
	if err != nil {
		return err
	}
	r.id = id
	r.hasData = true
	r.hasId = true
	err = json.Unmarshal(b, &r.inner)
	k.Log("done obj", r.inner, err)
	return err
}

func (r *Record[T]) MarshalJSON() ([]byte, error) {
	k := recordTKemba.Extend("MarshalJSON")
	switch {
	case r.hasData && r.hasId:
		k.Log("hasData && hasId", r.inner)
		return json.MarshalNoEscape(r.inner)
	case !r.hasData && r.hasId:
		k.Log("!hasData && hasId")
		return json.MarshalNoEscape(r.id.SurrealString())
	case !r.hasData && !r.hasId:
		k.Log("!hasData && !hasId")
		return []byte("null"), nil
	}
	panic("unreachable Record[T].MarshalJSON(...)")
}

func (r *Record[T]) Scan(src any) error {
	k := recordTKemba.Extend("Scan")
	switch data := src.(type) {
	case []byte:
		k.Log("Decoding bytes")
		return r.UnmarshalJSON(data)
	default:
		k.Printf("Got: %T", src)
		return fmt.Errorf("input must be []byte, found %T", src)
	}
}

func (r *Record[T]) Value() (driver.Value, error) {
	recordTKemba.Extend("Value").Log("-> MarshalJSON")
	return r.MarshalJSON()
}
