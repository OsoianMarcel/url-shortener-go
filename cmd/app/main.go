package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/app"
)

const (
	shutdownTimeout = 5 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New()
	if err != nil {
		log.Fatalf("init app: %v", err)
	}

	if args := os.Args[1:]; len(args) > 0 {
		if err := a.ServeCLI(ctx, args); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}

		return
	}

	var wg sync.WaitGroup
	srvErr := make(chan error, 1)

	wg.Go(func() {
		srvErr <- a.ServeHTTP(ctx)
	})

	select {
	case <-ctx.Done():
		log.Println("shutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := a.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		} else {
			log.Println("shutdown completed successfully")
		}

	case err := <-srvErr:
		if err != nil {
			log.Fatalf("server error: %v", err)
		}
	}

	wg.Wait()
}
