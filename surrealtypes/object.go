package surrealtypes

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/goccy/go-json"
)

var objectKemba = localKemba.Extend("Object")

type Object map[string]interface{}

var _ json.Marshaler = (*Object)(nil)
var _ json.Unmarshaler = (*Object)(nil)
var _ driver.Valuer = (*Object)(nil)
var _ sql.Scanner = (*Object)(nil)

func (o *Object) UnmarshalJSON(b []byte) error {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	if err != nil {
		return err
	}
	*o = m
	return nil
}

func (o *Object) MarshalJSON() ([]byte, error) {
	return json.MarshalNoEscape(o)
}

func (o *Object) Value() (driver.Value, error) {
	return o.MarshalJSON()
}

func (o *Object) Scan(value interface{}) error {
	k := objectKemba.Extend("Scan")
	k.Printf("Deserializing: %s", value)
	var err error
	if b, ok := value.([]byte); ok {
		err = o.UnmarshalJSON(b)
	} else {
		err = fmt.Errorf("surrealtypes/object: Needed []byte, got %T", value)
	}
	return err
}
