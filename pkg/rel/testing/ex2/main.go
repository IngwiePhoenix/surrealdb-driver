package main

import (
	"context"
	"fmt"
	"log"
	"time"

	driver "github.com/IngwiePhoenix/surrealdb-driver"
	srel "github.com/IngwiePhoenix/surrealdb-driver/pkg/rel"
	"github.com/go-rel/rel"
)

type Entity struct {
	Abbrev         string `json:"abbrev"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Name           string `json:"-"`
	AddressStreet  string `json:"address_street"`
	AddressNumber  string `json:"address_number"`
	AddressZipCode int    `json:"address_zip_code"`
	AddressCity    string `json:"address_city"`
	AddressCountry string `json:"address_country"`
	AddressExtra   string `json:"address_extra"`
	Phone          string `json:"phone"`
	Mobile         string `json:"mobile"`
	EMail          string `json:"email"`
	IsExternal     bool   `json:"is_external"`
	Iscompany      bool   `json:"is_company"`
	Organization   string `json:"organization"`
	Position       string `json:"position"`
}

type LegalBasis struct {
	Title string `json:"title"`
	Kind  string `json:"kind"`
	// TODO(KI): Figure out how to store documents for real
	Document []byte `json:"document"`
	Notes    string `json:"notes"`
}

type Risk struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Storage struct {
	Duration string       `json:"duration"`
	Reason   []LegalBasis `json:"reason"`
	Location string       `json:"location"`
}

type Task struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type Process struct {
	Title       string    `db:"title"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"` // TODO(KI): Read-Only
	/*
	   Responsible []Entity     `json:"responsible"`
	   LegalBasis  []LegalBasis `json:"legal_basis"`
	   Risks       []Risk       `json:"risks"`
	   Storage     []Storage    `json:"storage"`
	   Tasks       []Task       `json:"tasks"`
	*/
	Responsible  []string `json:"responsible"`
	LegalBasis   []string `json:"legal_basis"`
	Risks        []string `json:"risks"`
	Storage      []string `json:"storage"`
	Tasks        []string `json:"tasks"`
	AffectedData []string `db:"affected_data"`
}

func main() {
	// enable logging on the driver
	driver.SurrealDBDriver.SetLogger(log.Default())

	adapter, err := srel.Open("ws://root:root@127.0.0.1:8000/rpc?method=root&db=dsbt&ns=dsbt")
	if err != nil {
		log.Println("Could not make adapter")
		log.Fatal(err.Error())
	}
	defer adapter.Close()
	//srAdapter := adapter.(*srel.SurrealDB)

	repo := srel.NewRepo(adapter)
	repo.Instrumentation(func(ctx context.Context, op, message string, args ...interface{}) func(err error) {
		log.Printf("[LOG] (%s) %s : %v\n", op, message, args)
		return func(err error) {
			if err == nil {
				return
			}
			log.Fatalf("[ERR] failed: %s\n", err.Error())
		}
	})

	var process []Process
	repo.MustFindAll(
		context.Background(),
		&process,
		rel.From("process"),
	)
	fmt.Println(process)
}
