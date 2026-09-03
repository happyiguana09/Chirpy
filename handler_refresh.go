package main

import (
	"net/http"
	"time"

	"example.com/Chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Token string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "No token sent", err)
		return
	}

	refTokData, err := cfg.db.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find refresh token", err)
		return
	}

	if refTokData.ExpiresAt.Before(time.Now()) {
		respondWithError(w, http.StatusUnauthorized, "refresh token expired", err)
		return
	}

	if refTokData.RevokedAt.Valid {
		respondWithError(w, http.StatusUnauthorized, "refresh token revoked", err)
	}

	newToken, err := auth.MakeJWT(refTokData.UserID, cfg.jwt_secret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create new jwt", err)
	}

	respondWithJSON(w, http.StatusOK, response{
		Token: newToken,
	})

}
