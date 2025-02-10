package gorm

import (
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"

	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type SurrealDBMigrator struct {
	// Obtain original type to overload them but still be compatible
	migrator.Migrator

	// Use our own Dialector instead
	SurrealDialector
}

var _ gorm.Migrator = (*SurrealDBMigrator)(nil)

// SKIP: Use GORM's version
//func (m SurrealDBMigrator) AutoMigrate(dst ...interface{}) error

func (m SurrealDBMigrator) CurrentDatabase() string {
	var out string
	m.DB.Raw("SELECT db FROM $session;").Scan(&out)
	return out
}

// SKIP: It actually queries the dialector if it doesn't know.
//func (m SurrealDBMigrator) FullDataTypeOf(*schema.Field) clause.Expr

func (m SurrealDBMigrator) GetTypeAliases(databaseTypeName string) []string {
	// TODO: There aren't really type aliases in SurrealDB.
	return nil
}

// NOTE: This is a modified version from GORM's internal Migrator.
func (m SurrealDBMigrator) CreateTable(values ...interface{}) error {
	/*
		TODOs:
			- Should add `REFERENCE ON [REJECT|CASCADE|IGNORE|UNSET|THEN $sql]`
			- How to handle "DEFAULT $expr"?
			- Should somehow trigger `READONLY`
			- Handle `VALUE $expr`
			- Handle `DEFAULT $expr`
			- Handle `PERMISSIONS [NONE|FULL|FOR [select|create|update]...]`
	*/

	// Templates
	tableSql := "DEFINE TABLE ? SCHEMAFULL;"
	fieldSql := "DEFINE FIELD ? ON TABLE ? TYPE ?"   // DEFAULT ? READONLY VALUE ? ASSERT ?
	indexSql := "DEFINE INDEX ? ON TABLE ? FIELDS ?" // FIELDS|COLUMNS

	for _, value := range m.ReorderModels(values, false) {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) (err error) {

			if stmt.Schema == nil {
				return errors.New("failed to get schema")
			}

			var (
				createTableSQL = tableSql
				values         = []interface{}{m.CurrentTable(stmt)}
				//hasPrimaryKeyInDataType bool
			)

			for _, dbName := range stmt.Schema.DBNames {
				field := stmt.Schema.FieldsByDBName[dbName]
				if !field.IgnoreMigration {
					createTableSQL += fieldSql
					values = append(values,
						clause.Column{Name: dbName},
						m.CurrentTable(stmt),
						m.DB.Migrator().FullDataTypeOf(field),
					)
					if field.HasDefaultValue {
						createTableSQL += " DEFAULT ?"
						values = append(values, field.DefaultValue)
					}
					// TODO: Turn constraints to ASSERTs?...
					createTableSQL += ";\n"
				}
			}

			// UNUSED: SurrealDB has ID as it's always present and only PK
			/*
				if !hasPrimaryKeyInDataType && len(stmt.Schema.PrimaryFields) > 0 {
					createTableSQL += "PRIMARY KEY ?,"
					primaryKeys := make([]interface{}, 0, len(stmt.Schema.PrimaryFields))
					for _, field := range stmt.Schema.PrimaryFields {
						primaryKeys = append(primaryKeys, clause.Column{Name: field.DBName})
					}

					values = append(values, primaryKeys)
				}
			*/

			for _, idx := range stmt.Schema.ParseIndexes() {
				if m.CreateIndexAfterCreateTable {
					defer func(value interface{}, name string) {
						if err == nil {
							err = tx.Migrator().CreateIndex(value, name)
						}
					}(value, idx.Name)
				} else {
					createTableSQL += indexSql
					values = append(values,
						idx.Name,
						m.CurrentTable(stmt),
					)
					for i, field := range idx.Fields {
						if i > 0 {
							createTableSQL += ", "
						}
						createTableSQL += "?"
						values = append(values, field.DBName)
					}

					switch idx.Class {
					case "FULLTEXT":
						//createTableSQL += " SEARCH ANALYZER ascii BM25"
						// TODO: should warn instead
						return errors.New("'FULLTEXT' search not implemented for SurrealDB yet")
					case "SPATIAL":
						return errors.New("'SPATIAL' not supported by SurrealDB (as far as I know)")
					case "UNIQUE":
						createTableSQL += " UNIQUE"
					default:
						// Non-unqiue indexing is just... indexing.
						createTableSQL += " "
					}

					if idx.Comment != "" {
						createTableSQL += fmt.Sprintf(" COMMENT \"%s\"", idx.Comment)
					}

					// TODO: Could be used for fulltext stuff
					/*
						if idx.Option != "" {
							createTableSQL += " " + idx.Option
						}
					*/

					createTableSQL += ";\n"
					// TODO: Inspect IndexOptions more
					//values = append(values, clause.Column{Name: idx.Name}, tx.Migrator().(BuildIndexOptionsInterface).BuildIndexOptions(idx.Fields, stmt))
				}
			}

			// SurrealDB's relationships work wholly differently.
			// They are their own data type (record<T>) and do not need constraining.
			// Instead, they are defined on the table as such a type, and that's that.
			/*
				if !m.DB.DisableForeignKeyConstraintWhenMigrating && !m.DB.IgnoreRelationshipsWhenMigrating {
					for _, rel := range stmt.Schema.Relationships.Relations {
						if rel.Field.IgnoreMigration {
							continue
						}
						if constraint := rel.ParseConstraint(); constraint != nil {
							if constraint.Schema == stmt.Schema {
								sql, vars := constraint.Build()
								createTableSQL += sql + ","
								values = append(values, vars...)
							}
						}
					}
				}
			*/

			// We handled UNIQUE above. Plus, there is no such thing as CONSTRAINT.
			/*
				for _, uni := range stmt.Schema.ParseUniqueConstraints() {
					createTableSQL += "CONSTRAINT ? UNIQUE (?),"
					values = append(values, clause.Column{Name: uni.Name}, clause.Expr{SQL: stmt.Quote(uni.Field.DBName)})
				}
			*/

			// No CONSTRAINT ... CHECK support
			/*
				for _, chk := range stmt.Schema.ParseCheckConstraints() {
					createTableSQL += "CONSTRAINT ? CHECK (?),"
					values = append(values, clause.Column{Name: chk.Name}, clause.Expr{SQL: chk.Constraint})
				}
			*/

			// Those should, if even applicable, be handled above.
			// Ignore them for now, move later, and then adjust, if needed.
			/*
				if tableOption, ok := m.DB.Get("gorm:table_options"); ok {
					createTableSQL += fmt.Sprint(tableOption)
				}
			*/

			err = tx.Exec(createTableSQL, values...).Error
			return err
		}); err != nil {
			return err
		}
	}
	return nil
}

