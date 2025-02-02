package surrealdbdriver

import (
	"context"
	"database/sql/driver"
)

// implements driver.Connector
type SurrealConnector struct {
	Creds *CredentialConfig
}

var _ driver.Connector = (*SurrealConnector)(nil)

func (*SurrealConnector) Connect(ctx context.Context) (driver.Conn, error)

func (s *SurrealConnector) Driver() driver.Driver {
	panic("unimplemented")
}
