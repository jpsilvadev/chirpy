package main

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
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
