package gorm

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	sdbClause "github.com/IngwiePhoenix/surrealdb-driver/pkg/gorm/clauses"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	gormClause "gorm.io/gorm/clause"
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
	DriverName   string
	Url          string
}

type SurrealDialector struct {
	*SurrealGormConfig
	Conn gorm.ConnPool
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

// TODO: Should this match?
//var _ (gorm.ConnPool) = (*driver.SurrealConn)(nil)

func (SurrealDialector) Name() string {
	return "surrealdb"
}

func (dialector SurrealDialector) Initialize(db *gorm.DB) error {
	// Handle picking the proper SQL driver
	if dialector.DriverName == "" {
		dialector.DriverName = dialector.Name()
	}

	// Handle connections
	if dialector.Conn != nil {
		db.ConnPool = dialector.Conn
	} else {
		connPool, err := sql.Open(dialector.DriverName, dialector.Url)
		db.ConnPool = connPool
		if err != nil {
			return err
		}
	}

	// DB configuration
	db.DisableNestedTransaction = true
	db.DisableAutomaticPing = false
	db.SkipDefaultTransaction = false

	// Configure callback stuff
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

	// Overwriting special instructions:

	// TODO: We do not use joins - we use FETCH $field[, $field[, ...]] instead of JOIN
	//db.ClauseBuilders["SELECT"]
	//db.ClauseBuilders["JOIN"]

	// Non-standart SQL instruction to create. The only special one implemented so far
	db.ClauseBuilders["CREATE"] = CallbackToStructClause[
		sdbClause.Create,
		gormClause.Insert,
	]()

	// Delete allows to specify ONLY.
	db.ClauseBuilders["DELETE"] = CallbackToStructClause[
		sdbClause.Delete,
		gormClause.Delete,
	]()

	// SurrealDB can also insert relations here, so we need to customize it.
	db.ClauseBuilders["INSERT"] = CallbackToStructClause[
		sdbClause.Insert,
		gormClause.Insert,
	]()

	return nil
}

func (dialector SurrealDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return SurrealDBMigrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:        db,
				Dialector: dialector,
			},
		},
		SurrealDialector: dialector,
	}
}

func (dialector SurrealDialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	// TODO: SurrealDB uses $<name>[: <type>] = <value> notation.
	// Taken from the MSSql driver: https://github.com/go-gorm/sqlserver/blob/master/sqlserver.go#L147-L150
	writer.WriteString("$_")
	writer.WriteString(strconv.Itoa(len(stmt.Vars)))
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
	// TODO: Actually write a RegExp to s/($_[\d*])/\1/
	return logger.ExplainSQL(sql, nil, `\`, vars...)
}

func (dialector SurrealDialector) DataTypeOf(field *schema.Field) string {
	// TODO: Separate functions?
	// TODO: Unaccounted types:
	//         SurrealDB | Go            | Note
	//       - arrays    | []any         |
	//       - duration  | time.Duration |
	//       - object    | interface{}   |
	//       - literal   | any?          |
	//       - option    | option[T]     |
	//       - range     |               |
	//       - record    | T             | specific type needed
	//                   |               | supports constraint
	//                   |               | i.e.: record<T [| T2 [|...]]>
	//       - set       | make(T, N)    |
	// TODO: decimal or float...? Probably a config value...?
	switch field.DataType {
	case schema.Bool:
		return "bool"
	case schema.Int, schema.Uint:
		return "int"
	case schema.Float:
		return "decimal"
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

func (dialector SurrealDialector) DefaultValueOf(field *schema.Field) clause.Expression {
	return clause.Expr{SQL: "DEFAULT"}
}
