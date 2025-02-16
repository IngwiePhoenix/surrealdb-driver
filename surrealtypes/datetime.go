package surrealtypes

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/goccy/go-json"
)

type DateTime struct {
	time.Time
}

var _ json.Marshaler = (*DateTime)(nil)
var _ json.Unmarshaler = (*DateTime)(nil)
var _ driver.Valuer = (*DateTime)(nil)
var _ sql.Scanner = (*DateTime)(nil)
var _ SurrealMarshalable = (*DateTime)(nil)

func (t *DateTime) MarshalSurreal() ([]byte, error) {
	out := []byte{}
	tstr := t.Time.Format(time.RFC3339)
	out = append(out, 'd', byte('\''))
	out = append(out, []byte(tstr)...)
	out = append(out, byte('\''))
	return out, nil
}

func (t *DateTime) Scan(src any) error {
	switch data := src.(type) {
	case string:
		tmp, err := time.Parse(time.RFC3339, data)
		if err != nil {
			return err
		}
		t.Time = tmp
		return nil
	default:
		return fmt.Errorf("input must be string, found %T", src)
	}
}

func (t *DateTime) Value() (driver.Value, error) {
	return t.MarshalJSON()
}
