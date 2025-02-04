package surrealtypes

import (
	"database/sql"
	"encoding/json"
	"encoding/text"
	"fmt"
	"math/big"
	"strings"
	"time"

	geojson "github.com/paulmach/go.geojson"
)

// It's here, for now, but idk if I will use an interface for it...
type SurrealType interface {
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
type SurrealAny = any
type SurrealBool = bool
type SurrealBytes = []byte
type SurrealString = string

// ### numbers
type SurrealInt = int
type SurrealFloat = float64

// ### Date and time
type SurrealDateTime = time.Time
type SurrealDuration = time.Duration

// ### Objects
type SurrealGeometry = geojson.Geometry
type SurrealObject = map[string]interface{}
type SurrealLiteral = any          // TODO: Go has no type unions...so what do?
type SurrealRange = any            // TODO: this needs a custom type
type SurrealRecord = SurrealObject // TODO: Actually, this isn't true. in json its string, in db its object!
type SurrealSet = []interface{}    // TODO: User specified, thus technically generic

// ## Complex Types
type SurrealDecimal struct {
	*big.Float
}

var _ json.Marshaler = (*SurrealDecimal)(nil)

func (bf SurrealDecimal) MarshalJSON() ([]byte, error) {
	if bf.Float == nil {
		return []byte("null"), nil
	}
	// Convert to string and then JSON encode it
	return json.Marshal(bf.Float.Text('f', -1)) // 'f' format keeps it in decimal form, -1 means full precision
}
func (bf *SurrealDecimal) UnmarshalJSON(data []byte) error {
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

type SurrealOption[T SurrealNumber] struct {
	value *T
}

func (o SurrealOption[T]) IsEmpty() bool {
	return o.value != nil
}
func (o SurrealOption[T]) Get() T {
	return *o.value
}

// ## Type Constraints
type SurrealBasictypes interface {
	~SurrealBool | ~SurrealBytes | ~SurrealString
}
type SurrealNumberTypes interface {
	~SurrealInt | ~SurrealFloat | SurrealDecimal
}
type SurrealTimeTypes interface {
	SurrealDateTime | SurrealDuration
}
type SurrealComplexTypes interface {
	SurrealObject | SurrealSet | SurrealGeometry
}
type SurrealTypes interface {
	SurrealBasictypes | SurrealNumberTypes | SurrealTimeTypes | SurrealComplexTypes
}

type SimpleResult = interface{}
type ArrayResult = []SimpleResult
type QueryResult struct {
	Status string        `json:"status"`
	Time   time.Duration `json:"time"`
	Result interface{}   `json:"result"`
}
type BatchResult = []QueryResult

type RelationResult struct {
	ID     RecordID      `json:"id"`
	In     RecordID      `json:"in"`
	Out    RecordID      `json:"out"`
	Values SurrealObject `json:"-"`
}

func (r *RelationResult) UnmarshalJSON(data []byte) error {
	var head map[string]interface{}
	if err := json.Unmarshal(data, &head); err != nil {
		return err
	}
	var err error
	r.ID, err = NewRecordIDFromString(head["id"].(string))
	if err != nil {
		return err
	}
	r.In, err = NewRecordIDFromString(head["in"].(string))
	if err != nil {
		return err
	}
	r.Out, err = NewRecordIDFromString(head["out"].(string))
	if err != nil {
		return err
	}

	// clear the recycle bin
	delete(head, "id")
	delete(head, "In")
	delete(head, "out")

	// Grab rest and leave.
	r.Values = head
	return nil
}

type JsonPatchRepr struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}
type JsonPatchResult = []JsonPatchRepr

type ResultTypes interface {
	SimpleResult | ArrayResult | QueryResult | BatchResult
}

type RecordID struct {
	ID    string
	Thing string
}

func NewRecordIDFromString(id string) (RecordID, error) {
	out := RecordID{}
	err := text.Unmarshal([]byte(id), &out)
	return out, err
}

var _ json.Marshaler = (*RecordID)(nil)

func (r *RecordID) String() string {
	return r.ID + ":" + r.Thing
}
func (r *RecordID) UnmarshalText(data []byte) error {
	return r.UnmarshalJSON(data)
}
func (r *RecordID) UnmarshalJSON(data []byte) error {
	var out string
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	parts := strings.SplitN(out, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid ID format: %s", out)
	}
	r.ID = parts[0]
	r.Thing = parts[1]
	return nil
}
func (r *RecordID) MarshalJSON() ([]byte, error) {
	return []byte(r.String()), nil
}

type SurrealError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type GenericResponse[T ResultTypes] struct {
	Id     string       `json:"id,omitempty"`
	Error  SurrealError `json:"error"`
	Result T            `json:"result"`
}

// Helper
type Nullable[T any] struct {
	Value *T
}

// Denotes a definitive fatal error. (Auth, parsing, ...)
type FatalErrorResponse = GenericResponse[Nullable[interface{}]]

// A single entry was returned.
type QueryResponse = GenericResponse[QueryResult]

// More than one query was returned
type BatchResponse = GenericResponse[BatchResult]

// Response after creating a record relation
type RelationResponse = GenericResponse[RelationResult]

// Im so dead bro x.x
func IdentifyResponse(data []byte) GenericResponse[interface{}] {
	var initial struct {
		Id     interface{} `json:"id"`
		Result interface{} `json:"result"`
		Error  interface{} `json:"error"`
	}
	if err := json.Unmarshal(data, &initial); err != nil {
		return err
	}

	if initial.Error != nil {
		var errResp ErrorResponse
		if err := json.Unmarshal(data, &errResp); err != nil {
			return err
		}
	}

	switch r.originMethod {
	case "select", "query", "insert", "create", "update", "upsert", "merge", "delete":
		// .result = object | []object
		var tmp struct {
			Result interface{} `json:"result"`
		}
		out := BasicResponse{}
		if err := json.Unmarshal(data, &tmp); err != nil {
			return err
		}
		if _, isInterfaceArray := tmp.Result.([]interface{}); isInterfaceArray {
			out = BatchResponse{BasicResponse: out}
		}
	}

}

/*
	type ErrorResponse struct {
		BasicResponse
		Error struct {
			Code    int
			Message string
		}
	}

	type InfoResponse struct {
		BasicResponse
		Result struct {
			Version   string
			Build     string
			Timestamp string
		}
	}

	type QueryResponse struct {
		BasicResponse
		Result interface{}
	}

	type BatchResponse struct {
		BasicResponse
		Result []interface{}
	}
*/
type StringMap map[string]string

var _ sql.Scanner = (*StringMap)(nil)

func (q *StringMap) Scan(val interface{}) error {
	if *q == nil {
		*q = make(map[string]string)
	}

	mapVals, ok := val.(map[string]interface{})
	if !ok {
		return fmt.Errorf("could not convert %T to a map[string]interface{}", val)
	}
	for k, ival := range mapVals {
		if sval, ok := ival.(string); ok {
			(*q)[k] = sval
		} else {
			return fmt.Errorf("could not convert %T to a string", ival)
		}
	}
	return nil
}

type StringObject = map[string]StringMap
