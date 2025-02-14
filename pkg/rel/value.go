package rel

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var _ (driver.ValueConverter) = (*ValueConvert)(nil)

type ValueConvert struct{}

func (c ValueConvert) stringifySlice(s []string) (string, error) {
	strs := []string{}
	for _, vv := range s {
		str, err := c.ConvertValue(vv)
		if err != nil {
			return "", err
		}
		strs = append(strs, str.(string))
	}
	final := strings.Join(strs, ", ")
	fmt.Println("!! joined: ", final)
	return "[" + final + "]", nil
}

func (c ValueConvert) ConvertValue(v interface{}) (driver.Value, error) {
	fmt.Printf("!! ValueConvert/rel called: %T\n", v)
	if d, ok := v.(time.Time); ok {
		return `d'` + d.Format(time.RFC3339) + `'`, nil
	} else if s, ok := v.(string); ok {
		fmt.Println("!! Left it a string.")
		return "\"" + s + "\"", nil
		/*str, err := json.Marshal(s)
		if err != nil {
			return nil, err
		}
		return string(str), nil*/
	} else if sa, ok := v.([]string); ok {
		fmt.Println("!! string array")
		return c.stringifySlice(sa)
	} else if s, ok := v.(interface{ String() string }); ok {
		fmt.Println("!! It implements .String()string")
		return s.String(), nil
	}
	fmt.Println("!! We don't give a fuck.")
	return json.Marshal(v)
}
