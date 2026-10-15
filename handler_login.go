package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jpsilvadev/chirpy/internal/auth"
	"github.com/jpsilvadev/chirpy/internal/database"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	params := parameters{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Something went wrong decoding", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve user", err)
		return
	}
	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if !match || err != nil {
		respondWithError(
			w,
			http.StatusUnauthorized,
			"Incorrect email or password",
			err,
		)
		return
	}

	type loginResponse struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}

	// using a 1 hour default JWT lifetime
	jwtExpiryTime := time.Hour
	token, err := auth.MakeJWT(user.ID, cfg.jwtSecret, jwtExpiryTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create access token", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()
	now := time.Now().UTC()
	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:     refreshToken,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    user.ID,
		ExpiresAt: now.Add(time.Hour * 24 * 60), // refresh tokens will be stored for 60 days
	}
	_, err = cfg.db.CreateRefreshToken(req.Context(), refreshTokenParams)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create refresh token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, loginResponse{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		Token:        token,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve Bearer Token", err)
		return
	}

	user, err := cfg.db.GetUserFromRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve user from refresh token", err)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create access token", err)
		return
	}

	type response struct {
		Token string `json:"token"`
	}
	respondWithJSON(w, http.StatusOK, response{Token: jwt})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Failed to retrieve Bearer Token", err)
		return
	}
	err = cfg.db.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not revoke refresh token", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
