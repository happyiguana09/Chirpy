package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"example.com/Chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerWebhook(w http.ResponseWriter, r *http.Request) {
	type params struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No api key found in request", err)
		return
	}

	if apiKey != cfg.polka_key {
		respondWithError(w, http.StatusUnauthorized, "Invalid API key", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	inputs := params{}
	err = decoder.Decode(&inputs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse json", err)
		return
	}

	if inputs.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpdateChirpyRed(r.Context(), inputs.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Couldn't find user in database", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Couldn't update user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
