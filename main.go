package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading dotenv file")
	}

	apiSecret := os.Getenv("API_SECRET")

	fmt.Println("Api Secret: ", apiSecret)

}
