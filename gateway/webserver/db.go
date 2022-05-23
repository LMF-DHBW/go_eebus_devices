package webserver

import (
	"database/sql"
	"log"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

func ConnectDb() *sql.DB {
	dbName := "eebus"

	// Capture connection properties.
	cfg := mysql.Config{
		User:                 "pi",
		Passwd:               "",
		Net:                  "tcp",
		Addr:                 "localhost",
		DBName:               "mysql",
		AllowNativePasswords: true,
	}
	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Println("No connection to database")
		return nil
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Println("No connection to database")
		return nil
	}

	_, err = db.Query("CREATE DATABASE IF NOT EXISTS " + dbName)
	if err != nil {
		log.Fatal(err)
	}
	cfg.DBName = dbName
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	createTables()
	return db
}

func createTables() {
	// Create temperature table
	_, err := db.Query(`CREATE TABLE IF NOT EXISTS temperature ( 
			id int(11) NOT NULL AUTO_INCREMENT,
			timestamp DATETIME DEFAULT NOW(),
			value DOUBLE NOT NULL,
			PRIMARY KEY (id)
		   ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Println("Cant create temperature table")
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS solar ( 
		id int(11) NOT NULL AUTO_INCREMENT,
		timestamp DATETIME DEFAULT NOW(),
		value DOUBLE NOT NULL,
		PRIMARY KEY (id)
	   ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Println("Cant create solar table")
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS consumption ( 
		id int(11) NOT NULL AUTO_INCREMENT,
		timestamp DATETIME DEFAULT NOW(),
		value DOUBLE NOT NULL,
		PRIMARY KEY (id)
	   ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Println("Cant create consumption table")
	}

	_, err = db.Query(`CREATE TABLE IF NOT EXISTS battery ( 
		id int(11) NOT NULL AUTO_INCREMENT,
		timestamp DATETIME DEFAULT NOW(),
		value DOUBLE NOT NULL,
		PRIMARY KEY (id)
	   ) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4`)
	if err != nil {
		log.Println("Cant create battery table")
	}
}
