package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"errors"

	"github.com/IngwiePhoenix/surrealdb-driver/api"
	"github.com/IngwiePhoenix/surrealdb-driver/config"
	"github.com/gorilla/websocket"
)

// implements driver.Connector
type SurrealConnector struct {
	Creds  *config.Credentials
	Dialer *websocket.Dialer
	driver *SurrealDriver
}

var _ driver.Connector = (*SurrealConnector)(nil)

func (c *SurrealConnector) Connect(ctx context.Context) (driver.Conn, error) {
	c.driver.LogInfo("Connector:Connect start", c.Creds.GetDBUrl())
	conn, resp, err := c.Dialer.DialContext(ctx, c.Creds.GetDBUrl(), nil)
	if err != nil {
		c.driver.LogInfo("Connector:Connect error", err, resp)
		return nil, err
	}
	c.driver.LogInfo("Connector:Connect http response", resp)
	if resp.StatusCode != 200 && resp.StatusCode != 101 {
		return nil, errors.New("SurrealDB's initial response was not 200/101: " + resp.Status)
	}
	con := &SurrealConn{
		WSClient: conn,
		Driver:   c.driver,
		Caller:   api.MakeCaller(),
		creds:    c.Creds,
	}
	if err = con.performLogin(); err != nil {
		c.driver.LogInfo("Connector:Connect, signin error: ", err)
		return nil, err
	}
	return con, nil
}

func (s *SurrealConnector) Driver() driver.Driver {
	return s.driver
}
