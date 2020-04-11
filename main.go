package main

import (
	"fmt"
	"database/sql"
	"log"
	"os"
	"net/http"
	"html/template"

	_ "github.com/lib/pq"
)

func db_init(db *sql.DB) {
	var err error

	err = db.Ping();
	FatalOnError(err)

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cities (
						city_id integer PRIMARY KEY,
						city_name VARCHAR (50) NOT NULL,
						country VARCHAR (50) NOT NULL,
						weather VARCHAR (1000),
						last_update TIMESTAMP
						)`)
	FatalOnError(err)

	_, err = db.Exec(`INSERT INTO cities VALUES ($1, $2, $3)
						ON CONFLICT (city_id) DO UPDATE
						SET city_name=$2, country=$3`,
						524894, "Moscow", "RU")
	FatalOnError(err)
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "API")
}

func FatalOnError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	FatalOnError(err)

	db_init(db)

	templates := template.Must(template.ParseFiles("templates/index.html"))
	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templates.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	http.HandleFunc("/api", apiHandler)

	fmt.Println("Serving on port ", port)
	http.ListenAndServe(":" + port, nil)
}
