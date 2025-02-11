package rel

import (
	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

// Index builder.
type Index struct {
	BufferFactory builder.BufferFactory
}

// Build sql query for index.
func (i Index) Build(index rel.Index) string {
	buffer := i.BufferFactory.Create()

	switch index.Op {
	case rel.SchemaCreate:
		buffer.WriteString("DEFINE INDEX ")
		if index.Optional {
			buffer.WriteString(" IF NOT EXISTS ")
		}
		buffer.WriteEscape(index.Name)
		buffer.WriteString(" ON TABLE ")
		buffer.WriteTable(index.Table)
		buffer.WriteString(" ON FIELDS ")
		for i, c := range index.Columns {
			buffer.WriteField(index.Table, c)
			if i > 0 {
				buffer.WriteString(", ")
			}
		}
		if index.Unique {
			buffer.WriteString(" UNIQUE ")
		}
		// Just write the actual stuff verbatim... lazy. I know. So what.
		buffer.WriteString(index.Options)
		buffer.WriteString("; ")
	case rel.SchemaDrop:
		buffer.WriteString("REMOVE INDEX ")
		buffer.WriteEscape(index.Name)
		buffer.WriteString(" ON TABLE ")
		buffer.WriteTable(index.Table)
		buffer.WriteString("; ")
	}

	return buffer.String()
}

//			log.Print("[REL] Adapter does not support filtered/partial indexes")
