package rel

import (
	db "database/sql"
	"strings"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/builder"
)

// MySQL adapter.
type SurrealDB struct {
	sql.SQL
}

// Name of database type this adapter implements.
const Name string = "surrealdb"

func New(database *db.DB) rel.Adapter {
	var (
		bufferFactory = builder.BufferFactory{
			InlineValues:        true,
			ArgumentPlaceholder: "$_",
			ArgumentOrdinal:     true,
			BoolTrueValue:       "TRUE",
			BoolFalseValue:      "FALSE",
			Quoter:              Quote{},
			ValueConverter:      ValueConvert{},
		}
		filterBuilder = Filter{}
		queryBuilder  = Query{
			BufferFactory: bufferFactory,
			Filter:        filterBuilder,
		}
		InsertBuilder = Insert{
			BufferFactory:       bufferFactory,
			InsertDefaultValues: true,
		}
		insertAllBuilder = InsertAll{
			BufferFactory: bufferFactory,
		}
		updateBuilder = Update{
			BufferFactory: bufferFactory,
			Query:         queryBuilder,
			Filter:        filterBuilder,
		}
		deleteBuilder = Delete{
			BufferFactory: bufferFactory,
			Query:         queryBuilder,
			Filter:        filterBuilder,
		}
		ddlBufferFactory = builder.BufferFactory{
			InlineValues:   true,
			BoolTrueValue:  "true",
			BoolFalseValue: "false",
			Quoter:         Quote{},
			ValueConverter: ValueConvert{},
		}
		tableBuilder = Table{
			BufferFactory: ddlBufferFactory,
			DefinitionFilter: func(table rel.Table, def rel.TableDefinition) bool {
				if field, ok := def.(rel.Column); ok {
					return strings.ToLower(field.Name) != "id"
				}
				// Skip other stuff.
				return true
			},
		}
		indexBuilder = Index{
			BufferFactory: ddlBufferFactory,
		}
	)

	return &SurrealDB{
		SQL: sql.SQL{
			QueryBuilder:     queryBuilder,
			InsertBuilder:    InsertBuilder,
			InsertAllBuilder: insertAllBuilder,
			UpdateBuilder:    updateBuilder,
			DeleteBuilder:    deleteBuilder,
			TableBuilder:     tableBuilder,
			IndexBuilder:     indexBuilder,
			Increment:        -1, // SurrealDB has no AUTO_INCREMENT
			ErrorMapper:      errorMapper,
			DB:               database,
		},
	}
}

var dbOpen = db.Open

// Open mysql connection using dsn.
func Open(url string) (rel.Adapter, error) {
	database, err := dbOpen(Name, url)
	return New(database), err
}

// MustOpen mysql connection using dsn.
func MustOpen(url string) rel.Adapter {
	adapter, err := Open(url)
	if err != nil {
		panic(err)
	}
	return adapter
}

// Name of database adapter.
func (SurrealDB) Name() string {
	return Name
}

func errorMapper(err error) error {
	// TODO: Really huge todo...
	return err
}