// Copied from base method but modified with SurrealDB syntax.
func (m SurrealDBMigrator) DropTable(values ...interface{}) error {
	values = m.ReorderModels(values, false)
	for i := len(values) - 1; i >= 0; i-- {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			return tx.Exec("REMOVE TABLE IF EXISTS ?", m.CurrentTable(stmt)).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m SurrealDBMigrator) HasTable(value interface{}) bool {
	var found bool = false

	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		myTables, err := m.GetTables()
		if err != nil {
			return err
		}
		for _, t := range myTables {
			if t == stmt.Table {
				found = true
				return nil
			}
		}
		// If the loop falls through, the default is false.
		return nil
	})

	return found
}

func (m SurrealDBMigrator) RenameTable(oldName, newName interface{}) error {
	// SurrealDB has no rename method...which means, we have to copy-paste everything.
	// Would've thought this to be part of ALTER TABLE, but, nope.

	extractName := func(input interface{}) (clause.Table, error) {
		if v, ok := input.(string); ok {
			return clause.Table{Name: v}, nil
		} else {
			stmt := &gorm.Statement{DB: m.DB}
			if err := stmt.Parse(input); err == nil {
				return m.CurrentTable(stmt).(clause.Table), nil
			} else {
				// Not sure if returning an empty clause.Table is the way to go here?
				return clause.Table{}, err
			}
		}
	}

	var err error
	oldNameC, err := extractName(oldName)
	if err != nil {
		return err
	}
	newNameC, err := extractName(newName)
	if err != nil {
		return err
	}

	// Drop down to raw access
	db, err := m.DB.DB()
	if err != nil {
		return err
	}
	// First, we need to grab all the fields from the old table.
	// For that, we simply assume that we are in the same database.
	rows, err := db.Query("INFO FOR DB;")
	if err != nil {
		return err
	}
	var tables st.Object
	for rows.Next() {
		err = rows.Err()
		if err != nil {
			return err
		}
		rows.Scan(nil, nil, nil, nil, nil, nil, &tables, nil)
	}
	var sql string = "BEGIN TRANSACTION;\n"
	var sqlForOld, sqlForNew string = "", ""
	for tableName, tableDefs := range tables {
		if tableName != oldNameC.Name {
			continue
		}
		tableDefsArr, ok := tableDefs.([]string)
		if !ok {
			return errors.New("can not obtain table definitions")
		}
		sqlForNew += "// Definition from: " + oldNameC.Name + "\n"
		for _, def := range tableDefsArr {
			// We need to swap the old name with the new name
			tmpTblNameOld := " " + tableName + " "
			tmpTblNameNew := " " + newNameC.Name + " "
			newDef := strings.Replace(def, tmpTblNameOld, tmpTblNameNew, 1) + ";\n"
			sqlForNew += newDef + ";\n"
		}
		// We use parallel to hopefuly speed things up a bit.
		// Should probably be a config option...
		sqlForNew += "CREATE " + newNameC.Name + " CONTENT (SELECT * FROM " + oldNameC.Name + ") PARALLEL;\n"
	}
	sql += sqlForOld + "// old <-> new\n" + sqlForNew
	sql += "REMOVE TABLE " + oldNameC.Name + ";\n"
	sql += "COMMIT TRANSACTION;"

	return m.DB.Exec(sql).Error
}

