package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

// Insert builder.
type Insert struct {
	BufferFactory         builder.BufferFactory
	ReturningPrimaryValue bool
	InsertDefaultValues   bool
}

// Build sql query and its arguments.
func (i Insert) Build(table string, primaryField string, mutates map[string]rel.Mutate, onConflict rel.OnConflict) (string, []any) {
	buffer := i.BufferFactory.Create()

	i.WriteInsertInto(&buffer, table)
	i.WriteValues(&buffer, mutates)
	i.WriteOnConflict(&buffer, mutates, onConflict)
	i.WriteReturning(&buffer, primaryField)

	buffer.WriteString(";")

	return buffer.String(), buffer.Arguments()
}

func (i Insert) WriteInsertInto(buffer *builder.Buffer, table string) {
	buffer.WriteString("INSERT INTO ")
	buffer.WriteTable(table)
}

func (i Insert) WriteValues(buffer *builder.Buffer, mutates map[string]rel.Mutate) {
	n := 0
	arguments := make([]any, 0, len(mutates))
	buffer.WriteString(" (")
	for field, mut := range mutates {
		if mut.Type != rel.ChangeInvalidOp {
			buffer.WriteEscape(field)
			if n > 0 {
				buffer.WriteString(", ")
			}
			n = n + 1
		}
	}
	buffer.WriteString(") VALUES (")
	n = 0
	for _, mut := range mutates {
		if mut.Type != rel.ChangeInvalidOp {
			buffer.WritePlaceholder()
			arguments = append(arguments, mut.Value)
			if n > 0 {
				buffer.WriteString(", ")
			}
			n = n + 1
		}
	}
	buffer.WriteString(")")
	buffer.AddArguments(arguments...)
}

func (i Insert) WriteReturning(buffer *builder.Buffer, primaryField string) {
	if i.ReturningPrimaryValue && primaryField != "" {
		buffer.WriteString(" RETURN VALUE ")
		buffer.WriteEscape(primaryField)
	}
}

func (i Insert) WriteOnConflict(buffer *builder.Buffer, mutates map[string]rel.Mutate, onConflict rel.OnConflict) {
	var realMutates = 0
	for _, m := range mutates {
		if m.Type != rel.ChangeInvalidOp {
			realMutates++
		}
	}
	if realMutates > 0 {
		buffer.WriteString(" ON DUPLICATE KEY UPDATE")
		for field, mutate := range mutates {
			// TODO: escape
			buffer.WriteString(field)
			// TODO: I am guessing hard here.
			if mutate.Type == rel.ChangeIncOp {
				buffer.WriteString(" += ")
			} else if mutate.Type == rel.ChangeFragmentOp {
				buffer.WriteString(" = ")
			}
			buffer.WritePlaceholder()
			buffer.AddArguments(mutate.Value)
		}
	}
}
