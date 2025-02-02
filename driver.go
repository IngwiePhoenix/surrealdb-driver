package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"errors"
	"io"
	"log/slog"

	"github.com/gorilla/websocket"
)

// implements driver.Driver
type SurrealDriver struct{}

func (d *SurrealDriver) Open(address string) (*SurrealConn, error) {
	config, err := ParseUrl(address)
	if err != nil {
		return nil, err
	}
	dialer := websocket.Dialer{}
	conn, resp, err := dialer.Dial(config.GetDBUrl(), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("SurrealDB's initial response was not 200: " + resp.Status)
	}
	return &SurrealConn{
		WSClient: conn,
		Driver:   d,
	}, nil
}

// implements driver.DriverContext
func (*SurrealDriver) OpenConnector(address string) (*SurrealConnector, error) {
	config, err := ParseUrl(address)
	if err != nil {
		return nil, err
	}
	return &SurrealConnector{
		Creds: config,
	}, nil
}

// implements driver.Connector
type SurrealConnector struct {
	Creds *CredentialConfig
}

func (*SurrealConnector) Connect(ctx context.Context) (SurrealConn, error)

// implements driver.Conn
type SurrealConn struct {
	WSClient *websocket.Conn
	Driver   *SurrealDriver
	Logger   *slog.Logger
	Caller   *SurrealCaller
}

func (con *SurrealConn) execRaw(sql string, args map[string]interface{}) (*SurrealAPIResponse, error) {
	req := con.Caller.CallQuery(sql, args)
	if err := con.WSClient.WriteJSON(req); err != nil {
		return nil, err
	}
	res := &SurrealAPIResponse{}
	if err := con.WSClient.ReadJSON(res); err != nil {
		return nil, err
	}
	return res, nil
}
func (con *SurrealConn) execObj(obj *SurrealAPIRequest) (*SurrealAPIResponse, error) {
	if err := con.WSClient.WriteJSON(obj); err != nil {
		return nil, err
	}
	res := &SurrealAPIResponse{}
	if err := con.WSClient.ReadJSON(res); err != nil {
		return nil, err
	}
	return res, nil
}

func (con *SurrealConn) Prepare(query string) (driver.Stmt, error) {
	return &SurrealStmt{
		conn:  con,
		query: query,
	}, nil
}
func (con *SurrealConn) Close() error {
	con.WSClient.Close()
}
func (*SurrealConn) Begin() (driver.Tx, error)
func (con *SurrealConn) IsValid() bool {
	if err := con.WSClient.WriteMessage(websocket.PingMessage, nil); err != nil {
		return false
	}
	return true
}

// implements driver.Validator's method IsValid
func (con *SurrealConn) Exec(sql string, values []driver.Value) (driver.Result, error) {
	caller := MakeCaller()
	mappedValues := map[string]interface{}{}
	for key, v := range values {
		mappedValues[string(key)] = v
	}
	req := caller.CallQuery(sql, mappedValues)
	if err := con.WSClient.WriteJSON(req); err != nil {
		return nil, err
	}
	res := &SurrealAPIResponse{}
	if err := con.WSClient.ReadJSON(res); err != nil {
		return nil, err
	}

	return &SurrealResult{
		RawResult: res,
	}, nil
}

// implements driver.ConnBeginTx
func (con *SurrealConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if !con.IsValid() {
		return nil, driver.ErrBadConn
	}

	// TODO: Use the response to determine if everything is still fine.
	_, err := con.Exec("BEGIN TRANSACTION;", nil)
	if err != nil {
		return nil, err
	}
	return &SurrealConnBeginTx{
		conn: con,
	}, nil
}

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
func (stmt *SurrealStmt) NumInputs() int {
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

}

// implements driver.ConnBeginTx
type SurrealConnBeginTx struct {
	conn *SurrealConn
}

