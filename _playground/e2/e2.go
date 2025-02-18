package main

import (
	"fmt"

	_ "github.com/IngwiePhoenix/surrealdb-driver"
	st "github.com/IngwiePhoenix/surrealdb-driver/surrealtypes"
	"github.com/jmoiron/sqlx"
)

type Book struct {
	ID    string
	Title string
}

type Author struct {
	Id      string
	Name    string
	Likes   st.ArrayOf[string]
	Socials st.Object
	Written st.Records[Book]
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
		//DEFINE FIELD OVERWRITE socials ON authors FLEXIBLE TYPE object;
		DEFINE FIELD OVERWRITE socials ON authors TYPE object;
		DEFINE FIELD OVERWRITE socials.nostr ON authors TYPE string;
		DEFINE FIELD OVERWRITE written ON authors TYPE option<array<record<books>>>;
	`
	db, err := sqlx.Connect("surrealdb", "ws://db:db@localhost:8000/rpc?method=root&db=p2&ns=p2")
	if err != nil {
		panic(err.Error())
	}

	res, err := db.Exec(defSql)
	if err != nil {
		panic(err.Error())
	}
	ins, _ := res.RowsAffected()
	fmt.Println(ins)

	_, _ = db.Exec(`
		CREATE books:eragon CONTENT {
			title: "The Eragon Book"
		}
	`)
	/*
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

	res, err = db.Exec(`
		DELETE authors:chris;
		CREATE authors:chris CONTENT {
			name: "Christopher",
			likes: ["a", "lot", "of", "stuff"],
			socials: {
				nostr: "foo@mynip05.org"
			},
			written: [
				books:eragon
			]
		}
	`)
	if err != nil {
		fmt.Println(err.Error())
	}
	ins, _ = res.RowsAffected()
	rid, _ := res.LastInsertId()
	fmt.Println(ins, rid)
	if err != nil {
		fmt.Println(err.Error())
	}

	rows, err := db.Queryx("SELECT * FROM authors FETCH written;")
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
		a := Author{}

		fmt.Println("----------")
		fmt.Println(cols)
		err = rows.StructScan(&a)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(a)
		fmt.Println(a.Written.Get()[0].Get().Title)
		fmt.Println("----------")
	}

	fmt.Println("end of program")
}
