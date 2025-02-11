package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

/*
	This is missing a few of SurrealDB's custom comparison operators.
	I should probably add them somehow.
	Not implemented/accounted for:
	- !!a - determine truthyness
	- a??b - is either a or b truthy and not null?
	- a?:b - same as above, minus null
	- a=b - lax equal (a==b is strict)
	- a?=b - if any of a is equal to any of b
	- a*=b - all of a must equal b
	- a~b - fuzzy matching compare
	- a!~b - same as above but negative
	- a?~b - same as above but for sets
	- a*~b - same as a*=b but fuzzy
	- a CONTAINS b - is a inside b?
	- a CONTAINSNOT b - above but negate
	- a CONTAINSALL b - above but must match all
	- a CONTAINSANY b - kinda like CONTAINS, must at least be 1
	- b CONTAINSNONE b - like CONTAINSNOT
	- a OUTSIDE/INTERSECTS b - geo compare
	- a@@b - is a part of b's index?
	-
*/

// Filter builder.
type Filter struct{}

// Write SQL to buffer.
func (f Filter) Write(buffer *builder.Buffer, table string, filter rel.FilterQuery, queryWriter builder.QueryWriter) {
	switch filter.Type {
	case rel.FilterAndOp:
		f.WriteLogical(buffer, table, "AND", filter.Inner, queryWriter)
	case rel.FilterOrOp:
		f.WriteLogical(buffer, table, "OR", filter.Inner, queryWriter)
	case rel.FilterNotOp:
		buffer.WriteString("NOT ")
		f.WriteLogical(buffer, table, "AND", filter.Inner, queryWriter)
	case rel.FilterEqOp,
		rel.FilterNeOp,
		rel.FilterLtOp,
		rel.FilterLteOp,
		rel.FilterGtOp,
		rel.FilterGteOp:
		f.WriteComparison(buffer, table, filter, queryWriter)
	case rel.FilterNilOp:
		buffer.WriteField(table, filter.Field)
		buffer.WriteString(" IS NULL")
	case rel.FilterNotNilOp:
		buffer.WriteField(table, filter.Field)
		buffer.WriteString(" IS NOT NULL")
	case rel.FilterInOp,
		rel.FilterNinOp:
		f.WriteInclusion(buffer, table, filter, queryWriter)
	case rel.FilterLikeOp:
		buffer.WriteField(table, filter.Field)
		buffer.WriteString(" INSIDE ")
		buffer.WriteValue(filter.Value)
	case rel.FilterNotLikeOp:
		buffer.WriteField(table, filter.Field)
		buffer.WriteString(" NOTINSIDE ")
		buffer.WriteValue(filter.Value)
	case rel.FilterFragmentOp:
		buffer.WriteString(filter.Field)
		if !buffer.InlineValues {
			buffer.AddArguments(filter.Value.([]any)...)
		}
	}
}

// WriteLogical SQL to buffer.
func (f Filter) WriteLogical(buffer *builder.Buffer, table, op string, inner []rel.FilterQuery, queryWriter builder.QueryWriter) {
	var (
		length = len(inner)
	)

	if length > 1 {
		buffer.WriteByte('(')
	}

	for i, c := range inner {
		f.Write(buffer, table, c, queryWriter)

		if i < length-1 {
			buffer.WriteByte(' ')
			buffer.WriteString(op)
			buffer.WriteByte(' ')
		}
	}

	if length > 1 {
		buffer.WriteByte(')')
	}
}

// WriteComparison SQL to buffer.
func (f Filter) WriteComparison(buffer *builder.Buffer, table string, filter rel.FilterQuery, queryWriter builder.QueryWriter) {
	buffer.WriteField(table, filter.Field)

	switch filter.Type {
	case rel.FilterEqOp:
		buffer.WriteString("==")
	case rel.FilterNeOp:
		buffer.WriteString("!=")
	case rel.FilterLtOp:
		buffer.WriteByte('<')
	case rel.FilterLteOp:
		buffer.WriteString("<=")
	case rel.FilterGtOp:
		buffer.WriteByte('>')
	case rel.FilterGteOp:
		buffer.WriteString(">=")
	}

	switch v := filter.Value.(type) {
	case rel.SubQuery:
		// For warped sub-queries
		f.WriteSubQuery(buffer, v, queryWriter)
	case rel.Query:
		// For sub-queries without warp
		f.WriteSubQuery(buffer, rel.SubQuery{Query: v}, queryWriter)
	default:
		// For simple values
		buffer.WriteValue(filter.Value)
	}
}

// WriteInclusion SQL to buffer.
func (f Filter) WriteInclusion(buffer *builder.Buffer, table string, filter rel.FilterQuery, queryWriter builder.QueryWriter) {
	var (
		values = filter.Value.([]any)
	)

	if len(values) == 0 {
		if filter.Type == rel.FilterInOp {
			buffer.WriteString("1=0")
		} else {
			buffer.WriteString("1=1")
		}
	} else {
		buffer.WriteField(table, filter.Field)

		if filter.Type == rel.FilterInOp {
			buffer.WriteString(" INSIDE ")
		} else {
			buffer.WriteString(" NOTINSIDE ")
		}

		f.WriteInclusionValues(buffer, values, queryWriter)
	}
}

func (f Filter) WriteInclusionValues(buffer *builder.Buffer, values []any, queryWriter builder.QueryWriter) {
	if len(values) == 1 {
		if value, ok := values[0].(rel.Query); ok {
			f.WriteSubQuery(buffer, rel.SubQuery{Query: value}, queryWriter)
			return
		}
	}

	buffer.WriteByte('(')
	for i := 0; i < len(values); i++ {
		if i > 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteValue(values[i])
	}
	buffer.WriteByte(')')
}

func (f Filter) WriteSubQuery(buffer *builder.Buffer, sub rel.SubQuery, queryWriter builder.QueryWriter) {
	buffer.WriteString(sub.Prefix)
	buffer.WriteByte('(')
	queryWriter.Write(buffer, sub.Query)
	buffer.WriteByte(')')
}
