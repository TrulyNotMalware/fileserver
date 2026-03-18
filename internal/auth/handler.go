package auth

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fileServer/internal/db"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const refreshTokenCookieName = "refresh_token"

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
	Password string `json:"password"` // base64(RSA encrypted password)
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
}

type publicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

func (h *Handler) PublicKey(w http.ResponseWriter, r *http.Request) {
	pubPEM, err := h.jwtManager.PublicKeyPEM()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, publicKeyResponse{PublicKey: pubPEM})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	ciphertext, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	plainPassword, err := h.jwtManager.DecryptPassword(ciphertext)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUserByUsername(h.db, req.Username)
	if err != nil || user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainPassword)); err != nil {
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

	setRefreshTokenCookie(w, refreshToken, expiresAt)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value

	claims, err := h.jwtManager.Verify(refreshToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, expiresAt, err := db.GetRefreshToken(h.db, refreshToken)
	if err != nil || userID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if time.Now().After(expiresAt) {
		_ = db.DeleteRefreshToken(h.db, refreshToken)
		clearRefreshTokenCookie(w)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtManager.IssueAccessToken(claims.UserID, claims.Username, claims.Role, h.accessTokenTTL)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = db.DeleteRefreshToken(h.db, cookie.Value)
	clearRefreshTokenCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func setRefreshTokenCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth",
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/auth",
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
