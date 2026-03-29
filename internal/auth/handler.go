package auth

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fileServer/internal/db"
	"fileServer/internal/logger"
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

func (h *Handler) PublicKey(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("PublicKey request from %s", r.RemoteAddr)
	pubPEM, err := h.jwtManager.PublicKeyPEM()
	if err != nil {
		logger.Errorf("failed to get public key: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, publicKeyResponse{PublicKey: pubPEM})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Debugf("Login: failed to decode request body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	logger.Debugf("Login request for user: %s", req.Username)

	ciphertext, err := base64.StdEncoding.DecodeString(req.Password)
	if err != nil {
		logger.Debugf("Login: failed to base64 decode password for user %s: %v", req.Username, err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	plainPassword, err := h.jwtManager.DecryptPassword(ciphertext)
	if err != nil {
		logger.Debugf("Login: failed to decrypt password for user %s", req.Username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := db.GetUserByUsername(h.db, req.Username)
	if err != nil || user == nil {
		logger.Debugf("Login: user not found: %s", req.Username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(plainPassword)); err != nil {
		logger.Debugf("Login: invalid password for user %s", req.Username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtManager.IssueAccessToken(user.ID, user.Username, string(user.Role), h.accessTokenTTL)
	if err != nil {
		logger.Errorf("Login: failed to issue access token for user %s: %v", req.Username, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtManager.IssueRefreshToken(user.ID, user.Username, string(user.Role), h.refreshTokenTTL)
	if err != nil {
		logger.Errorf("Login: failed to issue refresh token for user %s: %v", req.Username, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	expiresAt := time.Now().Add(h.refreshTokenTTL)
	if err := db.SaveRefreshToken(h.db, user.ID, refreshToken, expiresAt); err != nil {
		logger.Errorf("Login: failed to save refresh token for user %s: %v", req.Username, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Infof("Login: user %s (role: %s) logged in", user.Username, user.Role)
	setRefreshTokenCookie(w, refreshToken, expiresAt)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("Refresh request from %s", r.RemoteAddr)
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		logger.Debugf("Refresh: no refresh token cookie")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value

	claims, err := h.jwtManager.Verify(refreshToken)
	if err != nil {
		logger.Debugf("Refresh: invalid refresh token: %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, expiresAt, err := db.GetRefreshToken(h.db, refreshToken)
	if err != nil || userID == 0 {
		logger.Debugf("Refresh: refresh token not found in db for user %s", claims.Username)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if time.Now().After(expiresAt) {
		logger.Debugf("Refresh: refresh token expired for user %s", claims.Username)
		_ = db.DeleteRefreshToken(h.db, refreshToken)
		clearRefreshTokenCookie(w)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	accessToken, err := h.jwtManager.IssueAccessToken(claims.UserID, claims.Username, claims.Role, h.accessTokenTTL)
	if err != nil {
		logger.Errorf("Refresh: failed to issue access token for user %s: %v", claims.Username, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logger.Debugf("Refresh: issued new access token for user %s", claims.Username)
	writeJSON(w, http.StatusOK, tokenResponse{AccessToken: accessToken})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	logger.Debugf("Logout request from %s", r.RemoteAddr)
	cookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		logger.Debugf("Logout: no refresh token cookie, nothing to do")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = db.DeleteRefreshToken(h.db, cookie.Value)
	clearRefreshTokenCookie(w)
	logger.Infof("Logout: refresh token removed")
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
