package surrealtypes

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/goccy/go-json"
)

type Duration struct {
	time.Duration
}

var _ json.Marshaler = (*Duration)(nil)
var _ json.Unmarshaler = (*Duration)(nil)
var _ driver.Valuer = (*Duration)(nil)
var _ sql.Scanner = (*Duration)(nil)

func (d *Duration) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str == "null" {
		d.Duration = 0
		return nil
	}

	str = str[1 : len(str)-1]

	duration, err := time.ParseDuration(str)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

func (d *Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

func (d *Duration) Scan(src any) error {
	switch data := src.(type) {
	case string:
		tmp, err := time.ParseDuration(data)
		if err != nil {
			return err
		}
		d.Duration = tmp
		return nil
	default:
		return fmt.Errorf("input must be string, found %T", src)
	}
}

func (d *Duration) Value() (driver.Value, error) {
	return d.MarshalJSON()
}
