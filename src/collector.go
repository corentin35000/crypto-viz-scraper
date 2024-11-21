package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gocolly/colly/v2"
)

/**
 * CollyService est une structure qui encapsule le collecteur Colly pour le scraping de données.
 * @property {colly.Collector} collector - Instance du collecteur Colly pour le scraping.
 * @property {chan error} errChan - Canal pour signaler les erreurs pendant le scraping.
 */
type CollyService struct {
	collector *colly.Collector
	errChan   chan error // Canal pour signaler les erreurs
}

/**
 * News est une structure pour stocker les informations sur les actualités des cryptomonnaies.
 * @property {string} Title - Titre de l'actualité.
 * @property {string} Link - Lien vers l'actualité.
 * @property {string} Description - Description de l'actualité.
 */
type News struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
}

/**
 * NewCollyService crée une nouvelle instance de CollyService avec une configuration de collecteur prédéfinie.
 * @return {CollyService} - Retourne une instance configurée de CollyService.
 */
func NewCollyService() *CollyService {
	// Configuration de base du collecteur
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"), // User-Agent pour éviter le blocage
		colly.IgnoreRobotsTxt(), // Ignorer les règles du fichier robots.txt
		colly.MaxDepth(1),       // Limiter la profondeur de recherche à 1 pour éviter les liens externes
		colly.Async(true),       // Activer le mode asynchrone pour le scraping
		colly.CacheDir("./tmp"), // Définir le répertoire de cache pour éviter de re-scraper les pages
		colly.DetectCharset(),   // Détecter automatiquement l'encodage de la page
	)

	// Définition des limites de requêtes pour éviter les blocages
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",             // Applique cette règle à tous les domaines visités par le collecteur
		Parallelism: 2,               // Limite à 2 requêtes simultanées pour ne pas surcharger le serveur
		Delay:       2 * time.Second, // Attente de 2 secondes entre chaque requête pour éviter un blocage par le serveur
	})

	// Retourne une nouvelle instance de CollyService avec un canal d'erreur
	return &CollyService{
		collector: c,
		errChan:   make(chan error), // Initialiser le canal d'erreurs
	}
}

/**
 * ScrapeNews lance le scraping des actualités sur les cryptomonnaies depuis une URL donnée dans une goroutine.
 * @param {string} url - L'URL de la page à scraper.
 */
func (collyService *CollyService) ScrapeNews(url string, natsService *NatsService) {
	// Afficher un message de démarrage
	fmt.Println("Démarrage du scraping des actualités depuis :", url)

	// Callback pour la div principale avec la classe `sc-dkzDqf cTsMI`
	collyService.collector.OnHTML("div.sc-dkzDqf.cTsMI", func(e *colly.HTMLElement) {
		e.ForEach("div", func(_ int, newsDiv *colly.HTMLElement) {
			title := newsDiv.ChildText("a")
			link := newsDiv.ChildAttr("a", "href")

			// Récupération de la description depuis la deuxième div enfant
			var description string
			newsDiv.ForEach("div", func(index int, descDiv *colly.HTMLElement) {
				if index == 1 {
					description = descDiv.Text
				}
			})

			// DEBUG : Affiche le titre, lien et description
			fmt.Printf("Titre : %s\nLien : %s\nDescription : %s\n\n", title, link, description)

			// Créez une instance de News avec les données extraites
			news := News{
				Title:       title,
				Link:        link,
				Description: description,
			}

			// Sérialisez l'instance News en JSON
			message, err := json.Marshal(news)
			if err != nil {
				fmt.Printf("Erreur lors de la conversion en JSON : %v\n", err)
				return
			}

			// Publier le message JSON via NatsService
			if err := natsService.Publish("crypto.news", string(message)); err != nil {
				fmt.Printf("Erreur lors de la publication sur NATS : %v\n", err)
			}
		})
	})

	// Gestion des erreurs pendant le scraping
	collyService.collector.OnError(func(_ *colly.Response, err error) {
		log.Printf("Erreur pendant le scraping : %v", err)
		select {
		case collyService.errChan <- err:
		default:
			fmt.Println("Canal d'erreurs déjà fermé")
		}
	})

	// Démarrer le scraping et capturer les erreurs
	if err := collyService.collector.Visit(url); err != nil {
		select {
		case collyService.errChan <- err:
		default:
			fmt.Println("Canal d'erreurs déjà fermé")
		}
	}

	// Attendre la fin des requêtes asynchrones
	collyService.collector.Wait()
}

/**
 * ErrorChannel retourne le canal d'erreurs pour le scraping.
 * @return {<-chan error} - Canal en lecture seule pour les erreurs de scraping.
 */
func (collyService *CollyService) ErrorChannel() <-chan error {
	return collyService.errChan
}
