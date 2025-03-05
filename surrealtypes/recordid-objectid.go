package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

type ObjectID struct {
	ID    string
	Thing gjson.Result
}

var _ (SurrealDBRecordID) = (*ObjectID)(nil)

// SurrealString implements SurrealDBRecordID.
func (id ObjectID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	out.WriteString(id.Thing.Raw)
	return out.String()
}

func (id *ObjectID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(ObjectID)
	*id = newId
	return nil
}
func (id *ObjectID) MarshalJSON() ([]byte, error) {
	s := id.SurrealString()
	return []byte(s), nil
}
func (id *ObjectID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *ObjectID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
