package rel

import (
	"log"
	"strconv"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql/builder"
)

type (
	ColumnMapper        func(*rel.Column) (string, int, int)
	ColumnOptionsMapper func(*rel.Column) string
	DropKeyMapper       func(rel.KeyType) string
	DefinitionFilter    func(table rel.Table, def rel.TableDefinition) bool
)

// Table builder.
type Table struct {
	BufferFactory       builder.BufferFactory
	ColumnMapper        ColumnMapper
	ColumnOptionsMapper ColumnOptionsMapper
	DropKeyMapper       DropKeyMapper
	DefinitionFilter    DefinitionFilter
}

// Build SQL query for table creation and modification.
func (t Table) Build(table rel.Table) string {
	buffer := t.BufferFactory.Create()

	switch table.Op {
	case rel.SchemaCreate:
		t.WriteDefineTable(&buffer, table)
	case rel.SchemaAlter:
		t.WriteAlterTable(&buffer, table)
	case rel.SchemaRename:
		t.WriteRenameTable(&buffer, table)
	case rel.SchemaDrop:
		t.WriteDropTable(&buffer, table)
	}

	return buffer.String()
}

// WriteCreateTable query to buffer.
func (t Table) WriteDefineTable(buffer *builder.Buffer, table rel.Table) {
	defs := t.definitions(table)

	// Head
	buffer.WriteString("DEFINE TABLE ")
	if table.Optional {
		buffer.WriteString("IF NOT EXISTS ")
	}
	// TODO: if table.Options == "overwrite"
	buffer.WriteTable(table.Name)
	buffer.WriteString("; ")

	// Body
	if len(defs) > 0 {
		for i, def := range defs {
			switch v := def.(type) {
			case rel.Column:
				// DEFINE FIELD
				t.WriteDefineField(buffer, table.Name, v)
			case rel.Key:
				t.WriteDefineFieldForKey(buffer, table.Name, v)
			case rel.Raw:
				buffer.WriteString(string(v))
				// TODO: rel.Index ?
			}
		}
	}
	//t.WriteOptions(buffer, table.Options)
}

