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
