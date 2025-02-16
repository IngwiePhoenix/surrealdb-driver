package surrealtypes

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math/big"

	"github.com/goccy/go-json"
)

type Decimal struct {
	*big.Float
}

var _ json.Marshaler = (*Decimal)(nil)
var _ json.Unmarshaler = (*Decimal)(nil)
var _ driver.Valuer = (*Decimal)(nil)
var _ sql.Scanner = (*Decimal)(nil)

func (bf Decimal) MarshalJSON() ([]byte, error) {
	if bf.Float == nil {
		return []byte("0"), nil
	}
	// Convert to string and then JSON encode it
	return json.Marshal(bf.Float.Text('f', -1)) // 'f' format keeps it in decimal form, -1 means full precision
}
func (bf *Decimal) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	// TODO: Set precision
	f, _, err := big.ParseFloat(str, 10, 256, big.ToNearestEven)
	if err != nil {
		return err
	}
	bf.Float = f
	return nil
}

func (d *Decimal) Scan(src any) error {
	switch data := src.(type) {
	case string:
		f, _, err := big.ParseFloat(data, 10, 256, big.ToNearestEven)
		if err != nil {
			return err
		}
		d.Float = f
		return nil
	default:
		return fmt.Errorf("input must be string, found %T", src)
	}
}

func (d *Decimal) Value() (driver.Value, error) {
	return d.MarshalJSON()
}
