package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jpsilvadev/chirpy/internal/auth"
	"github.com/jpsilvadev/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, req *http.Request) {
	chirps, err := cfg.db.GetAllChirps(req.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get chirps", err)
		return
	}

	response := make([]Chirp, len(chirps))
	for i, c := range chirps {
		response[i] = Chirp{
			ID:        c.ID,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
			Body:      c.Body,
			UserID:    c.UserID,
		}
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) HandlerGetChirp(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to parse chirp UUID", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get chirp", err)
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

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	// check for valid JWT
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve Bearer Token", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT", err)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding", err)
		return
	}

	cleanedBody, err := validateChirpContent(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid chirp content", err)
		return
	}

	chirpParams := database.CreateChirpParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      cleanedBody,
		UserID:    userID,
	}
	_, err = cfg.db.CreateChirp(req.Context(), chirpParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirpParams.ID,
		CreatedAt: chirpParams.CreatedAt,
		UpdatedAt: chirpParams.UpdatedAt,
		Body:      chirpParams.Body,
		UserID:    chirpParams.UserID,
	})
}

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	chirpUUID, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse chirp UUID", err)
		return
	}

	// grab and validate jwt
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve Bearer Token", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT", err)
		return
	}

	chirp, err := cfg.db.GetChirpByID(req.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Failed to get chirp", err)
		return
	}
	// ensure chirp belongs to requester ID
	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You are not allowed to delete this chirp", nil)
		return
	}

	numRowsDeleted, err := cfg.db.DeleteChirpByID(req.Context(), chirpUUID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete chirp", err)
		return
	}

	if numRowsDeleted == 0 {
		respondWithError(w, http.StatusNotFound, "Failed to locate chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// helpers
func replaceBadWords(s string) string {
	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(s, " ")
	for i, word := range words {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func validateChirpContent(body string) (cleanedBody string, err error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", fmt.Errorf("chirp content exceeds %d characters", maxChirpLength)
	}
	cleanedBody = replaceBadWords(body)
	return cleanedBody, nil
}
