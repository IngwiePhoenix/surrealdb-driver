package gorm

import (
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/senpro-it/dsb-tool/extras/surrealdb-driver/gorm/clauses"
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
	db.ClauseBuilders["CREATE"] = clauses.Create{}.Build

	return nil
}

func (dialector SurrealDialector) ClauseBuilders() map[string]clause.ClauseBuilder {
	// TODO: SurrealDB has slightly different syntax here and there.
	clauseBuilders := map[string]clause.ClauseBuilder{}

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
