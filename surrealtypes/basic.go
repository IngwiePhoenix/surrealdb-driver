package surrealtypes

import (
	"github.com/clok/kemba"
	geojson "github.com/paulmach/go.geojson"
)

// A small debug helper
var localKemba = kemba.New("surrealdb:surrealtypes")

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
type Bool = bool
type Bytes = []byte

type String = string
type StringArray = []string

// ### numbers
type Int = int
type Float = float64

//type Duration = sql.Null[time.Duration]

// ### Objects
// type Object = gjson.Result
type Geometry = geojson.Geometry
type Literal = []rune // TODO: Go has no type unions...so what do?
type Range = any      // TODO: this needs a custom type
// type Record = Object     // TODO: Actually, this isn't true. in json its string, in db its object!
type Set = []interface{} // TODO: User specified, thus technically generic

// ## Type Constraints
type BasicTypes interface {
	Bool | Bytes | String
}

/*
	type EmptyTypes interface {
		Null | None
	}
*/
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
	BasicTypes /*| EmptyTypes*/ | NumberTypes | TimeTypes | ComplexTypes
}
