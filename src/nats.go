package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

/**
 * NatsService est une structure qui encapsule la connexion NATS et les abonnements
 */
type NatsService struct {
	nc            *nats.Conn
	subscriptions map[string]*nats.Subscription
	mu            sync.Mutex
}

/**
 * NewNatsService crée une nouvelle instance de NatsService
 * @param {string} natsURL - L'URL du serveur NATS
 * @return {*NatsService} - Retourne une nouvelle instance de NatsService
 * @return {error} - Retourne une erreur si la connexion a échoué
 */
func NewNatsService(natsURL string) (*NatsService, error) {
	// Options de connexion avec NKEY
	opt, err := nats.NkeyOptionFromSeed("private/seed.txt")
	if err != nil {
		log.Fatal("Erreur lors de la création de l'option NKEY : ", err)
	}

	// Connexion au serveur NATS
	nc, err := nats.Connect(natsURL, opt)
	if err != nil {
		return nil, fmt.Errorf("erreur de connexion à NATS : %v", err)
	} else {
		fmt.Println("Connexion établie avec le serveur NATS :", nc.ConnectedUrl())
	}

	// Retourne une nouvelle instance de NatsService
	return &NatsService{nc: nc, subscriptions: make(map[string]*nats.Subscription)}, nil
}

/**
 * Subscribe permet de s'abonner à un sujet avec une fonction de callback pour gérer les messages reçus
 * @param {string} subject - Le sujet auquel s'abonner
 * @param {func(string)} callback - La fonction de rappel pour gérer les messages reçus
 * @return {error}
 */
func (natsService *NatsService) Subscribe(subject string, callback func(message string)) error {
	// Verrouillage pour gérer la concurrence
	natsService.mu.Lock()
	defer natsService.mu.Unlock()

	// Vérification de la connexion active
	if natsService.nc == nil {
		return fmt.Errorf("pas de connexion active avec le serveur NATS")
	}

	// Création de l'abonnement
	sub, err := natsService.nc.Subscribe(subject, func(msg *nats.Msg) {
		callback(string(msg.Data))
	})

	// Vérification des erreurs
	if err != nil {
		return fmt.Errorf("échec de l'abonnement à %s : %v", subject, err)
	}

	// Stockage de l'abonnement
	natsService.subscriptions[subject] = sub
	fmt.Printf("Abonné au sujet : %s\n", subject)

	// Retourne nil pour indiquer le succès
	return nil
}

/**
 * Publish envoie un message sur un sujet spécifique
 * @param {string} subject - Le sujet sur lequel publier le message
 * @param {string} message - Le message à publier
 * @return {error} - Retourne une erreur si la publication a échoué
 */
func (natsService *NatsService) Publish(subject string, message string) error {
	// Verrouillage pour gérer la concurrence
	natsService.mu.Lock()
	defer natsService.mu.Unlock()

	// Vérification de la connexion active
	if natsService.nc == nil {
		return fmt.Errorf("pas de connexion active avec NATS")
	}

	// Publication du message
	if err := natsService.nc.Publish(subject, []byte(message)); err != nil {
		return fmt.Errorf("échec de la publication sur %s : %v", subject, err)
	} else {
		fmt.Printf("Message publié sur %s\n", subject)
	}

	// Retourne nil pour indiquer le succès
	return nil
}

/**
 * Unsubscribe se désabonne d'un sujet spécifique
 * @param {string} subject - Le sujet duquel se désabonner
 * @return {error} - Retourne une erreur si la désinscription a échoué
 */
func (natsService *NatsService) Unsubscribe(subject string) error {
	// Verrouillage pour gérer la concurrence
	natsService.mu.Lock()
	defer natsService.mu.Unlock()

	// Vérification de la connexion active
	sub, exists := natsService.subscriptions[subject]
	if !exists {
		return fmt.Errorf("aucun abonnement trouvé pour le sujet %s", subject)
	}

	// Désinscription du sujet
	if err := sub.Unsubscribe(); err != nil {
		return fmt.Errorf("échec de la désinscription du sujet %s : %v", subject, err)
	} else {
		delete(natsService.subscriptions, subject)
		fmt.Printf("Désabonné du sujet : %s\n", subject)
	}

	// Retourne nil pour indiquer le succès
	return nil
}

/**
 * UnsubscribeAll se désabonne de tous les sujets
 * @return {void} - Aucune valeur de retour
 */
func (natsService *NatsService) UnsubscribeAll() {
	// Verrouillage pour gérer la concurrence
	natsService.mu.Lock()
	defer natsService.mu.Unlock()

	// Désinscription de tous les sujets
	for subject, sub := range natsService.subscriptions {
		if err := sub.Unsubscribe(); err != nil {
			// Log the error
			log.Printf("Erreur lors de la désinscription de %s : %v\n", subject, err)
		} else {
			// Remove the subscription
			fmt.Printf("Désabonné du sujet : %s\n", subject)
		}
	}

	// Clear the subscriptions map
	natsService.subscriptions = make(map[string]*nats.Subscription)
}

/**
 * Close ferme la connexion NATS proprement
 * @return {void} - Aucune valeur de retour
 */
func (natsService *NatsService) Close() {
	// Verrouillage pour gérer la concurrence
	natsService.mu.Lock()
	defer natsService.mu.Unlock()

	// Vérification de la connexion active
	if natsService.nc == nil {
		fmt.Println("Connexion déjà fermée.")
		return
	}

	// Désinscription de tous les sujets
	natsService.UnsubscribeAll()

	// Fermeture avec un délai pour drainer les messages en cours
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fermeture de la connexion
	natsService.nc.Close()
	natsService.nc = nil

	// Message de confirmation
	fmt.Println("Connexion NATS fermée.")
}
