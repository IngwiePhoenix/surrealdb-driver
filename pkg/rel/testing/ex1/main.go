package main

import (
	"context"
	"fmt"
	"log"

	_ "github.com/IngwiePhoenix/surrealdb-driver"
	srel "github.com/IngwiePhoenix/surrealdb-driver/pkg/rel"
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

	// It makes sense to cast into the "native" adapter; but also,
	// this unironically makes testing the finalized queries... harder.
	// Gonna see if I can PR against rel.Adapter to include a flat string builder.
	surrealAdapter := adapter.(*srel.SurrealDB)
	b := rel.Select("*").From("$session")
	str, vals := surrealAdapter.QueryBuilder.Build(b)
	fmt.Println(str)
	fmt.Println(vals)
}
