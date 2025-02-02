package gorm

import (
	"database/sql"
	"encoding/json"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var (
	CreateClauses = []string{"INSERT INTO", "VALUES", "ON DUPLICATE KEY UPDATE"}
	QueryClauses  = []string{}
	UpdateClauses = []string{"UPDATE", "SET"}
	DeleteClauses = []string{"DELETE", "WHERE"}
)

type SurrealGormConfig struct {
	AlwaysReturn bool
}

type SurrealDialector struct {
	*SurrealGormConfig
	DriverName string
	Conn       gorm.ConnPool
	Url        string
}

const (
	// ClauseOnConflict for clause.ClauseBuilder ON CONFLICT key
	ClauseOnConflict = "ON CONFLICT"
	// ClauseValues for clause.ClauseBuilder VALUES key
	ClauseValues = "VALUES"
	// ClauseFor for clause.ClauseBuilder FOR key
	ClauseFor = "FOR"
)

var _ (gorm.Dialector) = (*SurrealDialector)(nil)

func (SurrealDialector) Name() string {
	return "surrealdb"
}

func (dialector SurrealDialector) Initialize(db *gorm.DB) error {
	if dialector.DriverName == "" {
		dialector.DriverName = dialector.Name()
	}

	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		connPool, err := sql.Open(dialector.DriverName, dialector.Url)
		db.ConnPool = connPool
		if err != nil {
			return err
		}
	}

	callbackConfig := &callbacks.Config{
		CreateClauses: CreateClauses,
		QueryClauses:  QueryClauses,
		UpdateClauses: UpdateClauses,
		DeleteClauses: DeleteClauses,
	}

	if dialector.AlwaysReturn {
		callbackConfig.CreateClauses = append(callbackConfig.CreateClauses, "RETURN AFTER")
		callbackConfig.UpdateClauses = append(callbackConfig.UpdateClauses, "RETURN AFTER")
		callbackConfig.DeleteClauses = append(callbackConfig.DeleteClauses, "RETURN AFTER")
	}
	callbacks.RegisterDefaultCallbacks(db, callbackConfig)
	for k, v := range dialector.ClauseBuilders() {
		db.ClauseBuilders[k] = v
	}

	return nil
}

func (dialector SurrealDialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	// TODO: SurrealDB has slightly different syntax here and there.
	clauseBuilders := map[string]clause.ClauseBuilder{
		ClauseOnConflict: func(c clause.Clause, builder clause.Builder) {
			onConflict, ok := c.Expression.(clause.OnConflict)
			if !ok {
				c.Build(builder)
				return
			}

			builder.WriteString("ON DUPLICATE KEY UPDATE ")
			if len(onConflict.DoUpdates) == 0 {
				if s := builder.(*gorm.Statement).Schema; s != nil {
					var column clause.Column
					onConflict.DoNothing = false

					if s.PrioritizedPrimaryField != nil {
						column = clause.Column{Name: s.PrioritizedPrimaryField.DBName}
					} else if len(s.DBNames) > 0 {
						column = clause.Column{Name: s.DBNames[0]}
					}

					if column.Name != "" {
						onConflict.DoUpdates = []clause.Assignment{{Column: column, Value: column}}
					}

					builder.(*gorm.Statement).AddClause(onConflict)
				}
			}

			for idx, assignment := range onConflict.DoUpdates {
				if idx > 0 {
					builder.WriteByte(',')
				}

				builder.WriteQuoted(assignment.Column)
				builder.WriteByte('=')
				if column, ok := assignment.Value.(clause.Column); ok && column.Table == "excluded" {
					column.Table = ""
					builder.WriteString("VALUES(")
					builder.WriteQuoted(column)
					builder.WriteByte(')')
				} else {
					builder.AddVar(builder, assignment.Value)
				}
			}
		},
		ClauseValues: func(c clause.Clause, builder clause.Builder) {
			if values, ok := c.Expression.(clause.Values); ok && len(values.Columns) == 0 {
				builder.WriteString("VALUES()")
				return
			}
			c.Build(builder)
		},
	}

	return clauseBuilders
}

func (dialector SurrealDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: dialector,
			},
		},
		Dialector: dialector,
	}
}

func (dialector SurrealDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	// TODO: SurrealDB uses $<name>[: <type>] = <value> notation.
	writer.WriteByte('?')
}

func (dialector SurrealDialector) QuoteTo(writer clause.Writer, str string) {
	// SurrealDB is really easy about it's strings.
	// TODO: Handle inline variable substitution
	// HINT: Use peek-ahead method
	s, _ := json.Marshal(str)
	escaped := strings.ReplaceAll(string(s), "$", "\\$")
	writer.WriteString(escaped)
}

func (dialector SurrealDialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `\`, vars...)
}

func (dialector SurrealDialector) DataTypeOf(field *schema.Field) string {
	// TODO: Separate functions?
	switch field.DataType {
	case schema.Bool:
		return "bool"
	case schema.Int, schema.Uint:
		return "int"
	case schema.Float:
		return "float"
	case schema.String:
		return "string"
	case schema.Time:
		return "datetime"
	case schema.Bytes:
		return "bytes"
	default:
		return dialector.getSchemaCustomType(field)
	}
}

func (dialector SurrealDialector) getSchemaCustomType(field *schema.Field) string {
	// TODO: Record
	sqlType := string(field.DataType)

	/*if field.AutoIncrement {
		sqlType += "record<>"
	}*/

	return sqlType
}
