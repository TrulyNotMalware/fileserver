package internal

import "fileServer/internal/auth"

func InitRouter(server *Server, authHandler *auth.Handler) {
	server.mux.HandleFunc("GET /auth/public-key", authHandler.PublicKey)
	server.mux.HandleFunc("POST /auth/login", authHandler.Login)
	server.mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	server.mux.HandleFunc("POST /auth/logout", authHandler.Logout)
}
