package surrealtypes

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/oklog/ulid/v2"
	"github.com/tidwall/gjson"
)

const (
	// Footgun: Those aren't paranthesis, brackets, or anything alike.
	// ... they're _unicode_. o.o
	SRIDOpen  rune = '⟨'
	SRIDClose rune = '⟩'
)

type SurrealDBRecordID interface {
	SurrealString() string
}

type IDThings interface {
	int64 | float64 | []rune | gjson.Result | ulid.ULID | uuid.UUID | time.Time | AutoIDFunc
}

func ParseID(in string) (SurrealDBRecordID, error) {
	k := localKemba.Extend("NewRecord")

	isBracket := func(b []rune) bool {
		// can i avoid the copy?
		return b[0] == SRIDOpen && b[len(b)-1] == SRIDClose
	}
	isTicks := func(b []rune) bool {
		return b[0] == '`' && b[len(b)-1] == '`'
	}
	var left, right []rune
	var hasId bool = false
	var srid SurrealDBRecordID
	for _, b := range in {
		if b == ':' {
			hasId = true
			continue
		}
		if !hasId {
			left = append(left, b)
		} else {
			right = append(right, b)
		}
	}
	k.Printf("Scanned <%s> : <%s> (%s)", string(left), string(right), string(in))
	if len(left) <= 0 || len(right) <= 0 {
		return nil, fmt.Errorf("unaligned RecordID: %v, %v : %v", left, right, in)
	}
	if isBracket(right) || isTicks(right) {
		// -> tablename:`abc-def-ghi`
		// -> tablename:⟨abc-def-ghi⟩
		// Assume a string value.
		k.Log("Raw ID")
		srid = RawID{ID: string(left), Thing: right}
	} else if i, err := strconv.ParseInt(string(right), 10, 64); err == nil {
		k.Log("Integer ID")
		srid = IntID{
			ID:    string(left),
			Thing: i,
		}
	} else if f, err := strconv.ParseFloat(string(right), 64); err == nil {
		k.Log("Float ID")
		srid = FloatID{
			ID:    string(left),
			Thing: f,
		}
	} else if ulid_id, err := ulid.ParseStrict(string(right)); err == nil {
		k.Log("ULID ID")
		srid = ULIDID{
			ID:    string(left),
			Thing: ulid_id,
		}
	} else if uuid_id, err := uuid.FromString(string(right)); err == nil {
		k.Log("UUID ID")
		srid = UUIDID{
			ID:    string(left),
			Thing: uuid_id,
		}
	} else if gjson.Valid(string(right)) {
		k.Log("Object ID")
		srid = ObjectID{
			ID:    string(left),
			Thing: gjson.Parse(string(right)),
		}
	} else {
		k.Log("String ID (fallthrough)")
		srid = StringID{ID: string(left), Thing: string(right)}
	}
	// TODO: Range
	// Any other formatting is literally SurrealQL and I can not parse that.
	return srid, nil
}
