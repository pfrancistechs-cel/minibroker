package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pfrancistechs-cel/minibroker/backend/internal/broker"
	"github.com/pfrancistechs-cel/minibroker/backend/internal/handler"
	"github.com/pfrancistechs-cel/minibroker/backend/internal/storage"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://minibroker:minibroker_password@localhost:5434/minibroker?sslmode=disable"
	}

	log.Printf("[MiniBroker] Initializing message broker cluster on :%s...\n", port)

	// Initialize Neon PostgreSQL Storage
	db, err := storage.ConnectAndMigrate(databaseURL)
	if err != nil {
		log.Fatalf("[MiniBroker] Failed initializing database: %v\n", err)
	}
	defer db.Pool.Close()

	// Initialize Cluster Broker Coordinator
	b, err := broker.NewBroker(db)
	if err != nil {
		log.Fatalf("[MiniBroker] Failed initializing broker: %v\n", err)
	}

	// Register HTTP & SSE routes
	mux := http.NewServeMux()
	h := handler.NewHandler(b)
	h.RegisterRoutes(mux)

	// Wrap with CORS middleware
	corsMux := withCORS(mux)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      corsMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 0, // SSE streaming requires unbounded write timeout
		IdleTimeout:  60 * time.Second,
	}

	// Run server in background goroutine
	go func() {
		log.Printf("[MiniBroker] Broker daemon running on http://localhost:%s\n", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[MiniBroker] ListenAndServe failed: %v\n", err)
		}
	}()

	// Graceful shutdown handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[MiniBroker] Shutting down cluster cleanly...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[MiniBroker] Shutdown error: %v\n", err)
	}
	log.Println("[MiniBroker] Cluster offline.")
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
