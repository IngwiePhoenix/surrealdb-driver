package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type FloatID struct {
	ID    string
	Thing float64
}

var _ (SurrealDBRecordID) = (*FloatID)(nil)

func (id FloatID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	// TODO: Information loss...?
	out.WriteString(strconv.FormatFloat(id.Thing, 'f', -1, 64))
	return out.String()
}

func (id *FloatID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(FloatID)
	*id = newId
	return nil
}
func (id *FloatID) MarshalJSON() ([]byte, error) {
	s := strconv.QuoteToGraphic(id.SurrealString())
	return []byte(s), nil
}
func (id *FloatID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *FloatID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
