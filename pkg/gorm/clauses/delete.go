package clauses

import (
	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Delete)(nil)

type Delete struct {
	Only bool
}

func (c Delete) Build(builder clause.Builder) {
	builder.WriteString("DELETE")

	if c.Only {
		builder.WriteByte(' ')
		builder.WriteString("ONLY")
	}
}
