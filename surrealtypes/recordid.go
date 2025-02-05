package surrealtypes

import (
	"encoding"
	"encoding/json"
	"fmt"
	"strings"
)

type RecordID struct {
	ID    string
	Thing string
}

func NewRecordIDFromString(id string) (RecordID, error) {
	out := &RecordID{}
	err := out.UnmarshalText([]byte(id))
	// Now that feels like a proper hack lmao
	return *out, err
}

var _ json.Marshaler = (*RecordID)(nil)
var _ encoding.TextUnmarshaler = (*RecordID)(nil)
var _ encoding.TextMarshaler = (*RecordID)(nil)

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

func (r *RecordID) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *RecordID) MarshalJSON() ([]byte, error) {
	return []byte(r.String()), nil
}
