package surrealdbdriver

import "context"

// implements driver.Connector
type SurrealConnector struct {
	Creds *CredentialConfig
}

func (*SurrealConnector) Connect(ctx context.Context) (SurrealConn, error)
