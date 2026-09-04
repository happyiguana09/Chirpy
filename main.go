package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"example.com/Chirpy/internal/auth"
	"example.com/Chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwt_secret     string
	polka_key      string
}

type User struct {
	ID          uuid.UUID `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Email       string    `json:"email"`
	IsChirpyRed bool      `json:"is_chirpy_red"`
}

type UserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}

	jwt_secret := os.Getenv("JWT_SECRET")
	if dbURL == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	polka_key := os.Getenv("POLKA_KEY")
	if dbURL == "" {
		log.Fatal("POLKA_KEY must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	const filepathRoot = "."
	const port = "8080"

	cfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             database.New(db),
		platform:       platform,
		jwt_secret:     jwt_secret,
		polka_key:      polka_key,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("GET /admin/metrics", cfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", cfg.handlerHitsReset)
	mux.HandleFunc("POST /api/chirps", cfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", cfg.handlerGetChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", cfg.handlerGetChirpByID)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.handlerDeleteChirps)
	mux.HandleFunc("POST /api/users", cfg.handlerUserCreate)
	mux.HandleFunc("PUT /api/users", cfg.handlerUserUpdate)
	mux.HandleFunc("POST /api/login", cfg.handlerLogin)
	mux.HandleFunc("POST /api/refresh", cfg.handlerRefresh)
	mux.HandleFunc("POST /api/revoke", cfg.handlerRevoke)
	mux.HandleFunc("POST /api/polka/webhooks", cfg.handlerWebhook)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving on port: %s\n", port)
	log.Fatal(server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) handlerUserCreate(w http.ResponseWriter, r *http.Request) {
	type response struct {
		User
	}

	decoder := json.NewDecoder(r.Body)
	newUser := UserInput{}
	err := decoder.Decode(&newUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	hashedPassword, err := auth.HashPassword(newUser.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	createUserParams := database.CreateUserParams{
		Email:          newUser.Email,
		HashedPassword: hashedPassword,
	}

	user, err := cfg.db.CreateUser(r.Context(), createUserParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, response{
		User: User{
			ID:          user.ID,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
			Email:       user.Email,
			IsChirpyRed: user.IsChirpyRed,
		},
	})
}
