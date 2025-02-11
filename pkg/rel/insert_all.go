package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

// InsertAll builder.
type InsertAll struct {
	BufferFactory         builder.BufferFactory
	ReturningPrimaryValue bool
}

// Build SQL string and its arguments.
func (ia InsertAll) Build(table string, primaryField string, fields []string, bulkMutates []map[string]rel.Mutate, onConflict rel.OnConflict) (string, []any) {
	buffer := ia.BufferFactory.Create()

	ia.WriteInsertInto(&buffer, table)
	ia.WriteValues(&buffer, fields, bulkMutates)
	ia.WriteOnConflict(&buffer, bulkMutates, onConflict)
	ia.WriteReturning(&buffer, primaryField)
	buffer.WriteString(";")

	return buffer.String(), buffer.Arguments()
}

func (ia InsertAll) WriteInsertInto(buffer *builder.Buffer, table string) {
	buffer.WriteString("INSERT INTO ")
	buffer.WriteTable(table)
}

func (ia InsertAll) WriteValues(buffer *builder.Buffer, fields []string, bulkMutates []map[string]rel.Mutate) {
	// head
	buffer.WriteString(" (")
	for n, field := range fields {
		buffer.WriteEscape(field)
		if n > 0 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(") VALUES ")

	// body
	var (
		fieldsCount  = len(fields)
		mutatesCount = len(bulkMutates)
	)

	for i, mutates := range bulkMutates {
		buffer.WriteByte('(')

		for j, field := range fields {
			if mut, ok := mutates[field]; ok && mut.Type == rel.ChangeSetOp {
				buffer.WriteValue(mut.Value)
			} else {
				// TODO: There is no real way to statically denote a default.
				//       So, let's hope this works?
				buffer.WriteString("/* DEFAULT */")
			}

			if j < fieldsCount-1 {
				buffer.WriteString(", ")
			}
		}

		if i < mutatesCount-1 {
			buffer.WriteString("), ")
		} else {
			buffer.WriteByte(')')
		}
	}
	buffer.WriteString("; ")
}

func (ia InsertAll) WriteReturning(buffer *builder.Buffer, primaryField string) {
	if ia.ReturningPrimaryValue && primaryField != "" {
		buffer.WriteString(" RETURN VALUE ")
		buffer.WriteEscape(primaryField)
	}
}

// Copied from Insert and adopted
func (i InsertAll) WriteOnConflict(buffer *builder.Buffer, bulkMutates []map[string]rel.Mutate, onConflict rel.OnConflict) {
	//         index | field    | numMutates
	realMutates := make([]map[string]int, 0, len(bulkMutates))
	for mutateIndex, mutates := range bulkMutates {
		for fieldName, mutate := range mutates {
			if mutate.Type != rel.ChangeInvalidOp {
				realMutates[mutateIndex][fieldName]++
			}
		}
	}
	if len(realMutates) > 0 {
		buffer.WriteString(" ON DUPLICATE KEY UPDATE ")
		for mutatesIndex := range realMutates {
			buffer.WriteString("(")
			for field, mutate := range bulkMutates[mutatesIndex] {
				// TODO: escape
				buffer.WriteString(field)
				// TODO: I am guessing hard here.
				if mutate.Type == rel.ChangeIncOp {
					buffer.WriteString(" += ")
				} else if mutate.Type == rel.ChangeFragmentOp {
					buffer.WriteString(" = ")
				}
				buffer.WriteValue(mutate.Value)
			}
			// TODO: This comparison is jank.
			if mutatesIndex+1 == len(realMutates) {
				buffer.WriteString(")")
			} else {
				buffer.WriteString("), ")
			}
		}
		buffer.WriteString("; ")
	}
}
