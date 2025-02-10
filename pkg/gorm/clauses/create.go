package clauses

import (
	"fmt"

	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Create)(nil)

type Create struct {
	Table  string
	Values map[string]interface{}
}

func (c Create) Build(builder clause.Builder) {
	// TODO: quote table name or sanitize it otherwise.
	// Also, how do I best handle RecordIDs (foo:bar)?
	sql := fmt.Sprintf("CREATE %s SET ", c.Table)
	params := []interface{}{}

	i := 0
	for key, value := range c.Values {
		if i > 0 {
			sql += ", "
		}
		// TODO: Sanitize key (and possibly quote it)
		sql += fmt.Sprintf("%s = ?", key)
		params = append(params, value)
		i++
	}

	builder.WriteString(sql)
	builder.AddVar(builder, params...)
}
