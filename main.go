package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

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
	dbDSN := viper.GetString("database.dsn")

	// Set up database connection
	db, err := sql.Open("godror", fmt.Sprintf(`user="%s" password="%s" connectString="%s"`, dbUsername, dbPassword, dbDSN))
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
		// Get id from URL parameters and convert to int64
		idStr := ps.ByName("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}
		// Fetch image from database
		var image []byte
		err = db.QueryRow("select photo from TMS_MUNC_PHOTO where nrord=1 and cod = :1", id).Scan(&image)
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

	router.GET("/record/:id", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Define Record type inside the handler
		type Record struct {
			Id       int64  `json:"id"`
			Sex      string `json:"sex"`
			Fullname string `json:"fullname"`
		}

		// Get id from URL parameters and convert to int64
		idStr := ps.ByName("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid ID format", http.StatusBadRequest)
			return
		}

		// Execute query
		var record Record
		err = db.QueryRow(`
      select cod id
           , decode(kadr_pol_r_504,1,'m',2,'f',0) sex
           , familia||nvl2(numele,' ',null)||numele||nvl2(prenumele,' ',null)||prenumele fullname
      from tms_munc
      where cod = :1`, id).Scan(&record.Id, &record.Sex, &record.Fullname)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Stub data
		// record := Record{
		// 	Id:       id,
		// 	Sex:      "m",
		// 	Fullname: "John Doe",
		// }

		// Set response header
		w.Header().Set("Content-Type", "application/json")

		// Encode and write JSON response
		if err := json.NewEncoder(w).Encode(record); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	log.Printf("Listening on :%s ...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
