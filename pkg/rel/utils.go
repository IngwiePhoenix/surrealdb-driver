package rel

import "github.com/go-rel/rel"

func deepCopyTable(table rel.Table) rel.Table {
	// Copy primitive fields
	newTable := rel.Table{
		Name:   table.Name,
		Alias:  table.Alias,
		Limit:  table.Limit,
		Offset: table.Offset,
	}

	// Deep copy slices
	newTable.GroupBy = append([]string{}, table.GroupBy...)

	newTable.OrderBy = make([]rel.Sort, len(table.OrderBy))
	copy(newTable.OrderBy, table.OrderBy)

	// Deep copy Filter (if necessary, assuming Query is a struct)
	newTable.Filter = table.Filter // If `Query` contains pointers, you may need a deeper copy here.

	return newTable
}
