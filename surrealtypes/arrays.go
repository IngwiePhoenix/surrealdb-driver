package surrealtypes

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

// ArrayOf is a generic type that implements JSON and SQL interfaces.
type ArrayOf[T any] struct {
	Values []T
}

/*
// Generics can't adhere to an interface. sadge.
var _ json.Marshaler = (*ArrayOf[T])(nil)
var _ json.Unmarshaler = (*ArrayOf[T])(nil)
var _ driver.Valuer = (*ArrayOf[T])(nil)
var _ sql.Scanner = (*ArrayOf[T])(nil)
var _ SurrealMarshalable = (*ArrayOf[T])(nil)
*/

// MarshalJSON implements json.Marshaler
func (a ArrayOf[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Values)
}

// UnmarshalJSON implements json.Unmarshaler
func (a *ArrayOf[T]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &a.Values)
}

// Value implements driver.Valuer, returning a JSON-encoded value.
func (a ArrayOf[T]) Value() (driver.Value, error) {
	return json.Marshal(a.Values)
}

// Scan implements sql.Scanner, parsing JSON arrays from []byte.
func (a *ArrayOf[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("expected []byte, got %T", value)
	}

	if !gjson.ValidBytes(bytes) {
		return fmt.Errorf("invalid JSON: %s", string(bytes))
	}

	return json.Unmarshal(bytes, &a.Values)
}
