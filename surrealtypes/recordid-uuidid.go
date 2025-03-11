package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"

	"github.com/gofrs/uuid/v5"
)

type UUIDID struct {
	Table string
	Thing uuid.UUID
}

var _ (SurrealDBRecordID) = (*UUIDID)(nil)

func (id UUIDID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.Table)
	out.WriteByte(':')
	out.WriteRune(SRIDOpen)
	out.WriteString(id.Thing.String())
	out.WriteRune(SRIDClose)
	return out.String()
}

func (id *UUIDID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(UUIDID)
	*id = newId
	return nil
}
func (id *UUIDID) MarshalJSON() ([]byte, error) {
	s := strconv.QuoteToGraphic(id.SurrealString())
	return []byte(s), nil
}
func (id *UUIDID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *UUIDID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
