package utils

import (
	"fmt"
	"log"
	"strings"
	"io/ioutil"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

var dbName = "./ghra.db"

func NewDbConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbName)
	if err != nil {
		return nil, err
	}
	if err = migrateSchema(db); err != nil {
		return nil, err
	}
	return db, nil
	// defer db.Close()
}

func migrateSchema(db *sql.DB) error {
	queries, err := readSchemaFile()
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var querySucceeded, queryFailed int
	log.Println("Schema migration starting ...")
	for _, query := range queries {
		if query = strings.TrimSpace(query); query == "" {
			continue
		}
		_, err := tx.Exec(query)
		if err != nil {
			log.Printf("Failed to execute query: %v\nQuery: %s", err, query)
			queryFailed += 1
			continue
		}
		querySucceeded += 1
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("%d Query(s) run successful\n", querySucceeded)
	log.Printf("%d Query(s) run failed\n", queryFailed)
	log.Println("Schema migration completed!")
	return nil
}

func readSchemaFile() ([]string, error) {
	var queryList []string
	queries, err := ioutil.ReadFile("utils/schema.sql")
	if err != nil {
		return queryList, fmt.Errorf("Failed to load schema file ::" + err.Error())
	}
	queryList = strings.Split(string(queries), ";")
	return queryList, nil
}