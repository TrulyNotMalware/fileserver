package internal

import (
	"fileServer/internal/auth"
	"fileServer/internal/file"
	"net/http"
)

func InitRouter(server *Server, authHandler *auth.Handler, fileHandler *file.Handler, jwtMiddleware func(http.Handler) http.Handler) {
	// Auth
	server.mux.HandleFunc("GET /auth/public-key", authHandler.PublicKey)
	server.mux.HandleFunc("POST /auth/login", authHandler.Login)
	server.mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	server.mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	// File
	server.mux.Handle("GET /files", jwtMiddleware(http.HandlerFunc(fileHandler.List)))
	server.mux.Handle("GET /files/download", jwtMiddleware(http.HandlerFunc(fileHandler.Download)))
	server.mux.Handle("POST /files/upload", jwtMiddleware(http.HandlerFunc(fileHandler.Upload)))
}
