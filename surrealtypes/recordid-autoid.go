package surrealtypes

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
)

type AutoIDFunc string

const (
	AutoIDRand AutoIDFunc = "rand()"
	AutoIDUUID AutoIDFunc = "uuid()"
	AutoIDULID AutoIDFunc = "ulid()"
)

// TODO: this currently does not retrive the ID.
type AutoID struct {
	ID    string
	Thing AutoIDFunc
}

var _ (SurrealDBRecordID) = (*AutoID)(nil)

func (id AutoID) SurrealString() string {
	out := strings.Builder{}
	out.WriteString(id.ID)
	out.WriteByte(':')
	out.WriteString(string(id.Thing))
	return out.String()
}

func (id *AutoID) UnmarshalJSON(b []byte) error {
	genId, err := ParseID(string(b))
	if err != nil {
		return err
	}
	newId := genId.(AutoID)
	*id = newId
	return nil
}
func (id *AutoID) MarshalJSON() ([]byte, error) {
	s := strconv.QuoteToGraphic(id.SurrealString())
	return []byte(s), nil
}
func (id *AutoID) Scan(src any) error {
	switch data := src.(type) {
	case []byte:
		return id.UnmarshalJSON(data)
	case string:
		return id.UnmarshalJSON([]byte(data))
	default:
		return fmt.Errorf("input must be []byte or string, found %T", src)
	}
}
func (id *AutoID) Value() (driver.Value, error) {
	return id.MarshalJSON()
}
