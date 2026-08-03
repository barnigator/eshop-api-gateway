package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/barnigator/eshop-api-gateway/internal/domain"
)

type AuthUseCase interface {
	Register(ctx context.Context, email, password string) (int64, error)
	Login(ctx context.Context, email, password string) (string, error)
}

type Handler struct {
	uc AuthUseCase
}

func New(uc AuthUseCase) *Handler {
	return &Handler{uc: uc}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	UserID int64 `json:"user_id"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	err := decoder.Decode(&req)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid input data")
		return
	}

	err = decoder.Decode(&struct{}{})
	if err != io.EOF {
		respondWithError(w, http.StatusBadRequest, "invalid input data")
		return
	}

	id, err := h.uc.Register(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrEmailRequired) {
			respondWithError(w, http.StatusBadRequest, domain.ErrEmailRequired.Error())
			return
		}

		if errors.Is(err, domain.ErrPasswordRequired) {
			respondWithError(w, http.StatusBadRequest, domain.ErrPasswordRequired.Error())
			return
		}

		respondWithError(w, http.StatusInternalServerError, "internal error")
		return
	}

	respondWithJSON(w, http.StatusCreated, registerResponse{UserID: id})
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}
