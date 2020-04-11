package main

import (
	"fmt"
	"database/sql"
	"log"
	"os"
	"net/http"
	"html/template"
	"io/ioutil"

	_ "github.com/lib/pq"
)

var db *sql.DB

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

func getWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city_id := r.URL.Query().Get("city_id")
	queryString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?id=%s&units=metric&appid=%s", city_id, os.Getenv("API_KEY"))
	resp, err := http.Get(queryString)
	FatalOnError(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	FatalOnError(err)
	fmt.Fprintf(w, "%s", body)
}

func getCities(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	query := r.URL.Query().Get("q")
	fmt.Println(query)

	queryString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&appid=%s", query, os.Getenv("API_KEY"))
	resp, err := http.Get(queryString)
	FatalOnError(err)

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	FatalOnError(err)
	fmt.Fprintf(w, "%s", body)
}


func FatalOnError(err error) {
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func main() {
	var err error

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
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

	http.HandleFunc("/api/weather", getWeather)
	http.HandleFunc("/api/cities", getCities)

	fmt.Println("Serving on port ", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
