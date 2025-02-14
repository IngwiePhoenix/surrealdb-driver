package surrealdbdriver

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"hash/fnv"
	"strings"
	"time"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
)

// Basically a ripoff from: https://github.com/go-sql-driver/mysql/blob/341a5a5246835b2ac4b8d36bb12a9dfad70663f4/statement.go#L143
// Only the variable names were slightly changed but...that's that.
// Purpose of this method is to convert the value to something sensible, and error out
// when the value would technically not be compatible anymore.
// Further, the following is not respected yet:
// > If CheckNamedValue returns ErrRemoveArgument, the NamedValue will not be included
// > in the final query arguments. This may be used to pass special options to the query itself.
// >
// > If ErrSkip is returned the column converter error checking path is used for the argument.
// > Drivers may wish to return ErrSkip after they have exhausted their own special cases.
// (via: https://pkg.go.dev/database/sql/driver#NamedValueChecker)
func checkNamedValue(value any) (driver.Value, error) {
	if t, ok := value.(time.Time); ok {
		return `d'` + t.Format(time.RFC3339) + `'`, nil
	}
	return value, nil
	//bytes, err := json.Marshal(value)
	//fmt.Println("!! CONVERTED: ", string(bytes))
	//return string(bytes), err
	/*
		r := reflect.ValueOf(value)
		fmt.Printf("!! converting: %T\n", value)

		switch r.Kind() {
		case reflect.Ptr:
			if r.IsNil() {
				return nil, nil
			} else {
				return checkNamedValue(r.Elem().Interface())
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return r.Int(), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return r.Uint(), nil
		case reflect.Float32, reflect.Float64:
			return r.Float(), nil
		case reflect.Bool:
			return r.Bool(), nil
		case reflect.Slice:
			fmt.Printf("!! 2nd converting: %s\n", r.Type().Elem().Kind())
			switch t := r.Type(); {
			case t == reflect.TypeOf(json.RawMessage{}):
				return value, nil
			case t.Elem().Kind() == reflect.Uint8:
				return r.Bytes(), nil
			/*case t.Elem().Kind() == reflect.String:
			var strSlice []string
			for i := 0; i < r.Len(); i++ {
				strSlice = append(strSlice, r.Index(i).String())
			}
			return &strSlice, nil/
			default:
				return nil, fmt.Errorf("unsupported type %T, a slice of %s", value, t.Elem().Kind())
			}
		case reflect.String:
			return r.String(), nil
		}
		return nil, fmt.Errorf("unsupported type %T, a %s", value, r.Kind())
	*/
}

func assertJsonType(data json.RawMessage) string {
	trimmed := strings.TrimSpace(string(data)) // Remove leading/trailing whitespace
	if len(trimmed) == 0 {
		return "empty"
	}

	switch trimmed[0] {
	case '{':
		return "object"
	case '[':
		return "array"
	case '"':
		return "string"
	case 't', 'f': // true or false
		return "boolean"
	case 'n': // null
		return "null"
	default:
		if (trimmed[0] >= '0' && trimmed[0] <= '9') || trimmed[0] == '-' {
			return "number"
		}
	}
	return "unknown"
}

// Im so dead bro x.x
func IdentifyResponse(req api.Request, data []byte) (any, error) {
	// Create a raw response and make sure it's not empty
	initial := api.GenericResponse[json.RawMessage]{}

	if err := json.Unmarshal(data, &initial); err != nil {
		// Failed to decode initial message
		return nil, err
	}

	if initial.Error != nil {
		// A fatal error (parsing etc)
		errResp := api.FatalErrorResponse{}
		if err := json.Unmarshal(data, &errResp); err != nil {
			// could not decode the error itself O.o?
			return nil, err
		}
		return errResp, nil
	}

	// Grab ID
	id := req.ID
	var rawData json.RawMessage
	if initial.Result == nil {
		rawData = json.RawMessage{}
	} else {
		rawData = *initial.Result
	}

	switch req.Method {
	case "version":
		version := string(rawData)
		return api.VersionResponse{
			Id:     id,
			Result: &version, // <- cheat.
		}, nil

	case "info":
		// Only needed for record authentication (AuthTypeRecord)
		// It is always a single, simple object.
		out := st.Object{}
		err := json.Unmarshal(rawData, &out)
		if err != nil {
			return nil, err
		}
		return api.InfoResponse{
			Id:     id,
			Result: &out,
		}, nil

	case "query":
		out := []api.QueryResult{}
		err := json.Unmarshal(rawData, &out)
		if err != nil {
			return nil, err
		}
		return api.QueryResponse{
			Id:     id,
			Result: &out,
		}, nil

	case "graphql":
		panic("Not implemented")

	// Process NoSQL/CRUD operations
	// - Array of object returns
	case "select", "create", "insert", "merge":
		objs := []st.Object{}
		err := json.Unmarshal(rawData, &objs)
		if err != nil {
			return nil, err
		}
		return api.MultiNoSQLResponse{
			Id:     id,
			Result: &objs,
		}, nil

	// - Single object return
	case "update", "upsert", "delete":
		obj := st.Object{}
		err := json.Unmarshal(rawData, &obj)
		if err != nil {
			return nil, err
		}
		return api.SingleNoSQLResponse{
			Id:     id,
			Result: &obj,
		}, nil

	// - Special: relations
	case "relate", "insert_relation":
		out := api.RelationResult{}
		err := json.Unmarshal(rawData, &out)
		if err != nil {
			return nil, err
		}
		return api.RelationResponse{
			Id:     id,
			Result: &out,
		}, nil

	case "patch":
		out := api.JsonPatchResult{}
		err := json.Unmarshal(rawData, &out)
		if err != nil {
			return nil, err
		}
		return api.PatchResponse{
			Id:     id,
			Result: &out,
		}, nil

	case "signin", "signup":
		token := string(rawData)
		return api.AuthResponse{
			Id:     id,
			Result: &token,
		}, nil

	case "let", "unset", "use":
		return api.VarRequest{
			Id:     id,
			Result: nil,
		}, nil
	}

	panic("reached end of IdentifyResponse(...) unexpectedly")
}

