package main

import (
	"fmt"

	_ "github.com/gocolly/colly/v2"
	_ "github.com/nats-io/nats.go"
)

// Point d'entrée de l'application
func main() {
	fmt.Println("Hello world !")
}