// TODO: SurrealDB has no primary or foreign key - and thus no such defs...
func (t Table) WriteDefineFieldForKey(buffer *builder.Buffer, tname string, key rel.Key) {
	buffer.WriteString("DEFINE INDEX ")
	buffer.WriteString(key.Name)
	buffer.WriteString(" ON TABLE ")
	buffer.WriteString(tname)
	buffer.WriteString(" FIELDS ")
	for i, n := range key.Columns {
		buffer.WriteString(n)
		if i > 0 {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString(key.Options)
	buffer.WriteString("; ")
}

func (t Table) WriteDefineField(buffer *builder.Buffer, tname string, col rel.Column) {
	buffer.WriteString("DEFINE FIELD ")
	buffer.WriteString(col.Name)
	buffer.WriteString(" ON TABLE ")
	buffer.WriteString(tname)
	buffer.WriteString(" TYPE ")
	var typ string
	switch col.Type {
	case rel.Bool:
		// st.Bool
		typ = "bool"
	case rel.SmallInt, rel.Int, rel.BigID:
		// st.Int
		typ = "int"
	case rel.Float:
		// st.Float
		typ = "float"
	case rel.Decimal:
		// st.Decimal
		typ = "decimal"
	case rel.String, rel.Text:
		// st.String
		typ = "string"
	case rel.JSON:
		// st.Object
		typ = "object"
	case rel.Date, rel.DateTime:
		// st.Date
		typ = "date"
	case rel.Time:
		// st.Duration
		typ = "duration"
	default:
		// Uh... is this how we do custom types here?
		typ = string(col.Type)
	}
	if col.Required {
		// Wrap into an optional
		typ = "option<" + typ + ">"
	}
	buffer.WriteString(typ)
	// TODO: REFERENCE [ ON DELETE [REJECT | CASCADE | IGNORE | UNSET | THEN @expr ]]
	if col.Default != nil {
		buffer.WriteString(" DEFAULT ")
		buffer.WritePlaceholder()
		buffer.AddArguments(col.Default)
	}
	// TODO: READONLY
	// TODO: VALUE @expr
	// TODO: ASSERT @expr
	// TODO: PERMISSIONS...
	// TODO: COMMENT @expr
	buffer.WriteString("; ")

	if col.Unique {
		// HACK: Quick and dirty unique index insert
		t.WriteDefineUniqueIndex(buffer, tname, col.name)
	}
}

// WriteAlterTable query to buffer.
func (t Table) WriteAlterTable(buffer *builder.Buffer, table rel.Table) {
	defs := t.definitions(table)

	if table.Op == rel.SchemaAlter && table.Rename != "" {
		// We need to copy a whole table... which is super unoptimal.
		// 0. Pretend this is a wholly new table.
		newTable := deepCopyTable(table)
		newTable.Name = newTable.Rename
		newTable.Rename = ""
		// 1. Use that table to write a whole new definition
		t.WriteDefineTable(buffer, newTable)
		// 2. Copy **everything**
		buffer.WriteString("INSERT ")
		buffer.WriteTable(newTable.Name)
		buffer.WriteString(" CONTENT (SELECT * FROM ")
		buffer.WriteTable(table.Name)
		buffer.WriteString(") PARALLEL; ")
		// 3. Remove old
		buffer.WriteString("REMOVE TABLE ")
		buffer.WriteTable(table.Name)
		buffer.WriteString("; ")

		// from here on out, we have to assume the changes took effect.
		// I am pretty sure I'm shooting myself in the foot.
		table = newTable
	}

	for _, def := range defs {
		switch v := def.(type) {
		case rel.Column:
			switch v.Op {
			case rel.SchemaCreate:
				t.WriteDefineField(buffer, table.Name, v)
			case rel.SchemaRename:
				// three-step approach: Define, copy, delete
				// 1. Define new field
				t.WriteDefineField(buffer, table.Name, v)
				// 2. Copy old to new
				buffer.WriteString("UPDATE ")
				buffer.WriteTable(table.Name)
				buffer.WriteString(" SET ")
				buffer.WriteField(table.Name, v.Rename)
				buffer.WriteString(" = ")
				buffer.WriteField(table.Name, v.Name)
				buffer.WriteString("; ")
				// 3. Delete old
				if v.Unique {
					t.WriteRemoveUniqueIndex(buffer, table.Name, v.Name)
				}
				t.WriteRemoveField(buffer, table.Name, v.Name)
			case rel.SchemaDrop:
				if v.Unique {
					t.WriteRemoveUniqueIndex(buffer, table.Name, v.Name)
				}
				t.WriteRemoveField(buffer, table.Name, v.Name)
			}
		case rel.Key:
			// TODO: We only handle unique keys here.
			/*
				switch v.Op {
				case rel.SchemaCreate:
				case rel.SchemaDrop:
				}
			*/
		case rel.Raw:
			buffer.WriteString(string(v))
		}

		t.WriteOptions(buffer, table.Options)
		buffer.WriteByte(';')
	}
}

// WriteRenameTable query to buffer.
func (t Table) WriteRenameTable(buffer *builder.Buffer, table rel.Table) {
	buffer.WriteString("ALTER TABLE ")
	buffer.WriteTable(table.Name)
	buffer.WriteString(" RENAME TO ")
	buffer.WriteTable(table.Rename)
	buffer.WriteByte(';')
}

// WriteDropTable query to buffer.
func (t Table) WriteDropTable(buffer *builder.Buffer, table rel.Table) {
	buffer.WriteString("DROP TABLE ")

	if table.Optional {
		buffer.WriteString("IF EXISTS ")
	}

	buffer.WriteTable(table.Name)
	buffer.WriteByte(';')
}

// WriteColumn definition to buffer.
func (t Table) WriteColumn(buffer *builder.Buffer, column rel.Column) {
	typ, m, n := t.ColumnMapper(&column)

	buffer.WriteEscape(column.Name)
	buffer.WriteByte(' ')
	buffer.WriteString(typ)

	if m != 0 {
		buffer.WriteByte('(')
		buffer.WriteString(strconv.Itoa(m))

		if n != 0 {
			buffer.WriteByte(',')
			buffer.WriteString(strconv.Itoa(n))
		}

		buffer.WriteByte(')')
	}

	if opts := t.ColumnOptionsMapper(&column); opts != "" {
		buffer.WriteByte(' ')
		buffer.WriteString(opts)
	}

	if column.Default != nil {
		buffer.WriteString(" DEFAULT ")
		buffer.WriteValue(column.Default)
	}

	t.WriteOptions(buffer, column.Options)
}

// WriteKey definition to buffer.
func (t Table) WriteKey(buffer *builder.Buffer, key rel.Key) {
	typ := string(key.Type)

	buffer.WriteString(typ)

	if key.Name != "" {
		buffer.WriteByte(' ')
		buffer.WriteEscape(key.Name)
	}

	buffer.WriteString(" (")
	for i, col := range key.Columns {
		if i > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteEscape(col)
	}
	buffer.WriteString(")")

	if key.Type == rel.ForeignKey {
		buffer.WriteString(" REFERENCES ")
		buffer.WriteTable(key.Reference.Table)

		buffer.WriteString(" (")
		for i, col := range key.Reference.Columns {
			if i > 0 {
				buffer.WriteString(", ")
			}
			buffer.WriteEscape(col)
		}
		buffer.WriteString(")")

		if onDelete := key.Reference.OnDelete; onDelete != "" {
			buffer.WriteString(" ON DELETE ")
			buffer.WriteString(onDelete)
		}

		if onUpdate := key.Reference.OnUpdate; onUpdate != "" {
			buffer.WriteString(" ON UPDATE ")
			buffer.WriteString(onUpdate)
		}
	}

	t.WriteOptions(buffer, key.Options)
}

// WriteOptions sql to buffer.
func (t Table) WriteOptions(buffer *builder.Buffer, options string) {
	if options == "" {
		return
	}

	buffer.WriteByte(' ')
	buffer.WriteString(options)
}

func (t Table) definitions(table rel.Table) []rel.TableDefinition {
	if t.DefinitionFilter == nil {
		return table.Definitions
	}

	result := []rel.TableDefinition{}

	for _, def := range table.Definitions {
		if t.DefinitionFilter(table, def) {
			result = append(result, def)
		} else {
			log.Printf("[REL] An unsupported table definition has been excluded: %T", def)
		}
	}

	return result
}

func (t Table) WriteUniqueKeyName(buffer *builder.Buffer, fieldName string) {
	buffer.WriteEscape(fieldName + "_is_unique")
}

func (t Table) WriteDefineUniqueIndex(buffer *builder.Buffer, tableName string, colName string) {
	// HACK: Quick and dirty unique index insert
	buffer.WriteString("DEFINE INDEX ")
	t.WriteUniqueKeyName(buffer, colName)
	buffer.WriteString(" ON TABLE ")
	buffer.WriteString(tableName)
	buffer.WriteString(" FIELDS ")
	buffer.WriteString(colName)
	buffer.WriteString(" UNIQUE; ")
}

func (t Table) WriteRemoveUniqueIndex(buffer *builder.Buffer, tableName, colName string) {
	buffer.WriteString("REMOVE INDEX ")
	t.WriteUniqueKeyName(buffer, colName)
	buffer.WriteString(" ON TABLE ")
	buffer.WriteTable(tableName)
	buffer.WriteString("; ")
}

func (t Table) WriteRemoveField(buffer *builder.Buffer, tableName, colName string) {
	buffer.WriteString("REMOVE FIELD ")
	buffer.WriteField(table.Name, v.Name)
	buffer.WriteString("ON TABLE ")
	buffer.WriteTable(table.Name)
	buffer.WriteString("; ")
}
