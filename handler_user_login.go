package main

import (
	"encoding/json"
	"net/http"
)

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userInput := UserInput{}
	err := decoder.Decode(&userInput)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse input", err)
		return
	}
}
