package main

import (
	"database/sql"
	"fmt"

	_ "github.com/IngwiePhoenix/surrealdb-driver"
)

type Book struct {
	ID    string
	Title string
}

func main() {
	defSql := `
		DEFINE NAMESPACE IF NOT EXISTS p2;
		USE NS p2;
		DEFINE DATABASE IF NOT EXISTS p2;
		USE DB p2;
		
		DEFINE TABLE IF NOT EXISTS books SCHEMAFULL;
		DEFINE FIELD IF NOT EXISTS title ON books TYPE string;

		DEFINE TABLE IF NOT EXISTS authors SCHEMAFULL;
		DEFINE FIELD IF NOT EXISTS name ON authors TYPE string;
		DEFINE FIELD IF NOT EXISTS likes ON authors TYPE array<string>;
		// DEFINE FIELD IF NOT EXISTS written ON author TYPE array<optional<record<books>>>
	`
	db, err := sql.Open("surrealdb", "ws://db:db@localhost:8000/rpc?method=root&db=p2&ns=p2")
	if err != nil {
		panic(err.Error())
	}

	res, err := db.Exec(defSql)
	if err != nil {
		panic(err.Error())
	}
	ins, _ := res.RowsAffected()
	fmt.Println(ins)

	/*res, err = db.Exec(`
		CREATE books CONTENT {
			title: "The Eragon Book"
		}
	`)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(res)
	*/

	/*rows, err := db.Query("SELECT * FROM books;")
	if err != nil {
		panic(err.Error())
	}
	for rows.Next() {
		if rows.Err() != nil {
			panic(rows.Err())
		}
		cols, err := rows.Columns()
		if err != nil {
			panic(err.Error())
		}
		b := Book{}
		fmt.Println("----------")
		fmt.Println(cols)
		rows.Scan(&b.ID, &b.Title)
		fmt.Println(b)
		fmt.Println("----------")
	}*/

	/*res, err = db.Exec(`
		CREATE authors:chris CONTENT {
			name: "Christopher",
			likes: ["a", "lot", "of", "stuff"]
		}
	`)
	ins, _ = res.RowsAffected()
	fmt.Println(ins)*/
	rows, err := db.Query("SELECT * FROM authors;")
	if err != nil {
		panic(err.Error())
	}
	for rows.Next() {
		if rows.Err() != nil {
			panic(rows.Err())
		}
		cols, err := rows.Columns()
		if err != nil {
			panic(err.Error())
		}
		fmt.Println("----------")
		fmt.Println(cols)
		fmt.Println("----------")
	}
	fmt.Println("end of program")
}
