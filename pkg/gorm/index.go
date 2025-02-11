package gorm

import "gorm.io/gorm"

var _ gorm.Index = (*SurrealIndex)(nil)

type SurrealIndex struct {
	columns []string
	name    string
	table   string
	unique  bool
}

// Columns implements gorm.Index.
func (s SurrealIndex) Columns() []string {
	return s.columns
}

// Name implements gorm.Index.
func (s SurrealIndex) Name() string {
	return s.name
}

// Option implements gorm.Index.
func (s SurrealIndex) Option() string {
	panic("unimplemented")
}

// PrimaryKey implements gorm.Index.
func (s SurrealIndex) PrimaryKey() (isPrimaryKey bool, ok bool) {
	// Not implemented
	return false, false
}

// Table implements gorm.Index.
func (s SurrealIndex) Table() string {
	return s.table
}

// Unique implements gorm.Index.
func (s SurrealIndex) Unique() (unique bool, ok bool) {
	// Technically this should check if the index actually _can_ be unique or not.
	// ... At least, I think so. o.o"
	return s.unique, s.unique
}
