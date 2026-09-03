package main

import (
	"encoding/json"
	"net/http"
	"time"

	"example.com/Chirpy/internal/auth"
	"example.com/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
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

	token, err := auth.MakeJWT(userData.ID, cfg.jwt_secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create token", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()

	createRefreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    userData.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}

	_, err = cfg.db.CreateRefreshToken(r.Context(), createRefreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:        userData.ID,
			CreatedAt: userData.CreatedAt,
			UpdatedAt: userData.UpdatedAt,
			Email:     userData.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}
