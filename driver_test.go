package surrealdbdriver_test

import (
	"database/sql"
	"encoding/json"
	"testing"
)

func TestDriverCreation(t *testing.T) {
	db, err := sql.Open("surrealdb", "ws://root:root@localhost:8000/rpc?method=root&db=dsbt&ns=dsbt")
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
				var accesses, namespaces, nodes, system, users string
				err := res.Scan(&accesses, &namespaces, &nodes, &system, &users)
				t.Log("data: ", accesses, namespaces, nodes, users)
				t.Log(err)
				if err != nil {
					break
				}
			}
			t.Log("Loop done")
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
				t.Log("returned value is: ", foo)
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
				var testWords_raw []byte
				var testWords []string
				err := res.Scan(&life, &testWords_raw)
				t.Log("error:", err)
				err = json.Unmarshal(testWords_raw, &testWords)
				t.Log("error:", err)
				t.Log("Results from scan:", life, testWords)
				if err != nil {
					break
				}
			}
		} else {
			t.Error("An error occured: ", err)
		}
	})

	t.Run("ThrowOnPurpose", func(t *testing.T) {
		out, err := db.Query("THROW \"aqua\"")
		t.Log("error after statement: ", err)
		t.Logf("%T", out)
		if err.Error() != "An error occurred: aqua" {
			t.Fatal("Wrong error message:", err)
		}
	})

	t.Run("InfoQueryToStruct", func(t *testing.T) {
		row := db.QueryRow("INFO FOR DB;")
		if row.Err() != nil {
			t.Fatal(err.Error())
		} else {
			// Blind shot straight into nowhere!
			var a, b, c, d, e, f, g, h, i string
			err := row.Scan(&a, &b, &c, &d, &e, &f, &g, &h, &i)
			t.Log(a, b, c, d, e, f, g, h)
			t.Log(err)
		}
	})
}
