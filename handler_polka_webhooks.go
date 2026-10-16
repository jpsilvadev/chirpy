package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jpsilvadev/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	// ensure API Key is correct
	APIKey, err := auth.GetAPIKey(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve ApiKey", err)
		return
	}

	if APIKey != cfg.polkaAPIKey {
		respondWithError(w, http.StatusUnauthorized, "Invalid API Key", nil)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding", err)
		return
	}

	// dont care about other events
	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if params.Event == "user.upgraded" {
		_, err := cfg.db.UpgradeToChirpyRed(req.Context(), params.Data.UserID)
		if err != nil {
			if err == sql.ErrNoRows {
				respondWithError(w, http.StatusNotFound, "Failed to locate user ID", err)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to upgrade user", err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
