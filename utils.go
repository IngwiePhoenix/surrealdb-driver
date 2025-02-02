package surrealdbdriver

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"
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
		return &strSlice, nil*/
		default:
			return nil, fmt.Errorf("unsupported type %T, a slice of %s", value, t.Elem().Kind())
		}
	case reflect.String:
		return r.String(), nil
	}
	return nil, fmt.Errorf("unsupported type %T, a %s", value, r.Kind())
}

type AuthMethod string

const (
	AuthMethodRoot      AuthMethod = "root"
	AuthMethodDB        AuthMethod = "db"
	AuthMethodRecord    AuthMethod = "record"
	AuthMethodUnknown   AuthMethod = "unknown"
	AuthMethodToken     AuthMethod = "token"
	AuthMethodAnonymous AuthMethod = "none"
)
