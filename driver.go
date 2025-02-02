package surrealdbdriver

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/gorilla/websocket"
)

// implements driver.Driver
type SurrealDriver struct{}

var _ driver.Driver = (*SurrealDriver)(nil)
var _ driver.DriverContext = (*SurrealDriver)(nil)

func (d *SurrealDriver) Open(address string) (driver.Conn, error) {
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
func (*SurrealDriver) OpenConnector(address string) (driver.Connector, error) {
	config, err := ParseUrl(address)
	if err != nil {
		return nil, err
	}
	return &SurrealConnector{
		Creds: config,
	}, nil
}

func init() {
	sql.Register("surrealdb", &SurrealDriver{})
}
