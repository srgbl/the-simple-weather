package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"net/http"
	"encoding/json"
	"html/template"
)

// These types follows structure of JSON obtained from OpenWeatherMap API response.
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

// Error handler
func FatalOnError(err error) {
	if err != nil {
		log.Fatal("Error", err)
	}
}

// Main page handler
func getWeather(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Parsing request parameters
	err := r.ParseForm()
	FatalOnError(err)

	// Check weather for "Moscow" for default. Preparing query for API call
	query := "Moscow"
	if _, e := r.Form["city_name"]; e {
		query = r.Form["city_name"][0]
		query = strings.TrimSpace(query)
	}
	//log.Println(query)

	cityInfo := &cityInfo{Cod:404, NameNotFound:query}

	// Getting rid of queries with numbers instead of names. Because API counts them as city ids.
	if _, err := strconv.Atoi(query); err != nil && len(query) > 0 {

		// Calling API
		queryString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&lang=ru&appid=%s", query, os.Getenv("API_KEY"))
		resp, err := http.Get(queryString)
		FatalOnError(err)
		defer resp.Body.Close()

		// Overlaying JSON to object cityInfo
		if resp.StatusCode == http.StatusOK {
			err = json.NewDecoder(resp.Body).Decode(cityInfo)
			FatalOnError(err)
		} else {
			cityInfo.Cod = resp.StatusCode
			log.Printf("API status: %d, query: %s", resp.StatusCode, query)
		}

		// Converting pressure value from hPa to mm Hg
		cityInfo.Params.Pressure = int(float64(cityInfo.Params.Pressure) * 0.75006)
	}

	// Template processing
	templates := template.Must(template.ParseFiles("templates/index.html"))
	if err := templates.ExecuteTemplate(w, "index.html", cityInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {

	// Getting port number from environment. Important for Heroku
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", getWeather)

	log.Println("Serving on port ", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
