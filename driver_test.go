package surrealdbdriver_test

import (
	"database/sql"
	"log"
	"testing"

	driver "github.com/senpro-it/dsb-tool/extras/surrealdb-driver"
)

func TestDriverCreation(t *testing.T) {
	driver.SurrealDBDriver.SetLogger(log.Default())
	db, err := sql.Open("surrealdb", "ws://db:db@localhost:8000/rpc?method=root&db=dsbt&ns=dsbt")
	if err != nil {
		t.Error(err)
	}

	t.Run("RunInfoAsExec", func(t *testing.T) {
		t.Log("grab info for root")
		res, err := db.Exec("info for root;")
		t.Log("error after statement: ", err)
		if err == nil {
			affected, err := res.RowsAffected()
			t.Log("affected, err: ", affected, err)

			lastIdx, err := res.LastInsertId()
			t.Log("lastIdx, err: ", lastIdx, err)
		} else {
			t.Error("An error occured: ", err)
		}
	})

	t.Run("RunInfoAsQuery", func(t *testing.T) {
		t.Log("grab info for root")
		res, err := db.Query("info for root;")
		t.Log("error after statement: ", err)
		if err == nil {
			cols, err := res.Columns()
			t.Log("Columns: ", cols)
			t.Log("Error: ", err)

			for res.Next() {
				var accesses, namespaces, nodes, users any
				err := res.Scan(&accesses, &namespaces, &nodes, &users)
				t.Log(accesses, namespaces, nodes, users)
				t.Log(err)
				if err != nil {
					break
				}
			}
		} else {
			t.Error("An error occured: ", err)
		}
	})

	t.Run("ReturnBasicValue", func(t *testing.T) {
		res, err := db.Query("RETURN \"foo\";")
		t.Log("error after statement: ", err)
		if err == nil {
			cols, err := res.Columns()
			t.Log("Columns: ", cols)
			t.Log("Error: ", err)

			for res.Next() {
				var foo string
				err := res.Scan(&foo)
				t.Log(foo)
				t.Log(err)
				if err != nil {
					break
				}
			}
		} else {
			t.Error("An error occured: ", err)
		}
	})

	t.Run("ReturnStructuredData", func(t *testing.T) {
		res, err := db.Query("RETURN { \"life\": 42, \"testWords\": [\"foo\", \"bar\", \"baz\"] };")
		t.Log("error after statement: ", err)
		if err == nil {
			cols, err := res.Columns()
			t.Log("Columns: ", cols)
			t.Log("Error: ", err)

			for res.Next() {
				var life string
				var testWords []interface{}
				err := res.Scan(&life, &testWords)
				t.Log("Results from scan:", life, testWords)
				t.Log(err)
				if err != nil {
					break
				}
			}
		} else {
			t.Error("An error occured: ", err)
		}
	})

	t.Run("ThrowOnPurpose", func(t *testing.T) {
		_, err := db.Query("THROW \"aqua\"")
		t.Log("error after statement: ", err)
		if err.Error() != "ERR: An error occurred: aqua" {
			t.Fatal("Wrong error message:", err)
		}
	})
}
