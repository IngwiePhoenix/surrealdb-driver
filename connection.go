package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"log/slog"

	"github.com/gorilla/websocket"
)

// implements driver.Conn
type SurrealConn struct {
	WSClient *websocket.Conn
	Driver   *SurrealDriver
	Logger   *slog.Logger
	Caller   *SurrealCaller
}

var _ driver.Conn = (*SurrealConn)(nil)
var _ driver.ConnBeginTx = (*SurrealConn)(nil)
var _ driver.ConnPrepareContext = (*SurrealConn)(nil)
var _ driver.NamedValueChecker = (*SurrealConn)(nil)

// Execute directly on the underlying WebSockets connection by utilizing the
// raw API objects.
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

// Execute directly on the underlying WebSockets connection
func (con *SurrealConn) execRaw(sql string, args map[string]interface{}) (*SurrealAPIResponse, error) {
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	return res, err
}

func (con *SurrealConn) execWithArgs(sql string, args map[string]interface{}) (driver.Result, error) {
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	return &SurrealResult{
		RawResult: res,
	}, err
}

func (con *SurrealConn) queryWithArgs(sql string, args map[string]interface{}) (driver.Rows, error) {
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	return &SurrealRows{
		conn:      con,
		rawResult: res,
		resultIdx: 0,
	}, err
}

func (con *SurrealConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	// TODO: Take advantage of a given context.
	return &SurrealStmt{
		conn:  con,
		query: query,
	}, nil
}
func (con *SurrealConn) Prepare(query string) (driver.Stmt, error) {
	return con.PrepareContext(context.Background(), query)
}

func (con *SurrealConn) Close() error {
	return con.WSClient.Close()
}

func (con *SurrealConn) Begin() (driver.Tx, error) {
	return con.BeginTx(context.Background(), driver.TxOptions{})
}

func (con *SurrealConn) IsValid() bool {
	if err := con.WSClient.WriteMessage(websocket.PingMessage, nil); err != nil {
		return false
	}
	return true
}

func (con *SurrealConn) ExecContext(ctx context.Context, sql string, args []driver.NamedValue) (driver.Result, error) {
	// TODO: Use the provided context
	mappedValues := map[string]interface{}{}
	for _, v := range args {
		mappedValues[v.Name] = v.Value
	}
	return con.execWithArgs(sql, mappedValues)
}

func (con *SurrealConn) Exec(sql string, values []driver.Value) (driver.Result, error) {
	mappedValues := map[string]interface{}{}
	for key, v := range values {
		mappedValues["_"+string(rune(key))] = v
	}
	return con.execWithArgs(sql, mappedValues)
}

// implements driver.ConnBeginTx
func (con *SurrealConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	// TODO: Figure out how to take the provided ctx into account.
	// TODO: Can we use the TxOptions?
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

func (con *SurrealConn) CheckNamedValue(nv *driver.NamedValue) (err error) {
	nv.Value, err = checkNamedValue(nv.Value)
	return
}
