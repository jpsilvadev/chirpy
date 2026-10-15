package main

import "net/http"

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Forbidden", nil)
		return
	}

	// Reset visit counter
	cfg.fileserverHits.Store(0)

	// Delete all users from the database
	if err := cfg.db.DeleteAllUsers(req.Context()); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete all users", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Reset successful"))
}
