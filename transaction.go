package surrealdbdriver

import "database/sql/driver"

// implements driver.ConnBeginTx
type SurrealConnBeginTx struct {
	conn *SurrealConn
}

func (tx *SurrealConnBeginTx) Rollback() error {
	if !tx.conn.IsValid() {
		return driver.ErrBadConn
	}
	_, err := tx.conn.Exec("CANCEL TRANSACTION;", nil)
	return err
}
func (tx *SurrealConnBeginTx) Commit() error {
	if !tx.conn.IsValid() {
		return driver.ErrBadConn
	}
	_, err := tx.conn.Exec("COMMIT TRANSACTION;", nil)
	return err
}
