package main

import (
	"fmt"
	"log"
	"os"

	_ "github.com/gocolly/colly/v2"
	"github.com/joho/godotenv"
	_ "github.com/nats-io/nats.go"
)

// Point d'entrée de l'application
func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Erreur de chargement du fichier .env : %v", err)
	}

	// Get the value of the NATS_SERVER_URL environment variable
	natsURL := os.Getenv("NATS_SERVER_URL")
	NewNatsService(natsURL)

	// Print a message
	fmt.Printf("Hello, world!\n")
}