// HACK: Need to decode it too...
func stringToHash(str string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()
}

func countRows(res api.QueryResult) (int64, error) {
	if res.Status != "OK" {
		message := res.Result.(string)
		return 0, errors.New(message)
	}

	if out, ok := res.Result.([]struct {
		Id st.RecordID `json:"id"`
	}); ok {
		// - It's an array. (This should be the default for most queries.)
		return int64(len(out)), nil
	}

	if _, ok := res.Result.(struct {
		Id st.RecordID `json:"id"`
	}); ok {
		// - It's just a single value - so technically, one row.
		return 1, nil
	}

	return 0, errors.New("there was nothing to count...?")
}

/*
// IDK what to do with this, so I'll leave it here.
func handleSingleValueResponse(raw json.RawMessage) any {
	// old impl
	vType := assertJsonType(raw)
	switch vType {
	case "object":
		result := api.QueryResult{}
		if err := json.Unmarshal(initial.Result, &result); err != nil {
			return nil, err
		}
		return api.QueryResponse{
			Id:     id,
			Result: result,
		}, nil

	case "array":
		result := api.BatchResult{}
		if err := json.Unmarshal(initial.Result, &result); err != nil {
			return nil, err
		}
		return api.BatchResponse{
			Id:     id,
			Result: result,
		}, nil

	// Technically, this should never happen.
	// But, I will write it out here and possibly reuse it later.
	// Especially when processing a single-value response.
	// That said, a valid query is also `RETURN 1;` - which, in itself,
	// would _also_ be considered a query response.
	// Therefore, while this _shouldn't_ be neccessary, I'd rather
	// have that present and available in the case that it _does_ happen.
	// Lord knows what dem database peeps do be doin' owo
	case "null":
		return api.NullResponse{
			Id:     id,
			Result: st.Null{}, // I actually should have a method for this. o.o"
		}, nil

	case "boolean":
		b := st.Bool{}
		err := json.Unmarshal(initial.Result, &b)
		if err != nil {
			return nil, err
		}
		return api.GenericResponse[st.Bool]{
			Id:     id,
			Result: b,
		}, nil

	case "string":
		b := st.String{}
		err := json.Unmarshal(initial.Result, &b)
		if err != nil {
			return nil, err
		}
		return api.GenericResponse[st.String]{
			Id:     id,
			Result: b,
		}, nil

	case "number":
		// This is where type hints as per CBOR would have been super useful.
		// However, I doubt we'd get complex numbers in a single-value response.
		// ...right?
		numstr := strings.TrimSpace(string(initial.Result))
		dots := strings.Count(numstr, ".")
		switch dots {
		case 0:
			// int
			var i int
			return api.GenericResponse[st.Int]{
				Id:     id,
				Result: st.Int{V: i},
			}, err
		case 1:
			var f float64
			err := json.Unmarshal(initial.Result, &f)
			if err != nil {
				return nil, err
			}
			return api.GenericResponse[st.Float]{
				Id:     id,
				Result: st.Float{Float64: f},
			}, nil
		default:
			// ...got to be complex. oh no. Let's pray this works.
			var c st.Decimal
			err := json.Unmarshal(initial.Result, &c)
			if err != nil {
				return nil, err
			}
			return api.GenericResponse[st.Decimal]{
				Id:     id,
				Result: c,
			}, nil
		}

	default:
		// aka. "unknown". This shouldn't happen.
		panic("received invalid JSON assertation \"" + vType + "\"")
	}
}
*/
