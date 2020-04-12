package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"net/http"
	"encoding/json"
	"html/template"
)

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

	err := r.ParseForm()
	FatalOnError(err)

	query := "Moscow"
	if _, e := r.Form["city_name"]; e {
		query = r.Form["city_name"][0]
	}
	log.Println(query)

	cityInfo := &cityInfo{NameNotFound:query}
	if d, err := strconv.Atoi(query); err != nil {
		log.Println(d)
		queryString := fmt.Sprintf("http://api.openweathermap.org/data/2.5/weather?q=%s&units=metric&lang=ru&appid=%s", query, os.Getenv("API_KEY"))
		resp, err := http.Get(queryString)
		FatalOnError(err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			err = json.NewDecoder(resp.Body).Decode(cityInfo)
			FatalOnError(err)
		}
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.Handle("/static/",
		http.StripPrefix("/static/",
			http.FileServer(http.Dir("static"))))

	http.HandleFunc("/", getWeather)

	fmt.Println("Serving on port ", port)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}
