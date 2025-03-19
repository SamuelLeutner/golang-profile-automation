package main

import (
	"log"

	p "github.com/SamuelLeutner/golang-profile-automation/internal/services"
	"github.com/SamuelLeutner/golang-profile-automation/router"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	go p.RefilTokens()
	r := router.SetupRouter()
	r.Run(":8080")
}
