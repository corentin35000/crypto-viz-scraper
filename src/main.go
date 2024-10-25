// Point d'entrée de l'application

package main

import (
	"github.com/air-verse/air"
	"github.com/gocolly/colly/v2"
	"github.com/nats-io/nats.go"
)

func main() {
	_ = colly.NewCollector()             // Utilisation minimale de Colly
	_ = air.RootCmd                      // Utilisation minimale d'Air
	_, _ = nats.Connect(nats.DefaultURL) // Utilisation minimale de NATS
}
