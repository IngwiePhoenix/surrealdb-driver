package surrealdbdriver

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/IngwiePhoenix/surrealdb-driver/config"
	"github.com/clok/kemba"
	"github.com/gorilla/websocket"
)

// implements driver.Driver
type SurrealDriver struct {
	k *kemba.Kemba
	e *Debugger
}

var _ driver.Driver = (*SurrealDriver)(nil)
var _ driver.DriverContext = (*SurrealDriver)(nil)

//var _ sql.DB = (*SurrealDriver)(nil)

func (d *SurrealDriver) Open(address string) (driver.Conn, error) {
	k := d.k.Extend("Open")
	k.Println("address", address)
	config, err := config.ParseUrl(address)
	if err != nil {
		return nil, err
	}
	k.Println("config", config)
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
func (d *SurrealDriver) OpenConnector(address string) (driver.Connector, error) {
	k := d.k.Extend("OpenConnector")
	k.Println("address", address)
	config, err := config.ParseUrl(address)
	if err != nil {
		return nil, err
	}
	k.Println("config", config)
	newk := localKemba.Extend("connector")
	return &SurrealConnector{
		Creds:  config,
		Dialer: &websocket.Dialer{},
		driver: d,
		k:      newk,
		e:      makeErrorLogger(newk),
	}, nil
}

var SurrealDBDriver *SurrealDriver
var localKemba = kemba.New("surrealdb:driver")

func init() {
	k := localKemba
	SurrealDBDriver = &SurrealDriver{k: k, e: makeErrorLogger(k)}
	sql.Register("surrealdb", SurrealDBDriver)
	k.Log("Driver is now added.")
}