func (m SurrealDBMigrator) GetTables() (tableList []string, err error) {
	// GORM can not handle this kind of return type.
	// So we drop down into raw queries.
	// Technically, this could be handled like a custom SELECT statement,
	// by which we *could* scan into a struct. But, since it can not be customized,
	// it's better to just go low-level for once.
	db, err := m.DB.DB()
	if err != nil {
		return nil, err
	}
	rows, err := db.Query("INFO FOR DB;")
	if err != nil {
		return nil, err
	}
	// Order: accesses, analyzers, configs, functions, models, params, tables, users
	// Index: 0         1          2        3          4       5       6       7
	var tables st.Object
	var tableNames []string
	for rows.Next() {
		err = rows.Err()
		if err != nil {
			return nil, err
		}
		rows.Scan(nil, nil, nil, nil, nil, nil, &tables, nil)
	}
	for table := range tables {
		tableNames = append(tableNames, table)
	}
	return tableNames, nil
}
func (m SurrealDBMigrator) TableType(dst interface{}) (gorm.TableType, error) {
	return nil, errors.New("currently not supported; would only return SCHEMAFUL anyway")
}

// Adopted from the MySQL driver
func (m SurrealDBMigrator) AddColumn(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		// avoid using the same name field
		f := stmt.Schema.LookUpField(name)
		if f == nil {
			return fmt.Errorf("failed to look up field with name: %s", name)
		}

		if !f.IgnoreMigration {
			fieldType := m.FullDataTypeOf(f)
			columnName := clause.Column{Name: f.DBName}
			values := []interface{}{columnName, m.CurrentTable(stmt), fieldType}
			var alterSql strings.Builder
			// TODO: This should also include respective options...
			// IDEA: Could I use clauses to build those? o.o That'd actually be pretty dope.
			alterSql.WriteString("DEFINE FIELD ? ON TABLE ? TYPE ?;")
			return m.DB.Exec(alterSql.String(), values...).Error
		}

		return nil
	})
}

