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

	t.Run("InfoQueryToStruct", func(t *testing.T) {
		rows, err := db.Query("INFO FOR DB;")
		if err != nil {
			t.Fatal(err.Error())
		}
		cols, err := rows.Columns()
		if err != nil {
			t.Fatal(err.Error())
		}
		idx := driver.IndexOfString(cols, "tables")
		if idx == -1 {
			t.Error("could not determine tables column")
			return
		}
		for rows.Next() {
			valuePtrs := make([]interface{}, len(cols))
			values := make([]interface{}, len(cols))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			rows.Scan(valuePtrs...)
			results := map[string]map[string]interface{}{}
			for i, c := range cols {
				results[c] = values[i].(map[string]interface{})
			}
			t.Log("valuePtrs: ", valuePtrs)
			t.Log("values: ", values)
			t.Log("Columns: ", cols)
			t.Log("Table Column index: ", idx)
			t.Log("Tables: ", values[idx])
			t.Logf("results['tables'] = %T", results["tables"])
			t.Log(results["tables"])
		}
		//t.Log("Tables: ", info)
	})
}
