package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
)

// implements driver.Driver
type SurrealDriver struct{}

func (*SurrealDriver) OpenConnector(address string) (*SurrealConnector, error) {
	config, err := ParseUrl(address)
	if err != nil {
		return nil, err
	}
	return &SurrealConnector{
		Creds: config,
	}, nil
}

// implements driver.DriverContext
type SurrealDriverContext struct{}

func (*SurrealDriverContext) OpenConnector(name string) (SurrealConnector, error)

// implements driver.Connector
type SurrealConnector struct {
	Creds *CredentialConfig
}

func (*SurrealConnector) Connect(ctx context.Context) (SurrealConn, error)

// implements driver.Conn
type SurrealConn struct {
	HTTPClient *http.Client
	WSClient   *websocket.Conn
	Driver     *SurrealDriver
	Logger     *slog.Logger
}

func (*SurrealConn) Prepare(query string) (SurrealStmt, error)
func (con *SurrealConn) Close() error {
	con.wsClient.Close()
}
func (*SurrealConn) Begin() (driver.Tx, error)
func (con *SurrealConn) IsValid() bool {
	if err := con.WSClient.WriteMessage(websocket.PingMessage, nil); err != nil {
		return false
	}
	return true
}

// implements driver.Validator's method IsValid
func (*SurrealConn) Exec(sql string, values []driver.Value) (driver.Result, error) {
	// Write to WS client
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
