package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/builder"
)

// Insert builder.
type Insert struct {
	BufferFactory         builder.BufferFactory
	ReturningPrimaryValue bool
	InsertDefaultValues   bool
}

var _ (sql.InsertBuilder) = (*Insert)(nil)

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
	// Maps and order are kind of a thing...
	var fields []string
	for k := range mutates {
		fields = append(fields, k)
	}

	n := 0
	buffer.WriteString(" (")
	for _, field := range fields {
		mut := mutates[field]
		if mut.Type != rel.ChangeInvalidOp {
			if n > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteEscape(field)
			n = n + 1
		}
	}

	buffer.WriteString(") VALUES (")
	n = 0
	for _, field := range fields {
		mut := mutates[field]
		if mut.Type != rel.ChangeInvalidOp {
			if n > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteValue(mut.Value)
			n = n + 1
		}
	}
	buffer.WriteString(")")
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
		buffer.WriteString(" ON DUPLICATE KEY UPDATE ")
		var i int = 0
		for field, mutate := range mutates {
			if mutate.Type == rel.ChangeInvalidOp {
				continue
			}
			if i > 0 {
				buffer.WriteString(", ")
			}
			// TODO: escape
			buffer.WriteString(field)
			// TODO: I am guessing hard here.
			switch mutate.Type {
			case rel.ChangeFragmentOp:
				buffer.WriteValue(mutate.Value)
			case rel.ChangeSetOp:
				buffer.WriteString(" = ")
				buffer.WriteValue(mutate.Value)
			case rel.ChangeIncOp:
				// TODO: "inc" probably isn't just... this.
				buffer.WriteString(" += ")
				buffer.WriteValue(mutate.Value)
			}
			i++
		}
	}
}
