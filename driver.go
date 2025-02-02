package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"errors"
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
}

func (*SurrealConn) Prepare(query string) (SurrealStmt, error)
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

// implements driver.Stmt
type SurrealStmt struct{}

// implements driver.Tx
//type SurrealTx struct {}

// implements driver.ConnBeginTx
//type SurrealConnBeginTx struct {}
//func(*SurrealConnBeginTx) BeginTx(ctx context.Context, opts sql.TxOptions) (SurrealTx, error)

// implements driver.Rows
type SurrealRows struct{}

func (*SurrealRows) Columns() []string
func (*SurrealRows) Close() error
func (*SurrealRows) Next(dest []driver.Value) error

// Implement
type SurrealScanner struct{}

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
