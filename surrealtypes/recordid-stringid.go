package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type StringID struct {
	ID    string
	Thing string
}

var _ (SurrealDBRecordID) = (*StringID)(nil)

// SurrealString implements SurrealDBRecordID.
func (id StringID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	out.WriteString(id.Thing)
	return out.String()
}

func (id *StringID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(StringID)
	*id = newId
	return nil
}
func (id *StringID) MarshalJSON() ([]byte, error) {
	s := id.SurrealString()
	return []byte(s), nil
}
func (id *StringID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *StringID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
