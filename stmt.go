package surrealdbdriver

import "context"

// implements driver.Stmt
type SurrealStmt struct {
	conn  *SurrealConn
	query string
}

func (stmt *SurrealStmt) Close() error {
	if !stmt.conn.IsValid() {
		return driver.ErrBadConn
	}
	return stmt.conn.Close()
}
func (stmt *SurrealStmt) NumInput() int {
	// SurrealDB uses LET $<key> = <value>
	// ... so, we actually, literally, don't know. o.o
	return -1
}
func (stmt *SurrealStmt) Exec(args []driver.Value) (driver.Result, error)
func (stmt *SurrealStmt) Query(args []driver.Value) (driver.Rows, error)

// implements driver.StmtExecContext
func (stmt *SurrealStmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	// TODO: Apply context to stmt.conn.WSClient
	for _, arg := range args {
		res, err := stmt.conn.execObj(
			stmt.conn.Caller.CallLet(arg.Name, arg.Value)
		)
	}
	res, err := stmt.conn.execRaw(stmt.query, nil)
	return &SurrealResult{
		RawResult: res,
	}, err
}

// implements driver.StmtQueryContext
func (stmt *SurrealStmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	// TODO: Apply context to stmt.conn.WSClient
	for _, arg := range args {
		res, err := stmt.conn.execObj(
			stmt.conn.Caller.CallLet(arg.Name, arg.Value)
		)
	}
	res, err := stmt.conn.execRaw(stmt.query, nil)
	return &SurrealRows{
		conn: stmt.conn,
		rawResult: res,
		resultIdx: 0,
	}, err
}