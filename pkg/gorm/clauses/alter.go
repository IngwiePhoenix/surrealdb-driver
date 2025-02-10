package clauses

import (
	"gorm.io/gorm/clause"
)

var _ clause.Expression = (*Alter)(nil)

type Alter struct {
	Table  string
	Values map[string]interface{}
}

func (c Alter) Build(builder clause.Builder) {
	// TODO: This might actually not be needed.
}
