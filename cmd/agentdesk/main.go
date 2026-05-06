package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"biai/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := app.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
