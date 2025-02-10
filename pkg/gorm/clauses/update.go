package clauses

import (
	"fmt"

	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Update)(nil)

type Update struct {
	Modifier string
	Table    clause.Table
	Values   map[string]interface{}
}

func (c Update) Build(builder clause.Builder) {
	// TODO: quote table name or sanitize it otherwise.
	// Also, how do I best handle RecordIDs (foo:bar)?
	builder.WriteString("UPDATE ")
	if c.Modifier != "" {
		builder.WriteString(c.Modifier)
		builder.WriteByte(' ')
	}
	builder.WriteQuoted(c.Table)
	builder.WriteString(" SET ")

	params := []interface{}{}

	i := 0
	sql := ""
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
