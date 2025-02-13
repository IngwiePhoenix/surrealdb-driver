package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/builder"
)

// Update builder.
type Update struct {
	BufferFactory builder.BufferFactory
	Query         builder.QueryWriter
	Filter        Filter
}

var _ (sql.UpdateBuilder) = (*Update)(nil)

// Build SQL string and it arguments.
func (u Update) Build(table string, primaryField string, mutates map[string]rel.Mutate, filter rel.FilterQuery) (string, []any) {
	buffer := u.BufferFactory.Create()

	buffer.WriteString("UPDATE ")
	buffer.WriteTable(table)
	buffer.WriteString(" SET ")

	i := 0
	for field, mut := range mutates {
		if field == primaryField {
			continue
		}

		if i > 0 {
			buffer.WriteByte(',')
		}
		i++

		switch mut.Type {
		case rel.ChangeSetOp:
			buffer.WriteEscape(field)
			buffer.WriteString(" = ")
			buffer.WriteValue(mut.Value)
		case rel.ChangeIncOp:
			buffer.WriteEscape(field)
			buffer.WriteString(" += ")
			//buffer.WriteEscape(field)
			//buffer.WriteByte('+')
			buffer.WriteValue(mut.Value)
		case rel.ChangeFragmentOp:
			// TODO: Should I write it out verbatim?
			buffer.WriteString(field)
			buffer.AddArguments(mut.Value.([]any)...)
		}
	}

	if !filter.None() {
		buffer.WriteString(" WHERE ")
		u.Filter.Write(&buffer, table, filter, u.Query)
	}

	buffer.WriteString("; ")

	return buffer.String(), buffer.Arguments()
}
