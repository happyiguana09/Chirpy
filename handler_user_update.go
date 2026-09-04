package main

import (
	"encoding/json"
	"net/http"

	"example.com/Chirpy/internal/auth"
	"example.com/Chirpy/internal/database"
)

func (cfg *apiConfig) handlerUserUpdate(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no token found", err)
		return
	}

	userID, err := auth.ValidateJWT(accessToken, cfg.jwt_secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "invalid token", err)
		return
	}

	inputParams := params{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&inputParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to read body", err)
		return
	}

	hashedPassword, err := auth.HashPassword(inputParams.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't hash password", err)
		return
	}

	user, err := cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          inputParams.Email,
		HashedPassword: hashedPassword,
		ID:             userID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	respondWithJSON(w, http.StatusOK, User{
		ID:          user.ID,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	})
}
