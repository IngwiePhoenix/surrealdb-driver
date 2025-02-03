package gorm

import "gorm.io/gorm/clause"

func CallbackToStructClause[
	targetClause clause.Expression,
	originClause clause.Expression,
]() clause.ClauseBuilder {
	return func(c clause.Clause, b clause.Builder) {
		if _, ok := c.Expression.(originClause); ok {
			x := new(targetClause)
			(*x).Build(b)
		}
	}
}
