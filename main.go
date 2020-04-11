package main

import (
	"database/sql"
	"log"
	"os"
	"net/http"
	"fmt"

	_ "github.com/lib/pq"
)

func db_init(db *sql.DB) {
	if err := db.Ping(); err != nil {
		log.Fatal("Can't connect tot database %q", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS cities (
							city_id integer PRIMARY KEY,
							city_name VARCHAR (50) NOT NULL,
							country VARCHAR (50) NOT NULL,
							weather VARCHAR (1000),
							last_update TIMESTAMP
						)`); err != nil {
		log.Fatal("Error creating database table: %q", err)
	}
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Done")
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5050"
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Error opening database: %q", err)
	}
	db_init(db)

	http.HandleFunc("/", mainHandler)
	http.ListenAndServe(":" + port, nil)
}
