package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/OsoianMarcel/url-shortener/internal/app"
)

const (
	shutdownTimeout = 5
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	var wg sync.WaitGroup
	srvErr := make(chan error, 1)

	// Run the server.
	wg.Go(func() {
		srvErr <- a.Serve(ctx)
	})

	// Watch for signal or server error.
	select {
	case <-ctx.Done():
		log.Println("shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}

	case err := <-srvErr:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}

	wg.Wait()
}
