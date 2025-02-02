package surrealdbdriver

import (
	"context"
	"database/sql/driver"
	"errors"

	"github.com/gorilla/websocket"
)

// implements driver.Connector
type SurrealConnector struct {
	Creds  *CredentialConfig
	Dialer *websocket.Dialer
	driver *SurrealDriver
}

var _ driver.Connector = (*SurrealConnector)(nil)

func (c *SurrealConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, resp, err := c.Dialer.DialContext(ctx, c.Creds.GetDBUrl(), nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.New("SurrealDB's initial response was not 200: " + resp.Status)
	}
	return &SurrealConn{
		WSClient: conn,
		Driver:   c.driver,
	}, nil
}

func (s *SurrealConnector) Driver() driver.Driver {
	return s.driver
}
