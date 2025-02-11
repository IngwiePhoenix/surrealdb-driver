package rel

import (
	"strconv"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

// Query builder.
type Query struct {
	BufferFactory builder.BufferFactory
	Filter        Filter
}

// Build SQL string and it arguments.
func (q Query) Build(query rel.Query) (string, []any) {
	buffer := q.BufferFactory.Create()

	q.Write(&buffer, query)

	return buffer.String(), buffer.Arguments()
}

// Write SQL to buffer.
func (q Query) Write(buffer *builder.Buffer, query rel.Query) {
	if query.SQLQuery.Statement != "" {
		buffer.WriteString(query.SQLQuery.Statement)
		buffer.AddArguments(query.SQLQuery.Values...)
		return
	}

	rootQuery := buffer.Len() == 0

	q.WriteSelect(buffer, query.Table, query.SelectQuery)
	q.WriteQuery(buffer, query)

	if rootQuery {
		buffer.WriteByte(';')
	}
}

// WriteSelect SQL to buffer.
func (q Query) WriteSelect(buffer *builder.Buffer, table string, selectQuery rel.SelectQuery) {
	if len(selectQuery.Fields) == 0 {
		buffer.WriteString("SELECT ")
		/*if selectQuery.OnlyDistinct {
			buffer.WriteString("array::distinct(")
		}*/
		buffer.WriteField(table, "*")
		/*if selectQuery.OnlyDistinct {
			buffer.WriteString(")")
		}*/
		return
	}

	buffer.WriteString("SELECT ")

	/*if selectQuery.OnlyDistinct {
		buffer.WriteString("DISTINCT ")
	}*/

	l := len(selectQuery.Fields) - 1
	for i, f := range selectQuery.Fields {
		buffer.WriteField(table, f)

		if i < l {
			buffer.WriteByte(',')
		}
	}
}

// WriteQuery SQL to buffer.
func (q Query) WriteQuery(buffer *builder.Buffer, query rel.Query) {
	q.WriteFrom(buffer, query.Table)
	// SurrealDB has no joins - it just has relations.
	// In this case, we can use the structure of the JoinQuery to our advantage.
	// Basically, rewrite them as FETCH $field[, $field[, ...]]
	//q.WriteJoin(buffer, query.Table, query.JoinQuery)
	q.WriteWhere(buffer, query.Table, query.WhereQuery)

	if len(query.GroupQuery.Fields) > 0 {
		q.WriteGroupBy(buffer, query.Table, query.GroupQuery.Fields)
		q.WriteHaving(buffer, query.Table, query.GroupQuery.Filter)
	}

	q.WriteOrderBy(buffer, query.Table, query.SortQuery)
	q.WriteLimitOffset(buffer, query.LimitQuery, query.OffsetQuery)
	// This is modified and not a real JOIN.
	q.WriteJoin(buffer, query.Table, query.JoinQuery)

	//q.WriteFetch(buffer, query.JoinQuery)

	if query.LockQuery != "" {
		buffer.WriteByte(' ')
		buffer.WriteString(string(query.LockQuery))
	}
}

// WriteFrom SQL to buffer.
func (q Query) WriteFrom(buffer *builder.Buffer, table string) {
	buffer.WriteString(" FROM ")
	buffer.WriteTable(table)
}

// WriteJoin SQL to buffer.
func (q Query) WriteJoin(buffer *builder.Buffer, table string, joins []rel.JoinQuery) {
	if len(joins) == 0 {
		return
	}

	// That's all. We just need a field list.
	// No innter, outer, left or right join.
	// The only other technicality we COULD handle is edges.
	// However, edges are basically just fields:
	// - SELECT likes->post FROM user;
	// "likes" is a n:m (pivot) table, but SurrealDB sees it
	// as a field, returning { "->likes": { "->post": ... } }
	// So a go struct would have to define:
	// - struct { likes struct { post []Post } }
	// ...while setting "column name" to be `->likes` and
	// `->post` respectively.
	// So, while this is technically a join, it should rather
	// be part of the SELECT statement, and be handled properly.
	// FETCH is much more indiect and just specifies which fields
	// should be returned as raw data rather than just the ID.
	buffer.WriteString(" FETCH ")
	for i, join := range joins {
		// TODO: .To or .From? 50/50!
		buffer.WriteString(join.To)
		if i > 0 {
			buffer.WriteString(", ")
		}
	}
}

// WriteWhere SQL to buffer.
func (q Query) WriteWhere(buffer *builder.Buffer, table string, filter rel.FilterQuery) {
	if filter.None() {
		return
	}

	buffer.WriteString(" WHERE ")
	q.Filter.Write(buffer, table, filter, q)
}

// WriteGroupBy SQL to buffer.
func (q Query) WriteGroupBy(buffer *builder.Buffer, table string, fields []string) {
	buffer.WriteString(" GROUP BY ")

	l := len(fields) - 1
	for i, f := range fields {
		buffer.WriteField(table, f)

		if i < l {
			buffer.WriteByte(',')
		}
	}
}

// WriteHaving SQL to buffer.
func (q Query) WriteHaving(buffer *builder.Buffer, table string, filter rel.FilterQuery) {
	if filter.None() {
		return
	}

	buffer.WriteString(" HAVING ")
	q.Filter.Write(buffer, table, filter, q)
}

// WriteOrderBy SQL to buffer.
func (q Query) WriteOrderBy(buffer *builder.Buffer, table string, orders []rel.SortQuery) {
	length := len(orders)

	if length == 0 {
		return
	}

	buffer.WriteString(" ORDER BY ")
	for i, order := range orders {
		if i > 0 {
			buffer.WriteString(", ")
		}

		buffer.WriteField(table, order.Field)

		// TODO: Implement COLLATE, NUMERIC and RAND()
		if order.Asc() {
			buffer.WriteString(" ASC")
		} else {
			buffer.WriteString(" DESC")
		}
	}
}

// WriteLimitOffset SQL to buffer.
func (q Query) WriteLimitOffset(buffer *builder.Buffer, limit rel.Limit, offset rel.Offset) {
	if limit > 0 {
		buffer.WriteString(" LIMIT ")
		buffer.WriteString(strconv.Itoa(int(limit)))

		if offset > 0 {
			buffer.WriteString(" START AT ")
			buffer.WriteString(strconv.Itoa(int(offset)))
		}
	}
}
