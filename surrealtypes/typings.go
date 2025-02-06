package surrealtypes

import (
	"database/sql"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	geojson "github.com/paulmach/go.geojson"
)

// It's here, for now, but idk if I will use an interface for it...
type SurrealDBType interface {
	DBTypeName() string
}

//
// # SurrealDB Type Mapping
// These types are a mapping of the SurrealDB types into Go.
// this can become helpful when defining schemas or working with JSON.
// further, it might come in handy for CBOR down the line...
//

// TODO: Properly work this out
// From and with: https://surrealdb.com/docs/surrealql/datamodel
// ## Simple
// ### basic
type Any = any
type Bool = sql.NullBool
type Bytes = []sql.NullBool
type String = sql.NullString

// ### numbers
type Int = sql.Null[int]
type Float = sql.NullFloat64

// ### Date and time
type DateTime = sql.NullTime

//type Duration = sql.Null[time.Duration]

// ### Objects
type Geometry = sql.Null[geojson.Geometry]
type Object = map[string]interface{}
type Literal = any       // TODO: Go has no type unions...so what do?
type Range = any         // TODO: this needs a custom type
type Record = Object     // TODO: Actually, this isn't true. in json its string, in db its object!
type Set = []interface{} // TODO: User specified, thus technically generic

// ## Complex Types
// ...because someone /had/ to add more to the pile.
// In SurrealDB, none != null. x.x
// ```
// > RETURN (none == null)
// -- Query 1 (execution time: 59.158µs)
// false
// ```
// Layout inspired by sql.Null
type None struct {
	isNone bool
}

func (n *None) UnmarshalJSON(data []byte) error {
	v := strings.TrimSpace(strings.ToLower(string(data)))
	n.isNone = v == "none"
	return nil
}
func (n *None) MarshalJSON() ([]byte, error) {
	return []byte("NONE"), nil
}

// The fact that this is required kinda drives me nuts. o.o...
type Null struct {
	isNull bool
}

func (n *Null) UnmarshalJSON(data []byte) error {
	v := strings.TrimSpace(strings.ToLower(string(data)))
	n.isNull = v == "null"
	return nil
}
func (n *Null) MarshalJSON() ([]byte, error) {
	return []byte("null"), nil
}

type Decimal struct {
	*big.Float
}

var _ json.Marshaler = (*Decimal)(nil)

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

type Duration struct {
	time.Duration
}

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
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

type Option[T NumberTypes] struct {
	value *T
}

func (o Option[T]) IsEmpty() bool {
	return o.value != nil
}
func (o Option[T]) Get() T {
	return *o.value
}

// ## Type Constraints
type BasicTypes interface {
	Bool | Bytes | String
}
type EmptyTypes interface {
	Null | None
}
type NumberTypes interface {
	Int | Float | Decimal
}
type TimeTypes interface {
	DateTime | Duration
}
type ComplexTypes interface {
	Object | Set | Geometry
}
type Types interface {
	BasicTypes | EmptyTypes | NumberTypes | TimeTypes | ComplexTypes
}
