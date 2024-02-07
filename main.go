package main

import (
	"context"
	"log"

	"github.com/jhondevcode/orders-api/application"
)

func main() {
	app := application.New()
	err := app.Start(context.TODO())

	if err != nil {
		log.Fatal(err)
	}
}
