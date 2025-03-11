package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type IntID struct {
	ID    string
	Thing int64
}

var _ (SurrealDBRecordID) = (*IntID)(nil)

func (id IntID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	// TODO: Information loss...?
	out.WriteString(strconv.Itoa(int(id.Thing)))
	return out.String()
}

func (id *IntID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(IntID)
	*id = newId
	return nil
}
func (id *IntID) MarshalJSON() ([]byte, error) {
	s := strconv.QuoteToGraphic(id.SurrealString())
	return []byte(s), nil
}
func (id *IntID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *IntID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
