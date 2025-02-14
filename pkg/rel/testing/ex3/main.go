package main

import (
	"context"
	"fmt"
	"log"
	"time"

	driver "github.com/IngwiePhoenix/surrealdb-driver"
	srel "github.com/IngwiePhoenix/surrealdb-driver/pkg/rel"
)

type Post struct {
	ID        string    `db:"id"`
	Title     string    `db:"title"`
	CreatedAt time.Time `db:"created_at"`
}

func main() {
	// enable logging on the driver
	driver.SurrealDBDriver.SetLogger(log.Default())

	// context
	ctx := context.Background()

	adapter, err := srel.Open("ws://root:root@127.0.0.1:8000/rpc?method=root&db=ex3&ns=ex3")
	if err != nil {
		log.Println("Could not make adapter")
		log.Fatal(err.Error())
	}
	defer adapter.Close()

	repo := srel.NewRepo(adapter) // Using SurrealDB adapter

	// Logging
	repo.Instrumentation(func(ctx context.Context, op, message string, args ...interface{}) func(err error) {
		log.Printf("[LOG] (%s) %s : ", op, message)
		log.Print(args...)
		log.Print("\n")
		return func(err error) {
			if err == nil {
				return
			}
			log.Fatalf("[ERR] failed: %s\n", err.Error())
		}
	})

	// Write schemas if they do not exist
	sql := `
		DEFINE NAMESPACE IF NOT EXISTS ex3;
		USE NS ex3;
		DEFINE DATABASE IF NOT EXISTS ex3;
		USE DB ex3;
		DEFINE TABLE IF NOT EXISTS posts SCHEMAFULL;
		DEFINE FIELD IF NOT EXISTS title ON posts TYPE string;
		DEFINE FIELD IF NOT EXISTS created_at ON posts TYPE datetime;
	`
	affected, lastidx := repo.MustExec(ctx, sql)
	log.Println("Ran query", affected, lastidx)

	// Let's make an empty process and try to insert.
	post_example3 := Post{
		ID:        "example3",
		Title:     "Example 3",
		CreatedAt: time.Now(),
	}

	repo.MustInsert(ctx, &post_example3)
	log.Println("Inserted data")

	var postList []Post
	repo.MustFindAll(ctx, &postList)
	fmt.Println(postList)
	fmt.Println(postList[0].ID)
	fmt.Println(len(postList))
	for _, p := range postList {
		fmt.Println(p.ID)
	}
}
