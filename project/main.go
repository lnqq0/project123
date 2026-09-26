package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// withCORS разрешает браузеру обращаться к API со страницы с другого адреса
// и отвечает на предварительный запрос OPTIONS.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// withLogging пишет в лог каждый запрос и время его обработки.
func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")

	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("failed to parse database config: %v", err)
	}

	config.MaxConns = 10
	config.MinConns = 2

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("failed to create database pool: %v", err)
	}
	defer db.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := db.Ping(pingCtx); err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}

	log.Println("Connected to PostgreSQL")

	store := NewStore(db)
	handler := NewHandler(store)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Users
	mux.HandleFunc("GET /users", handler.GetUsers)
	mux.HandleFunc("GET /users/{id}", handler.GetUser)
	mux.HandleFunc("POST /users", handler.CreateUser)
	mux.HandleFunc("PATCH /users/{id}", handler.UpdateUser)
	mux.HandleFunc("DELETE /users/{id}", handler.DeleteUser)

	// Categories
	mux.HandleFunc("GET /categories", handler.GetCategories)
	mux.HandleFunc("GET /categories/{id}", handler.GetCategory)
	mux.HandleFunc("POST /categories", handler.CreateCategory)
	mux.HandleFunc("PATCH /categories/{id}", handler.UpdateCategory)
	mux.HandleFunc("DELETE /categories/{id}", handler.DeleteCategory)

	// Tickets
	mux.HandleFunc("GET /tickets", handler.GetTickets)
	mux.HandleFunc("GET /tickets/{id}", handler.GetTicket)
	mux.HandleFunc("POST /tickets", handler.CreateTicket)
	mux.HandleFunc("PATCH /tickets/{id}", handler.UpdateTicket)
	mux.HandleFunc("DELETE /tickets/{id}", handler.DeleteTicket)

	// Comments
	mux.HandleFunc("GET /comments", handler.GetComments)
	mux.HandleFunc("GET /comments/{id}", handler.GetComment)
	mux.HandleFunc("POST /comments", handler.CreateComment)
	mux.HandleFunc("PATCH /comments/{id}", handler.UpdateComment)
	mux.HandleFunc("DELETE /comments/{id}", handler.DeleteComment)

	// Attachments
	mux.HandleFunc("GET /attachments", handler.GetAttachments)
	mux.HandleFunc("GET /attachments/{id}", handler.GetAttachment)
	mux.HandleFunc("POST /attachments", handler.CreateAttachment)
	mux.HandleFunc("PATCH /attachments/{id}", handler.UpdateAttachment)
	mux.HandleFunc("DELETE /attachments/{id}", handler.DeleteAttachment)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           withLogging(withCORS(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("Server started on http://localhost:8080")

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
