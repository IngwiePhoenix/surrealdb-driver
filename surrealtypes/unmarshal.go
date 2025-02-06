package surrealtypes

import (
	"encoding/json"

	"github.com/IngwiePhoenix/surrealdb-driver/utils"
)

func extractKeys(data []byte) ([]string, error) {
	var temp map[string]json.RawMessage
	if err := json.Unmarshal(data, &temp); err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(temp))
	for key := range temp {
		keys = append(keys, key)
	}
	return keys, nil
}

func allStringsInSlice(haystack []string, targets []string) bool {
	for _, target := range targets {
		if utils.IndexOfString(haystack, target) == -1 {
			return false
		}
	}
	return true
}