func (m SurrealDBMigrator) DropColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt.Schema != nil {
			if field := stmt.Schema.LookUpField(field); field != nil {
				return m.DB.Exec(
					"REMOVE FIELD ? ON TABLE ?",
					clause.Column{Name: field.DBName}, m.CurrentTable(stmt),
				).Error
			}
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
}
func (m SurrealDBMigrator) AlterColumn(dst interface{}, name string) error {
	return errors.New("currently there is no implementation for SurrealDB to alter columns")
}
func (m SurrealDBMigrator) MigrateColumn(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	return errors.New("currently there is no implementation for SurrealDB to migrate columns")
}

// MigrateColumnUnique migrate column's UNIQUE constraint, it's part of MigrateColumn.
func (m SurrealDBMigrator) MigrateColumnUnique(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	return errors.New("not implemented; because I actually don't know what this is ment to do")
}
func (m SurrealDBMigrator) HasColumn(value interface{}, field string) bool {
	found := false
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if stmt == nil {
			return nil
		}

		db, err := m.DB.DB()
		if err != nil {
			return err
		}
		tbl := m.CurrentTable(stmt).(clause.Table)
		rows, err := db.Query("INFO FOR TABLE " + tbl.Name + ";")
		if err != nil {
			return err
		}
		// Order: events, fields, indexes, lives, tables
		// Index: 0       1       2        3      4
		var fields st.Object
		for rows.Next() {
			err = rows.Err()
			if err != nil {
				return err
			}
			rows.Scan(nil, &fields, nil, nil, nil)
		}
		for dbField := range fields {
			if field == dbField {
				found = true
				return nil
			}
		}
		return nil
	})
	return found && err == nil
}

func (m SurrealDBMigrator) RenameColumn(dst interface{}, oldName, field string) error {
	return errors.New("currently column renaming is not implemented with SurrealDB")
}

func (m SurrealDBMigrator) ColumnTypes(value interface{}) ([]gorm.ColumnType, error) {
	colTypes := make([]gorm.ColumnType, 0)
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		r := regexp.MustCompile(" TYPE ([a-zA-Z0-9_<>]*) ")
		if stmt == nil {
			return nil
		}

		db, err := m.DB.DB()
		if err != nil {
			return err
		}
		tbl := m.CurrentTable(stmt).(clause.Table)
		rows, err := db.Query("INFO FOR TABLE " + tbl.Name + ";")
		if err != nil {
			return err
		}
		// Order: events, fields, indexes, lives, tables
		// Index: 0       1       2        3      4
		var fields st.Object
		for rows.Next() {
			err = rows.Err()
			if err != nil {
				return err
			}
			rows.Scan(nil, &fields, nil, nil, nil)
		}
		for dbField := range fields {
			var fieldData migrator.ColumnType
			dbTyp := r.FindString(dbField)
			fieldData.NameValue = sql.NullString{String: dbTyp, Valid: true}
			// TODO: We can not parse the default value or anything else...
			colTypes = append(colTypes, fieldData)
		}
		return nil
	})
	return colTypes, err
}

func (m SurrealDBMigrator) CreateView(name string, option gorm.ViewOption) error {
	// TODO: Technically supported by DEFINE TABLE ... AS SELECT ...
	return errors.New("currently Views are not implemented with SurrealDB")
}
func (m SurrealDBMigrator) DropView(name string) error {
	// TODO: Should just delete the noted view. Probably by looking for a table ending in _view ?
	return errors.New("currently Views are not implemented with SurrealDB")
}

func (m SurrealDBMigrator) CreateConstraint(dst interface{}, name string) error {
	return errors.New("the SurrealDB engine does not do constraints, only indexes")
}
func (m SurrealDBMigrator) DropConstraint(dst interface{}, name string) error {
	return errors.New("the SurrealDB engine does not do constraints, only indexes")
}
func (m SurrealDBMigrator) HasConstraint(dst interface{}, name string) bool {
	//return errors.New("the SurrealDB engine does not do constraints, only indexes")
	return false
}

