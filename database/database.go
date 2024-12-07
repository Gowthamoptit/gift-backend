package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// Define constants for the database credentials
const (
	DB_Host     = "localhost"
	DB_User     = "root"
	DB_Password = "pass1234"
	DB_Name     = "gifts"
)

func DatabaseConnection() (*sql.DB, error) {
	dbclient := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s", DB_User, DB_Password, DB_Host, DB_Name)
	//dbclient := "root:pass1234@tcp(localhost:3306)/gifts"
	db, err := sql.Open("mysql", dbclient)
	if err != nil {
		log.Fatal(err)

	}

	db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("The database connected successfully...")
	return db, nil
}
