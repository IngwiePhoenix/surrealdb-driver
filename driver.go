package surrealdbdriver

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"log"

	"github.com/gorilla/websocket"
	"github.com/IngwiePhoenix/surrealdb-driver/config"
)

// implements driver.Driver
type SurrealDriver struct {
	logger *log.Logger
}

var _ driver.Driver = (*SurrealDriver)(nil)
var _ driver.DriverContext = (*SurrealDriver)(nil)

//var _ sql.DB = (*SurrealDriver)(nil)

func (d *SurrealDriver) Open(address string) (driver.Conn, error) {
	config, err := config.ParseUrl(address)
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
func (d *SurrealDriver) OpenConnector(address string) (driver.Connector, error) {
	config, err := config.ParseUrl(address)
	if err != nil {
		return nil, err
	}
	return &SurrealConnector{
		Creds:  config,
		Dialer: &websocket.Dialer{},
		driver: d,
	}, nil
}

func (d *SurrealDriver) SetLogger(l *log.Logger) {
	l.Println("Hello!")
	d.logger = l
}
func (d *SurrealDriver) GetLogger() *log.Logger {
	return d.logger
}

func (d *SurrealDriver) LogInfo(arg ...any) {
	if d.logger != nil {
		d.logger.Print(arg...)
	}
}
func (d *SurrealDriver) LogFatal(arg ...any) {
	if d.logger != nil {
		d.logger.Fatal(arg...)
	}
}
func (d *SurrealDriver) LogPanic(arg ...any) {
	if d.logger != nil {
		d.logger.Panic(arg...)
	}
}

var SurrealDBDriver *SurrealDriver

func init() {
	SurrealDBDriver = &SurrealDriver{logger: nil}
	sql.Register("surrealdb", SurrealDBDriver)
}
