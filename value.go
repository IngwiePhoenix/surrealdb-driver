package surrealdbdriver

import (
	"database/sql/driver"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/tidwall/gjson"
)

func gjsonNumberToDriverValue(input gjson.Result) (driver.Value, error) {
	k := localKemba.Extend("gjsonNumberToDriverValue")

	str := input.String()
	if strings.Contains(str, ".") || strings.ContainsAny(str, "eE") {
		k.Printf("attempting to parse as float (%s)\n", str)
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, nil
		} else {
			return nil, err
		}
	} else {
		k.Printf("attempting to parse as int (%s)\n", str)
		// Try parsing as int
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, nil
		} else if f, err := strconv.ParseFloat(str, 64); err == nil && (f > math.MaxInt64 || f < math.MinInt64) {
			return f, nil
		}
	}

	// Fallback: parse as float (shouldn't reach here under normal circumstances)
	k.Printf("falling through! (%s)\n", str)
	f, err := strconv.ParseFloat(str, 64)
	return f, err
}

func convertValue(input gjson.Result) (driver.Value, error) {
	k := localKemba.Extend("convertValue")
	switch input.Type {
	case gjson.Null:
		k.Log("Converting NULL")
		return nil, nil
	case gjson.JSON:
		k.Log("Converting JSON")
		return []byte(input.Raw), nil
	case gjson.True, gjson.False:
		k.Log("Converting boolean")
		return input.Bool(), nil
	case gjson.Number:
		k.Log("Converting number")
		return gjsonNumberToDriverValue(input)
	case gjson.String:
		k.Log("Converting string")
		k.Log("...is it a time, duration or just a string?")
		if t, err := time.Parse(time.RFC3339Nano, input.String()); err == nil {
			k.Printf("it's time: %v\n", t)
			return t, nil
		} else if t, err := time.ParseDuration(input.String()); err == nil {
			k.Printf("it's duration: %v\n", t)
			return t, nil
		} else {
			k.Printf("it's string: %v\n", input.String())
			return input.String(), nil
		}
	}
	panic("convertValue: did fall through entirely")
}

//func surrealizeValue(in any) driver.Value {}
