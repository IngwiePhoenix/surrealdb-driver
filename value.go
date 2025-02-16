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
	str := input.Str
	if strings.Contains(str, ".") || strings.ContainsAny(str, "eE") {
		if f, err := strconv.ParseFloat(str, 64); err == nil {
			return f, nil
		} else {
			return nil, err
		}
	} else {
		// Try parsing as int
		if i, err := strconv.ParseInt(str, 10, 64); err == nil {
			return i, nil
		} else if f, err := strconv.ParseFloat(str, 64); err == nil && (f > math.MaxInt64 || f < math.MinInt64) {
			return f, nil
		}
	}

	// Fallback: parse as float (shouldn't reach here under normal circumstances)
	f, err := strconv.ParseFloat(str, 64)
	return f, err
}

func convertValue(input gjson.Result) (driver.Value, error) {
	switch input.Type {
	case gjson.Null:
		return nil, nil
	case gjson.JSON:
		return []byte(input.Raw), nil
	case gjson.True, gjson.False:
		return input.Bool(), nil
	case gjson.Number:
		return gjsonNumberToDriverValue(input)
	case gjson.String:
		if t, err := time.Parse(time.RFC3339Nano, input.String()); err == nil {
			return t, nil
		} else if t, err := time.ParseDuration(input.String()); err == nil {
			return t, nil
		} else {
			return input.String(), nil
		}
	}
	panic("convertValue: did fall through entirely")
}
