package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/IngwiePhoenix/surrealdb-driver"
	srel "github.com/IngwiePhoenix/surrealdb-driver/pkg/rel"
	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
	"github.com/go-rel/rel"
)

func main() {
	adapter, err := srel.Open("ws://root:root@127.0.0.1:8000/rpc?method=root")
	if err != nil {
		log.Println("Could not make adapter")
		log.Fatal(err.Error())
	}
	defer adapter.Close()

	repo := rel.New(adapter)
	affected, lastidx := repo.MustExec(context.Background(), "INFO FOR ROOT;")
	log.Println(affected)
	log.Println(lastidx)

	sql := rel.SQL("INFO FOR ROOT;")
	cursor, err := adapter.Query(context.TODO(), rel.Query{
		SQLQuery: sql,
	})
	if err != nil {
		log.Println("Could not query")
		log.Fatal(err.Error())
	}

	fields, err := cursor.Fields()
	if err != nil {
		log.Println("Failed to get fields")
		log.Fatal(err.Error())
	}
	fmt.Println(fields)

	// Attempt to decode into a struct
	var out struct {
		// accesses namespaces nodes system users
		Accesses   st.Object
		Namespaces st.Object
		Nodes      st.Object
		System     st.Object
		Users      st.Object
	}
	sql = rel.SQL("INFO FOR ROOT;")
	err = repo.Find(context.Background(), &out, sql)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println(out)
	fmt.Println(out.System["available_parallelism"])

	// Let's try making stuff.
	var schema rel.Schema
	schema.CreateTableIfNotExists("rel_version_tbl", func(t *rel.Table) {
		t.ID("id")
		t.BigInt("version", rel.Unsigned(true), rel.Unique(true), rel.Required(true))
		t.DateTime("created_at")
		t.DateTime("updated_at")
		t.Column("data", "record<other>")
	})
	tbl := schema.Migrations[0].(rel.Table)
	surrealAdapter := adapter.(*srel.SurrealDB)
	str := surrealAdapter.TableBuilder.Build(tbl)
	fmt.Println(str)

	/*
		m := migration.New(repo)
		m.Register(
			1, // Version
			func(s *rel.Schema) {
				// up
				s.CreateTable()
			},
			func(s *rel.Schema) {
				// down
			},
		)
	*/
}
