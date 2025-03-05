package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strings"

	"github.com/oklog/ulid/v2"
)

type ULIDID struct {
	ID    string
	Thing ulid.ULID
}

var _ (SurrealDBRecordID) = (*ULIDID)(nil)

func (id ULIDID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	out.Write(id.Thing.Bytes())
	return out.String()
}

func (id *ULIDID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(ULIDID)
	*id = newId
	return nil
}
func (id *ULIDID) MarshalJSON() ([]byte, error) {
	s := id.SurrealString()
	return []byte(s), nil
}
func (id *ULIDID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *ULIDID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
