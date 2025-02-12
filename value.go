package surrealdbdriver

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
)

// isMixedTypeSet checks if a slice contains mixed types.
func allMixed(input []interface{}) bool {
	if len(input) == 0 {
		return false // Empty slices are trivially homogeneous
	}

	firstType := reflect.TypeOf(input[0])

	for _, item := range input[1:] {
		if reflect.TypeOf(item) != firstType {
			return true // Found a mixed type
		}
	}

	return false // All elements are of the same type
}

func allOf[T any](input []interface{}) bool {
	if len(input) == 0 {
		return true // Empty slices are trivially homogeneous
	}

	// Get the type of the first element
	targetType := reflect.TypeOf((*T)(nil)).Elem() // Extract type of T

	for _, item := range input {
		if reflect.TypeOf(item) != targetType {
			return false
		}
	}

	return true
}

func convertValue(input any) (any, error) {
	// check in this order: Special, Complex, Primitive

	// Short-circuit: is empty?
	if input == nil {
		// should probably use st.Null ?
		return nil, nil
	}

	// time.Time
	if s, ok := input.(string); ok {
		t, err := time.Parse(time.RFC3339Nano, s)
		if !errors.Is(err, &time.ParseError{}) {
			// We ignore parse errors, everything else is an issue.
			return nil, err
		}
		return t, nil
	}

	// time.Duration
	if s, ok := input.(string); ok {
		t, err := time.ParseDuration(s)
		if !errors.Is(err, &time.ParseError{}) {
			// We ignore parse errors, everything else is an issue.
			return nil, err
		}
		return st.Duration{Duration: t}, nil
	}

	// GeoJSON
	if g, ok := input.(st.Geometry); ok {
		return g, nil
	}

	// Array or Set?
	if a, ok := input.([]interface{}); ok {
		// Is it empty?
		if len(a) == 0 {
			// idk bruh
			return a, nil
		}

		// Is it a Set?
		if allMixed(a) {
			out := st.Set{}
			copy(out, a)
			return out, nil
		}

		// It's an array, determine the type.
		if allOf[st.Bool](a) {
			out := make([]st.Bool, len(a))
			for i, v := range a {
				out[i] = v.(st.Bool)
			}
			return out, nil
		}

		if allOf[st.Bytes](a) {
			out := make([]st.Bytes, len(a))
			for i, v := range a {
				out[i] = v.(st.Bytes)
			}
			return out, nil
		}

		if allOf[st.Int](a) {
			out := make([]st.Int, len(a))
			for i, v := range a {
				out[i] = v.(st.Int)
			}
			return out, nil
		}

		if allOf[st.Float](a) {
			out := make([]st.Float, len(a))
			for i, v := range a {
				out[i] = v.(st.Float)
			}
			return out, nil
		}

		if allOf[st.String](a) {
			out := make([]st.String, len(a))
			for i, v := range a {
				out[i] = v.(st.String)
			}
			return out, nil
		}

		// SKIP: NONE
	}

	// It's likely an object at this point.
	if o, ok := input.(st.Object); ok {
		out := map[string]any{}
		for key, value := range o {
			v, err := convertValue(value)
			if err != nil {
				return nil, err
			}
			out[key] = v
		}
		return out, nil
	}

	panic(fmt.Sprintf("did not match anything for %T", input))
}
