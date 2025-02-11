package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

// Delete builder.
type Delete struct {
	BufferFactory builder.BufferFactory
	Query         builder.QueryWriter
	Filter        builder.Filter
}

// Build SQL query and its arguments.
func (ds Delete) Build(table string, filter rel.FilterQuery) (string, []any) {
	buffer := ds.BufferFactory.Create()

	buffer.WriteString("DELETE ")
	buffer.WriteTable(table) // should technically be a "thing"

	if !filter.None() {
		buffer.WriteString(" WHERE ")
		ds.Filter.Write(&buffer, table, filter, ds.Query)
	}

	buffer.WriteString(";")

	return buffer.String(), buffer.Arguments()
}
