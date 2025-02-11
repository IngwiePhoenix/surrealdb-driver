package main

import (
	"fmt"

	_ "github.com/IngwiePhoenix/surrealdb-driver"
	surrealdb "github.com/IngwiePhoenix/surrealdb-driver/pkg/gorm"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
}

func main() {
	db, _ := gorm.Open(surrealdb.Open(), &gorm.Config{})

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("name = ?", "Alice").Find(&User{})
	})

	fmt.Println(sql) // Output: SELECT * FROM `users` WHERE name = 'Alice'
}
