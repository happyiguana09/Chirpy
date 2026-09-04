package main

import (
	"net/http"
	"sort"

	"example.com/Chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	userIDString := r.URL.Query().Get("author_id")

	var chirps []database.Chirp
	var err error

	if userIDString == "" {
		chirps, err = cfg.db.GetChirps(r.Context())
	} else {
		userID, err := uuid.Parse(userIDString)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid id", err)
			return
		}
		chirps, err = cfg.db.GetUserChirps(r.Context(), userID)
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	sortType := r.URL.Query().Get("sort")
	if sortType == "desc" {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		})
	}

	jsonChirps := []Chirp{}

	for _, chirp := range chirps {
		jsonChirps = append(jsonChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, jsonChirps)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse input uuid", err)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't find chirp", err)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}
