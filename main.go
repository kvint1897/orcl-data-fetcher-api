package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/godror/godror"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

func main() {
	// Read configuration from file
	viper.SetConfigFile("config.yml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %s", err)
	}

	// Extract database connection parameters from config
	dbUsername := viper.GetString("database.username")
	dbPassword := viper.GetString("database.password")
	dbHost := viper.GetString("database.host")
	dbPort := viper.GetString("database.port")
	dbName := viper.GetString("database.name")

	// Set up database connection
	db, err := sql.Open("godror", fmt.Sprintf("%s/%s@%s:%s/%s", dbUsername, dbPassword, dbHost, dbPort, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Set up HTTP server
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	router := httprouter.New()

	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprint(w, "pong\n")
	})

	router.GET("/photo/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Fetch image from database
		var image []byte
		id := ps.ByName("id")
		err := db.QueryRow("select photo from TMS_MUNC_PHOTO where nrord=1 and cod = :1", id).Scan(&image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(image)))

		// Write image to response
		_, err = w.Write(image)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Printf("Listening on :%s ...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
