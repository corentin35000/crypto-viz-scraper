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

	// Obtenir l'URL du serveur NATS depuis les variables d'environnement
	natsURL := os.Getenv("NATS_SERVER_URL")
	if natsURL == "" {
		log.Fatal("NATS_SERVER_URL n'est pas défini dans les variables d'environnement")
	}

	// Créer une nouvelle instance de NatsService
	natsService, err := NewNatsService(natsURL)
	if err != nil {
		// Log error and exit
		log.Fatalf("Erreur lors de la création de NatsService : %v", err)
	} else {
		// Log success message
		fmt.Println("NatsService créé avec succès")

		// S'abonner à un sujet for testing
		natsService.Subscribe("crypto.news", func(message string) {
			fmt.Println("Message reçu : ", message)
		})
	}
}
