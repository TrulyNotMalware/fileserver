package main

import (
	"database/sql"
	"fileServer/configs"
	"fileServer/internal"
	"fileServer/internal/auth"
	"fileServer/internal/db"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		internal.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	level, err := internal.ParseLevel(cfg.LogLevel)
	if err != nil {
		internal.Errorf("failed to parse log level: %v", err)
		os.Exit(1)
	}
	internal.Init(level)

	cfg.Print()

	if cfg.PrivateKeyPath == "" {
		internal.Errorf("auth.private_key_path is required")
		os.Exit(1)
	}
	if cfg.AdminPassword == "" || cfg.GuestPassword == "" {
		internal.Errorf("auth.admin_password and auth.guest_password are required")
		os.Exit(1)
	}

	jwtManager, err := auth.NewJWTManager(cfg.PrivateKeyPath)
	if err != nil {
		internal.Errorf("failed to load private key: %v", err)
		os.Exit(1)
	}

	database, err := db.New("./fileserver.db")
	if err != nil {
		internal.Errorf("failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer func(database *sql.DB) {
		err := database.Close()
		if err != nil {
			return
		}
	}(database)

	if err := seedUsers(database, cfg.AdminPassword, cfg.GuestPassword); err != nil {
		internal.Errorf("failed to seed users: %v", err)
		os.Exit(1)
	}

	srv, err := internal.NewServer(
		cfg.Host, cfg.Port, cfg.StaticDir, cfg.Permission, string(cfg.Mode),
		database, jwtManager,
		cfg.AccessTokenTTL, cfg.RefreshTokenTTL,
	)
	if err != nil {
		internal.Errorf("failed to initialize server: %v", err)
		os.Exit(1)
	}

	if err := srv.Run(); err != nil {
		internal.Errorf("server error: %v", err)
		os.Exit(1)
	}
}

func seedUsers(database *sql.DB, adminPassword, guestPassword string) error {
	adminHash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	guestHash, err := bcrypt.GenerateFromPassword([]byte(guestPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if err := db.UpsertUser(database, "admin", string(adminHash), db.RoleAdmin); err != nil {
		return err
	}

	return db.UpsertUser(database, "guest", string(guestHash), db.RoleGuest)
}
