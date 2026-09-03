package main

import (
	"encoding/json"
	"net/http"
	"time"

	"example.com/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	type response struct {
		User
		Token string `json:"token"`
	}

	decoder := json.NewDecoder(r.Body)
	userInput := parameters{}
	err := decoder.Decode(&userInput)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse input", err)
		return
	}

	userData, err := cfg.db.GetHashedPassword(r.Context(), userInput.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	loginSuccess, err := auth.CheckPasswordHash(userInput.Password, userData.HashedPassword)
	if err != nil || !loginSuccess {
		respondWithError(w, http.StatusUnauthorized, "incorrect email or password", err)
		return
	}

	expiresIn := time.Hour
	if userInput.ExpiresInSeconds > 0 || userInput.ExpiresInSeconds < 3600 {
		expiresIn = time.Duration(userInput.ExpiresInSeconds) * time.Second
	}

	token, err := auth.MakeJWT(userData.ID, cfg.jwt_secret, expiresIn)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        userData.ID,
			CreatedAt: userData.CreatedAt,
			UpdatedAt: userData.UpdatedAt,
			Email:     userData.Email,
		},
		Token: token,
	})
}
