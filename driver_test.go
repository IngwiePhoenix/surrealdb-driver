package surrealdbdriver_test

import (
	"database/sql"
	"log"
	"testing"
)

func TestDriverCreation(t *testing.T) {
	//driver.SurrealDBDriver.SetLogger(log.Default())
	db, err := sql.Open("surrealdb", "ws://db:db@localhost:8000/rpc?method=root&db=dsbt&ns=dsbt")
	if err != nil {
		t.Error(err)
	}

	log.Println("grab info for root")
	res, err := db.Exec("info for root;")
	log.Println("error after statement: ", err)

	if err == nil {
		affected, err := res.RowsAffected()
		log.Println("affected, err: ", affected, err)

		lastIdx, err := res.LastInsertId()
		log.Println("lastIdx, err: ", lastIdx, err)
	} else {
		log.Fatal("An error occured: ", err)
	}
}
