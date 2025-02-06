package gorm

import (
	driver "github.com/IngwiePhoenix/surrealdb-driver"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type SurrealDBMigrator struct {
	migrator.Migrator
}

var _ gorm.Migrator = (*SurrealDBMigrator)(nil)

const SDB_getTables = ""

func (m SurrealDBMigrator) AutoMigrate(dst ...interface{}) error

func (m SurrealDBMigrator) CurrentDatabase() string
func (m SurrealDBMigrator) FullDataTypeOf(*schema.Field) clause.Expr
func (m SurrealDBMigrator) GetTypeAliases(databaseTypeName string) []string

func (m SurrealDBMigrator) CreateTable(dst ...interface{}) error
func (m SurrealDBMigrator) DropTable(dst ...interface{}) error
func (m SurrealDBMigrator) HasTable(dst interface{}) bool
func (m SurrealDBMigrator) RenameTable(oldName, newName interface{}) error
func (m SurrealDBMigrator) GetTables() (tableList []string, err error) {
	info := driver.SurrealInfoResponse{}
	m.DB.Raw("INFO FOR DB;").Scan(&info)
	for tableName, _ := range info.Result.Tables {
		tableList = append(tableList, tableName)
	}
}
func (m SurrealDBMigrator) TableType(dst interface{}) (gorm.TableType, error)

func (m SurrealDBMigrator) AddColumn(dst interface{}, field string) error
func (m SurrealDBMigrator) DropColumn(dst interface{}, field string) error
func (m SurrealDBMigrator) AlterColumn(dst interface{}, field string) error
func (m SurrealDBMigrator) MigrateColumn(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error

// MigrateColumnUnique migrate column's UNIQUE constraint, it's part of MigrateColumn.
func (m SurrealDBMigrator) MigrateColumnUnique(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error
func (m SurrealDBMigrator) HasColumn(dst interface{}, field string) bool
func (m SurrealDBMigrator) RenameColumn(dst interface{}, oldName, field string) error
func (m SurrealDBMigrator) ColumnTypes(dst interface{}) ([]gorm.ColumnType, error)

func (m SurrealDBMigrator) CreateView(name string, option gorm.ViewOption) error
func (m SurrealDBMigrator) DropView(name string) error

func (m SurrealDBMigrator) CreateConstraint(dst interface{}, name string) error
func (m SurrealDBMigrator) DropConstraint(dst interface{}, name string) error
func (m SurrealDBMigrator) HasConstraint(dst interface{}, name string) bool

// Indexes
func (m SurrealDBMigrator) CreateIndex(dst interface{}, name string) error
func (m SurrealDBMigrator) DropIndex(dst interface{}, name string) error
func (m SurrealDBMigrator) HasIndex(dst interface{}, name string) bool
func (m SurrealDBMigrator) RenameIndex(dst interface{}, oldName, newName string) error
func (m SurrealDBMigrator) GetIndexes(dst interface{}) ([]gorm.Index, error)