// Indexes
func (m SurrealDBMigrator) CreateIndex(value interface{}, name string) error {
	indexSql := "DEFINE INDEX ? ON TABLE ? FIELDS ?" // FIELDS|COLUMNS
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		var err error
		for _, idx := range stmt.Schema.ParseIndexes() {
			createTableSQL := indexSql
			values := []interface{}{}
			values = append(values,
				idx.Name,
				m.CurrentTable(stmt),
			)
			for i, field := range idx.Fields {
				if i > 0 {
					createTableSQL += ", "
				}
				createTableSQL += "?"
				values = append(values, field.DBName)
			}

			switch idx.Class {
			case "FULLTEXT":
				//createTableSQL += " SEARCH ANALYZER ascii BM25"
				// TODO: should warn instead
				return errors.New("'FULLTEXT' search not implemented for SurrealDB yet")
			case "SPATIAL":
				return errors.New("'SPATIAL' not supported by SurrealDB (as far as I know)")
			case "UNIQUE":
				createTableSQL += " UNIQUE"
			default:
				// Non-unqiue indexing is just... indexing.
				createTableSQL += " "
			}

			if idx.Comment != "" {
				createTableSQL += fmt.Sprintf(" COMMENT \"%s\"", idx.Comment)
			}

			// TODO: Could be used for fulltext stuff
			/*
				if idx.Option != "" {
					createTableSQL += " " + idx.Option
				}
			*/

			createTableSQL += ";\n"
			err = m.DB.Exec(createTableSQL, values...).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func (m SurrealDBMigrator) DropIndex(value interface{}, name string) error {
	removeSql := "REMOVE INDEX ? ON TABLE ?;"
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		tbl := m.CurrentTable(stmt).(clause.Table)
		return m.DB.Exec(removeSql, name, tbl.Name).Error
	})
}
func (m SurrealDBMigrator) HasIndex(value interface{}, name string) bool {
	found := false
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		db, err := m.DB.DB()
		if err != nil {
			return err
		}
		tbl := m.CurrentTable(stmt).(clause.Table)
		rows, err := db.Query("INFO FOR TABLE " + tbl.Name + ";")
		if err != nil {
			return err
		}
		// Order: events, fields, indexes, lives, tables
		// Index: 0       1       2        3      4
		var indexes st.Object
		for rows.Next() {
			err = rows.Err()
			if err != nil {
				return err
			}
			rows.Scan(nil, nil, &indexes, nil, nil)
		}
		for index := range indexes {
			if index == name {
				found = true
				return nil
			}
		}
		return nil
	})
	return found && err == nil
}
func (m SurrealDBMigrator) RenameIndex(dst interface{}, oldName, newName string) error {
	return errors.New("currently SurrealDB does not support renaming indexes")
}
func (m SurrealDBMigrator) GetIndexes(value interface{}) ([]gorm.Index, error) {
	foundIndexes := make([]gorm.Index, 0)
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		db, err := m.DB.DB()
		if err != nil {
			return err
		}
		tbl := m.CurrentTable(stmt).(clause.Table)
		rows, err := db.Query("INFO FOR TABLE " + tbl.Name + ";")
		if err != nil {
			return err
		}
		// Order: events, fields, indexes, lives, tables
		// Index: 0       1       2        3      4
		var indexes st.Object
		for rows.Next() {
			err = rows.Err()
			if err != nil {
				return err
			}
			rows.Scan(nil, nil, &indexes, nil, nil)
		}
		for indexName := range indexes {
			newIndex := SurrealIndex{
				columns: []string{},
				table:   tbl.Name,
				unique:  false,
				name:    indexName,
			}
			foundIndexes = append(foundIndexes, newIndex)
		}
		return nil
	})
	return foundIndexes, err
}
