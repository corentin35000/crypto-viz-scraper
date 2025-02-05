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
 * @property {string} Content - Description de l'actualité.
 * @property {string} Image - Lien de l'image principale de l'actualité.
 */
type News struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Content string `json:"content"`
	Image   string `json:"image"`
}

/**
 * NewCollyService crée une nouvelle instance de CollyService avec une configuration de collecteur prédéfinie.
 * @return {CollyService} - Retourne une instance configurée de CollyService.
 */
func NewCollyService() *CollyService {
	// Configuration du collecteur
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
		colly.MaxDepth(2), // Aller jusqu'à la page article
		colly.Async(true),
		colly.DetectCharset(),
	)

	// Définition des limites pour éviter d'être bloqué
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		Delay:       2 * time.Second,
	})

	// Retourne une nouvelle instance de CollyService
	return &CollyService{
		collector: c,
		errChan:   make(chan error),
	}
}

/**
 * ScrapeNews récupère les liens des articles depuis la page principale puis scrape chaque article individuellement.
 * @param {string} url - L'URL de la page à scraper.
 */
func (collyService *CollyService) ScrapeNews(url string, natsService *NatsService) {
	fmt.Println("Démarrage du scraping des actualités depuis :", url)

	// Scraping de la liste des articles
	collyService.collector.OnHTML("div#article-grid", func(e *colly.HTMLElement) {
		e.ForEach("a", func(_ int, element *colly.HTMLElement) {
			articleURL := element.Request.AbsoluteURL(element.Attr("href")) // Récupérer le lien absolu de l'article
			fmt.Println("Article trouvé :", articleURL)

			// Scraper l'article
			collyService.ScrapeArticle(articleURL, natsService)
		})
	})

	// Gestion des erreurs
	collyService.collector.OnError(func(_ *colly.Response, err error) {
		log.Printf("Erreur pendant le scraping : %v", err)
		select {
		case collyService.errChan <- err:
		default:
			fmt.Println("Canal d'erreurs déjà fermé")
		}
	})

	// Lancer le scraping
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
 * ScrapeArticle récupère les informations d'un article donné.
 * @param {string} articleURL - URL de l'article à scraper.
 */
func (collyService *CollyService) ScrapeArticle(articleURL string, natsService *NatsService) {
	// Création d'un nouveau collecteur pour éviter les conflits
	articleCollector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.IgnoreRobotsTxt(),
		colly.Async(true),
		colly.DetectCharset(),
	)

	// Structure pour stocker l'article
	var news News
	news.Link = articleURL

	// Récupérer le titre
	articleCollector.OnHTML("h1.article-title", func(e *colly.HTMLElement) {
		news.Title = e.Text
	})

	// Récupérer le contenu
	articleCollector.OnHTML("div.article-content > h4.article-subtitle, div.article-content > p", func(e *colly.HTMLElement) {
		// Vérifier si ce n'est pas le titre déjà stocké
		if e.Text != news.Title {
			news.Content += e.Text + "\n"
		}
	})

	// Récupérer l'image principale
	articleCollector.OnHTML("img.main-illustration", func(e *colly.HTMLElement) {
		news.Image = e.Attr("src")
	})

	// Callback après le scraping de la page
	articleCollector.OnScraped(func(_ *colly.Response) {
		// Vérifier que les champs ne sont pas vides
		if news.Title == "" || news.Content == "" || news.Image == "" {
			fmt.Println("Données incomplètes, article ignoré :", news.Link)
			return
		}

		// Afficher les résultats
		fmt.Printf("Article scrappé :\nTitre : %s\nLien : %s\nImage : %s \nContenu : %s\n", news.Title, news.Link, news.Image, news.Content)

		// Convertir l'article en JSON
		message, err := json.Marshal(news)
		if err != nil {
			fmt.Printf("Erreur lors de la conversion en JSON : %v\n", err)
			return
		}

		// Publier sur NATS
		if err := natsService.Publish("crypto.news", string(message)); err != nil {
			fmt.Printf("Erreur lors de la publication sur NATS : %v\n", err)
		}
	})

	// Lancer le scraping
	if err := articleCollector.Visit(articleURL); err != nil {
		fmt.Printf("Erreur lors de la visite de l'article : %v\n", err)
	}
}
