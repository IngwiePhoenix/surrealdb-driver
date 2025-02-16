package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"errors"
	"net/http"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	"github.com/IngwiePhoenix/surrealdb-driver/config"
	"github.com/clok/kemba"
	"github.com/gorilla/websocket"
)

// implements driver.Connector
type SurrealConnector struct {
	Creds  *config.Credentials
	Dialer *websocket.Dialer
	driver *SurrealDriver
	k      *kemba.Kemba
	e      *Debugger
}

var _ driver.Connector = (*SurrealConnector)(nil)

func (c *SurrealConnector) Connect(ctx context.Context) (driver.Conn, error) {
	k := c.k.Extend("Connect")

	k.Log("start", c.Creds.GetDBUrl())

	headers := http.Header{}
	headers.Add("Content-Type", "application/json")
	headers.Add("Accept", "application/json")
	headers.Add("Sec-WebSocket-Protocol", "json") // why x.x

	conn, resp, err := c.Dialer.DialContext(ctx, c.Creds.GetDBUrl(), headers)
	if c.e.Debug(err) {
		return nil, err
	}
	k.Log("http response", resp)
	if resp.StatusCode != 200 && resp.StatusCode != 101 {
		return nil, errors.New("SurrealDB's initial response was not 200/101: " + resp.Status)
	}
	connk := localKemba.Extend("connection")
	con := &SurrealConn{
		WSClient: conn,
		Driver:   c.driver,
		Caller:   api.MakeCaller(),
		creds:    c.Creds,
		k:        connk,
		e:        makeErrorLogger(connk),
	}

	if err = con.performLogin(); c.e.Debug(err) {
		return nil, err
	}
	return con, nil
}

func (s *SurrealConnector) Driver() driver.Driver {
	return s.driver
}
