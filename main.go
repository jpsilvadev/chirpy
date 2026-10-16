package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	"github.com/jpsilvadev/chirpy/internal/database"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaAPIKey    string
}

func main() {
	const filepathRoot = "."
	const port = "8080"

	godotenv.Load()

	// i.e. dev | prod
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM env must be set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET env must be set")
	}

	polkaAPIKey := os.Getenv("POLKA_KEY")
	if polkaAPIKey == "" {
		log.Fatal("POLKA_KEY env must be set")
	}
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL env must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	dbQueries := database.New(db)
	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
		polkaAPIKey:    polkaAPIKey,
	}

	mux := http.NewServeMux()
	fileServer := http.StripPrefix(
		"/app",
		http.FileServer(http.Dir(filepathRoot)),
	)
	mux.Handle("/app/", cfg.middlewareMetricsInc(fileServer))

	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.HandlerGetChirp)

	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevokeToken)

	mux.HandleFunc("POST /api/users", cfg.handlerCreateUser)
	mux.HandleFunc("POST /api/chirps", cfg.handlerCreateChirp)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerPolkaWebhooks)

	mux.HandleFunc("PUT /api/users", cfg.handlerUpdateUser)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirp)

	mux.HandleFunc("POST /admin/reset", cfg.handlerReset)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)

	server := http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
