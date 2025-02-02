package surrealdbdriver

import (
	"context"
	"database/sql/driver"
)

// implements driver.Stmt
type SurrealStmt struct {
	conn  *SurrealConn
	query string
}

// Checking interface compatibility per intellisense
var _ driver.Stmt = (*SurrealStmt)(nil)
var _ driver.StmtExecContext = (*SurrealStmt)(nil)
var _ driver.StmtQueryContext = (*SurrealStmt)(nil)
var _ driver.NamedValueChecker = (*SurrealStmt)(nil)
var _ driver.ValueConverter = (*SurrealStmt)(nil)

func (stmt *SurrealStmt) Close() error {
	if !stmt.conn.IsValid() {
		return driver.ErrBadConn
	}
	return stmt.conn.Close()
}

func (stmt *SurrealStmt) NumInput() int {
	stmt.conn.Driver.LogInfo("stmt:NumInput called")
	// SurrealDB uses LET $<key> = <value>
	// ... so, we actually, literally, don't know. o.o
	// Technically we could count the number of $-signs, but that would be misleading,
	// since some of those are reserved.
	// So, it is possible - but, honestly, I can't be arsed to implement it...yet.
	return -1
}

func (stmt *SurrealStmt) Exec(args []driver.Value) (driver.Result, error) {
	stmt.conn.Driver.LogInfo("stmt:Exec called")
	mappedValues := map[string]interface{}{}
	for key, v := range args {
		mappedValues["_"+string(rune(key))] = v
	}
	return stmt.conn.execWithArgs(stmt.query, mappedValues)
}

// implements driver.StmtExecContext
func (stmt *SurrealStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	stmt.conn.Driver.LogInfo("stmt:ExecContext called")
	// TODO: Apply context to stmt.conn.WSClient
	// NOTE: copying the default method here - not sure if values come in once in a while or not.
	mappedValues := map[string]interface{}{}
	for _, v := range args {
		mappedValues[v.Name] = v.Value
	}
	return stmt.conn.execWithArgs(stmt.query, mappedValues)
}

func (stmt *SurrealStmt) Query(args []driver.Value) (driver.Rows, error) {
	stmt.conn.Driver.LogInfo("stmt:Query called")
	mappedValues := map[string]interface{}{}
	for key, v := range args {
		mappedValues["_"+string(rune(key))] = v
	}
	return stmt.conn.queryWithArgs(stmt.query, mappedValues)
}

// implements driver.StmtQueryContext
func (stmt *SurrealStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	stmt.conn.Driver.LogInfo("stmt:QueryContext called")
	// TODO: Apply context to stmt.conn.WSClient
	// NOTE: copying the default method here - not sure if values come in once in a while or not.
	mappedValues := map[string]interface{}{}
	for key, v := range args {
		mappedValues["_"+string(rune(key))] = v
	}
	return stmt.conn.queryWithArgs(stmt.query, mappedValues)
}

func (stmt *SurrealStmt) CheckNamedValue(nv *driver.NamedValue) (err error) {
	stmt.conn.Driver.LogInfo("stmt:CheckNamedValue called")
	nv.Value, err = checkNamedValue(nv.Value)
	return
}

func (stmt *SurrealStmt) ConvertValue(v any) (driver.Value, error) {
	return checkNamedValue(v)
}
