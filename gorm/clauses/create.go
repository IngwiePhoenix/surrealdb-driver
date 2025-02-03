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
	sql := fmt.Sprintf("CREATE %s SET ", s.Table)
	params := []interface{}{}

	i := 0
	for key, value := range c.Values {
		if i > 0 {
			sql += ", "
		}
		sql += fmt.Sprintf("%s = ?", key)
		params = append(params, value)
		i++
	}

	builder.WriteString(sql)
	builder.AddVar(builder, params...)
}
