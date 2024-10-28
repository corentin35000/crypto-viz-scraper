package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

/**
 * RunScraper initialise les services NATS et Colly, et lance le scraping en continu.
 * Cette fonction est appelée depuis le point d'entrée de l'application.
 * return {void}
 */
func RunScraper() {
	// Charger le fichier .env
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
		log.Fatalf("Erreur lors de la création de NatsService : %v", err)
	}

	// Créer une nouvelle instance de CollyService pour le scraping
	collyService := NewCollyService()

	// Boucle infinie pour effectuer le scraping en continu
	for {
		// Démarrer le scraping
		collyService.ScrapeNews("https://www.coindesk.com/", natsService)

		// Attendre 1 minute avant le prochain cycle de scraping
		time.Sleep(1 * time.Minute)
	}
}
