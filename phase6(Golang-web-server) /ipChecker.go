package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	_ "github.com/lib/pq"
)

type Record struct {
	IP      string `json:"ip"`
	Country string `json:"country"`
	Message string `json:"message"`
}

type CountResponse struct {
	Count int `json:"count"`
}

type GeoIPResponse struct {
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Timezone    string  `json:"timezone"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	AS          string  `json:"as"`
	Query       string  `json:"query"`
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	dbURI := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("postgres", dbURI+"?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS iplocation (ip TEXT PRIMARY KEY, location TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/records", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT COUNT(*) FROM iplocation")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var count int
		for rows.Next() {
			err := rows.Scan(&count)
			if err != nil {
				log.Fatal(err)
			}
		}

		response := CountResponse{Count: count}
		json.NewEncoder(w).Encode(response)
	})

	http.HandleFunc("/records/", func(w http.ResponseWriter, r *http.Request) {
		ip := r.URL.Path[len("/records/"):]

		var location string
		err := db.QueryRow("SELECT location FROM iplocation WHERE ip = $1", ip).Scan(&location)
		if err != nil {
			if err == sql.ErrNoRows {
				// Look up the location using a third-party API
				geoIPResponse, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
				if err != nil {
					log.Fatal(err)
				}
				defer geoIPResponse.Body.Close()

				var geoIP GeoIPResponse
				err = json.NewDecoder(geoIPResponse.Body).Decode(&geoIP)
				if err != nil {
					log.Fatal(err)
				}

				location = geoIP.Country

				// Insert the new record into the table
				_, err = db.Exec("INSERT INTO iplocation (ip, location) VALUES ($1, $2)", ip, location)
				if err != nil {
					log.Fatal(err)
				}

				response := Record{IP: ip, Country: location, Message: "This IP address was not found in the database and has been added."}
				json.NewEncoder(w).Encode(response)
				return
			} else {
				log.Fatal(err)
			}
		}

		response := Record{IP: ip, Country: location, Message: "This IP address was found in the database."}
		json.NewEncoder(w).Encode(response)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
