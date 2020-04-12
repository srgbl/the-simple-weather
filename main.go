package main

import (
	"fmt"
	"log"
	"os"
	"net/http"
	"encoding/json"
	"html/template"

	"database/sql"
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

func getWeather2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	city_id := r.URL.Query().Get("city_id")
	queryString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/forecast?id=%s&units=metric&appid=%s", city_id, os.Getenv("API_KEY"))
	resp, err := http.Get(queryString)
	FatalOnError(err)
	defer resp.Body.Close()

//	body, err := ioutil.ReadAll(resp.Body)
//	FatalOnError(err)
//	fmt.Fprintf(w, "%s", body)
}

type cityInfo struct {
	Cod 			int			`json:"cod"`
	Name			string		`json:"name"`
	Weather			[]Weather	`json:"weather"`
	Params			Params		`json:"main"`
	Wind			Wind		`json:"wind"`
	NameNotFound 	string
}

type Weather struct {
	Main			string	`json:"main"`
	Description 	string	`json:"description"`
	Icon			string	`json:"icon"`
}

type Params struct {
	Temperature		float64	`json:"temp"`
	FeelsLike		float64	`json:"feels_like"`
	Humidity		int		`json:"humidity"`
	Pressure		int		`json:"pressure"`
}

type Wind struct {
	Speed			float64	`json:"speed"`
}


func getWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var err error

	err = r.ParseForm()
	FatalOnError(err)

	query := "Moscow"
	if _, e := r.Form["city_name"]; e {
		query = r.Form["city_name"][0]
	}
	log.Println(query)

	queryString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&lang=ru&appid=%s", query, os.Getenv("API_KEY"))
	resp, err := http.Get(queryString)
	FatalOnError(err)
	defer resp.Body.Close()

	cityInfo := &cityInfo{}

	if resp.StatusCode == http.StatusOK {
		/*
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		FatalOnError(err)
		bodyString := string(bodyBytes)
		log.Println(bodyString)
		*/

		err = json.NewDecoder(resp.Body).Decode(cityInfo)
		FatalOnError(err)
	}

	cityInfo.Params.Pressure = int(float64(cityInfo.Params.Pressure) * 0.75006)
	templates := template.Must(template.ParseFiles("templates/index.html"))
	if err := templates.ExecuteTemplate(w, "index.html", cityInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/w", func(w http.ResponseWriter, r *http.Request) {

	})

	http.HandleFunc("/", getWeather)

	fmt.Println("Serving on port ", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
