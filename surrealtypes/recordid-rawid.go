package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type RawID struct {
	ID    string
	Thing []rune
}

var _ (SurrealDBRecordID) = (*RawID)(nil)

// SurrealString implements SurrealDBRecordID.
func (id RawID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	out.WriteRune(SRIDOpen)
	for _, r := range id.Thing {
		out.WriteRune(r)
	}
	out.WriteRune(SRIDClose)
	return out.String()
}

func (id *RawID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(RawID)
	*id = newId
	return nil
}
func (id *RawID) MarshalJSON() ([]byte, error) {
	s := id.SurrealString()
	return []byte(s), nil
}
func (id *RawID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *RawID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
