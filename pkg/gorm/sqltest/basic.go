package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type User struct {
	ID   uint
	Name string
}

func main() {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	sql := db.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&User{}).Where("name = ?", "Alice").Find(&User{})
	})

	fmt.Println(sql) // Output: SELECT * FROM `users` WHERE name = 'Alice'
}
