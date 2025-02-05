package surrealtypes

import (
	"database/sql"
	"fmt"
)

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
