package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"github.com/gorilla/websocket"
)

// implements driver.Conn
type SurrealConn struct {
	WSClient *websocket.Conn
	Driver   *SurrealDriver
	Logger   *slog.Logger
	Caller   *SurrealCaller
	creds    *CredentialConfig
}

var _ driver.Conn = (*SurrealConn)(nil)
var _ driver.ConnBeginTx = (*SurrealConn)(nil)
var _ driver.ConnPrepareContext = (*SurrealConn)(nil)
var _ driver.NamedValueChecker = (*SurrealConn)(nil)
var _ driver.Pinger = (*SurrealConn)(nil)

// Execute directly on the underlying WebSockets connection by utilizing the
// raw API objects.
func (con *SurrealConn) execObj(obj *SurrealAPIRequest) (*SurrealAPIResponse, error) {
	con.Driver.LogInfo("Conn:execObj start", obj)
	if err := con.WSClient.WriteJSON(obj); err != nil {
		con.Driver.LogInfo("Conn:execObj, writeJSON error", err)
		return nil, err
	}
	res := &SurrealAPIResponse{}
	if _, msg, err := con.WSClient.ReadMessage(); err != nil {
		con.Driver.LogInfo("Conn:execObj, ReadMessage: ", err)
		return nil, err
	} else {
		err := json.Unmarshal(msg, res)
		con.Driver.LogInfo("Conn:execObj, ReadMessage, json.Unmarshal: ", err, string(msg))
		if err != nil {
			return nil, err
		}
		if res.Error.Code != 0 {
			return nil, errors.New(
				strconv.FormatInt(res.Error.Code, 10) +
					": " +
					res.Error.Message,
			)
		}
	}
	con.Driver.LogInfo("Conn:execObj end: ", res)
	return res, nil
}

// Execute directly on the underlying WebSockets connection
func (con *SurrealConn) execRaw(sql string, args map[string]interface{}) (*SurrealAPIResponse, error) {
	con.Driver.LogInfo("Conn:execRaw start", sql, args)
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	return res, err
}

func (con *SurrealConn) execWithArgs(sql string, args map[string]interface{}) (driver.Result, error) {
	con.Driver.LogInfo("Conn:execWithArgs start", sql, args)
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	return &SurrealResult{
		RawResult: res,
	}, err
}

func (con *SurrealConn) queryWithArgs(sql string, args map[string]interface{}) (driver.Rows, error) {
	con.Driver.LogInfo("Conn:queryWithArgs start", sql, args)
	res, err := con.execObj(con.Caller.CallQuery(sql, args))
	return &SurrealRows{
		conn:      con,
		rawResult: res,
		resultIdx: 0,
	}, err
}

func (con *SurrealConn) performLogin() error {
	msg, err := con.Caller.CallSignin(con.creds)
	if err != nil {
		con.Driver.LogInfo("Conn:performLogin message error: ", err)
	}
	rawMsg, _ := json.Marshal(msg)
	con.Driver.LogInfo("Conn:performLogin message: ", string(rawMsg))
	res, err := con.execObj(msg)
	if err != nil {
		con.Driver.LogInfo("Conn:performLogin error: ", err)
		return err
	}
	con.Driver.LogInfo("Conn:performLogin success: ", res)
	return err
}

func (con *SurrealConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	con.Driver.LogInfo("Conn:PrepareContext start", ctx, query)
	// TODO: Take advantage of a given context.
	return &SurrealStmt{
		conn:  con,
		query: query,
	}, nil
}
func (con *SurrealConn) Prepare(query string) (driver.Stmt, error) {
	con.Driver.LogInfo("Conn:Prepare start", query)
	return con.PrepareContext(context.Background(), query)
}

func (con *SurrealConn) Close() error {
	con.Driver.LogInfo("Conn:Close")
	return con.WSClient.Close()
}

func (con *SurrealConn) Begin() (driver.Tx, error) {
	con.Driver.LogInfo("Conn:Begin start")
	return con.BeginTx(context.Background(), driver.TxOptions{})
}

func (con *SurrealConn) IsValid() bool {
	con.Driver.LogInfo("Conn:IsValid")
	if err := con.WSClient.WriteMessage(websocket.PingMessage, nil); err != nil {
		con.Driver.LogInfo("Conn:IsValid PingMessage failed: ", err)
		return false
	}
	con.Driver.LogInfo("Conn:IsValid is valid")
	return true
}

func (con *SurrealConn) ExecContext(ctx context.Context, sql string, args []driver.NamedValue) (driver.Result, error) {
	con.Driver.LogInfo("Conn:ExecContext start", ctx, sql, args)
	// TODO: Use the provided context
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		mappedValues := map[string]interface{}{}
		for _, v := range args {
			mappedValues[v.Name] = v.Value
		}
		return con.execWithArgs(sql, mappedValues)
	}
}

func (con *SurrealConn) Exec(sql string, values []driver.Value) (driver.Result, error) {
	con.Driver.LogInfo("Conn:Exec start")
	mappedValues := map[string]interface{}{}
	for key, v := range values {
		mappedValues["_"+string(rune(key))] = v
	}
	return con.execWithArgs(sql, mappedValues)
}

// implements driver.ConnBeginTx
func (con *SurrealConn) BeginTx(ctx context.Context, _ driver.TxOptions) (driver.Tx, error) {
	con.Driver.LogInfo("Conn:BeginTx start", ctx)
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
	con.Driver.LogInfo("Conn:CheckNamedValue")
	nv.Value, err = checkNamedValue(nv.Value)
	return
}

func (con *SurrealConn) Ping(ctx context.Context) error {
	con.Driver.LogInfo("Conn:Ping")
	// TODO: Figure out how to use contexts
	if !con.IsValid() {
		return errors.New("invalid connection")
	}
	msg, err := websocket.NewPreparedMessage(websocket.PingMessage, nil)
	if err != nil {
		return err
	}
	return con.WSClient.WritePreparedMessage(msg)
}
