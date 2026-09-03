package main

import (
	"net/http"

	"example.com/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	refTok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "No token found in header", err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), refTok)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update record", err)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
