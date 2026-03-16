package auth

import (
	"database/sql"
	"encoding/json"
	"fileServer/internal/db"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	db              *sql.DB
	jwtManager      *JWTManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewHandler(database *sql.DB, jwtManager *JWTManager, accessTTL, refreshTTL time.Duration) *Handler {
	return &Handler{
		db:              database,
		jwtManager:      jwtManager,
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	user, err := db.GetUserByUsername(h.db, req.Username)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtManager.IssueAccessToken(user.ID, user.Username, string(user.Role), h.accessTokenTTL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtManager.IssueRefreshToken(user.ID, user.Username, string(user.Role), h.refreshTokenTTL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(h.refreshTokenTTL)
	if err := db.SaveRefreshToken(h.db, user.ID, refreshToken, expiresAt); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	claims, err := h.jwtManager.Verify(req.RefreshToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, expiresAt, err := db.GetRefreshToken(h.db, req.RefreshToken)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if time.Now().After(expiresAt) {
		_ = db.DeleteRefreshToken(h.db, req.RefreshToken)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtManager.IssueAccessToken(claims.UserID, claims.Username, claims.Role, h.accessTokenTTL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken:  accessToken,
		RefreshToken: req.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req logoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := db.DeleteRefreshToken(h.db, req.RefreshToken); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
