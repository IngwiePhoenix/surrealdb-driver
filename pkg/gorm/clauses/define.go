package clauses

import (
	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Define)(nil)

type Define struct {
	Table  string
	Values map[string]interface{}
}

func (c Define) Build(builder clause.Builder) {
	// TODO: This might actually not be needed.
}
