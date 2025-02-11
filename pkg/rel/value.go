package rel

import (
	"database/sql/driver"
	"encoding/json"
)

type ValueConvert struct{}

func (c ValueConvert) ConvertValue(v interface{}) (driver.Value, error) {
	// Naively marshall stuff.
	// I kinda have to get a feel for this first.
	// But, SurrealQL is kinda-sorta subset-ish of JSON.
	// So, in most cases, valid JSON is valid SurrealQL.
	return json.Marshal(v)
}
