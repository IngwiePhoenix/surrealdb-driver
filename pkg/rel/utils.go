package rel

import "github.com/go-rel/rel"

func deepCopyTable(table rel.Table) rel.Table {
	// Copy primitive fields
	newTable := rel.Table{
		Name:     table.Name,
		Rename:   table.Rename,
		Optional: table.Optional,
		Options:  table.Options,
	}

	// Deep copy slices
	newTable.Definitions = make([]rel.TableDefinition, len(table.Definitions))
	copy(newTable.Definitions, table.Definitions)

	return newTable
}
