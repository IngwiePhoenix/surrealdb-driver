package clauses

import (
	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Insert)(nil)

type Insert struct {
	Table    clause.Table
	Modifier string
}

func (c Insert) Build(builder clause.Builder) {
	if c.Modifier != "" {
		builder.WriteString(c.Modifier)
		builder.WriteByte(' ')
	}

	builder.WriteString("INTO ")
	// TODO: Sanitize?
	builder.WriteQuoted(c.Table)
}
