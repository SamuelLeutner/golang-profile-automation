package main

import (
	"log"

	"github.com/SamuelLeutner/golang-profile-automation/router"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	r := router.SetupRouter()
	r.Run(":8080")
}
