package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	apiSecret := os.Getenv("API_SECRET")
	fmt.Println("Api Secret: ", apiSecret)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading dotenv file")
	}

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}

	fmt.Println("Chat service started on port: ", port)

	log.Fatal(http.ListenAndServe(port, r))
}