func (tx *SurrealConnBeginTx) Rollback() error {
	if !tx.conn.IsValid() {
		return driver.ErrBadConn
	}
	_, err := tx.conn.Exec("CANCEL TRANSACTION;", nil)
	return err
}
func (tx *SurrealConnBeginTx) Commit() error {
	if !tx.conn.IsValid() {
		return driver.ErrBadConn
	}
	_, err := tx.conn.Exec("COMMIT TRANSACTION;", nil)
	return err
}

//func(*SurrealConnBeginTx) BeginTx(ctx context.Context, opts sql.TxOptions) (SurrealTx, error)

type SurrealResult struct {
	RawResult *SurrealAPIResponse
}

// SurrealDB's "record IDs" are strings, not ints.
// this is likely going to be a problem and a half...
// I WISH there was a way to represent a record ID nummerically.
func (r *SurrealResult) LastInsertId() (int64, error) {
	if r.RawResult.Error != nil {
		return 0, errors.New(r.RawResult.Error.Message)
	}
	return 0, nil
}
func (r *SurrealResult) RowsAffected() (int64, error) {
	if value, ok := r.RawResult.Result.([]interface{}); ok {
		return int64(len(value)), nil
	}
	if _, ok := r.RawResult.Result.(interface{}); ok {
		return 1, nil
	}
	if r.RawResult.Error != nil {
		return 0, errors.New(r.RawResult.Error.Message)
	}
	return 0, errors.New("RowsAffected() fell through")
}

// implements driver.Rows
type SurrealRows struct{
	conn *SurrealConn
	rawResult *SurrealAPIResponse
	resultIdx int
}

func (rows *SurrealRows) Columns() (cols []string) {
	if value, ok := rows.rawResult.Result.(map[string]interface{}); ok {
		// Response contains key-value pairs
		for k, _ := range value {
			cols = append(cols, k)
		}
		return cols
	}
	if value, ok := rows.rawResult.Result.([]map[string]interface{}); ok {
		// Response contains an array of k-v pairs
		seen := map[string]bool{}
		for _, v := range value {
			for k, _ := range v {
				if !seen[k] { // avoid dupes
					seen[k] = true
					cols = append(cols, k)
				} 
			}
		}
		return cols
	}
	if value, ok := rows.rawResult.Result.(string); ok {
		// Single string response
		cols = []string{"value"}
		return cols
	}
	if value, ok := rows.rawResult.Result.([]string); ok {
		// Array-of-string response
		cols = []string{"values"}
		return cols
	}
}
func (rows *SurrealRows) Close() error {
	if !rows.conn.IsValid() {
		return driver.ErrBadConn
	}
	return rows.conn.Close()
}
func (rows *SurrealRows) Next(dest []driver.Value) error {
	// SurrealDB returns all results, at all time, with no paging.
	// That means we have to write the result back one by one.
	// This, however, only works if the result IS an array.
	// If it is not, then we kinda can't index it.
	// So we have to run multiple strats.
	if value, ok := rows.rawResult.Result.(map[string]interface{}); ok {
		// Single k-v response
		if rows.resultIdx == 1 {
			return io.EOF
		}
		for i, v := range value {
			dest[i] = v
		}
		rows.resultIdx = rows.resultIdx + 1
		return nil
	}
	if value, ok := rows.rawResult.Result.([]map[string]interface{}); ok {
		// List of k-v responses
		if res, ok := value[rows.resultIdx]; ok {
			for i, v := range res {
				dest[i] = v
			}
			rows.resultIdx = rows.resultIdx + 1
			return nil
		} else {
			return io.EOF
		}
	}
	if value, ok := rows.rawResult.Result.([]string); ok {
		// Multi-string response. Column is "values", so we just
		// put all of them in there, immediately.
		if rows.resultIdx == 1 {
			return io.EOF
		}
		dest[0] = value
		rows.resultIdx = rows.resultIdx + 1
		return nil
	}
	if value, ok := rows.rawResult.Result.(string); ok {
		// Single string response, column is "value"
		if rows.resultIdx == 1 {
			return io.EOF
		}
		dest[0] = value
		rows.resultIdx = rows.resultIdx + 1
		return nil
	}
}

// Implement
type SurrealScanner struct{}
