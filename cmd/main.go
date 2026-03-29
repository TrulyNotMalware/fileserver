package main

import (
	"fileServer/configs"
	"fileServer/internal"
	"fileServer/internal/logger"
	"os"
)

func main() {
	cfg, err := configs.Load()
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		os.Exit(1)
	}

	level, err := logger.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Errorf("failed to parse log level: %v", err)
		os.Exit(1)
	}
	logger.Init(level)

	cfg.Print()

	srv, err := internal.NewServer(cfg)
	if err != nil {
		logger.Errorf("failed to initialize server: %v", err)
		os.Exit(1)
	}

	if err := srv.Run(); err != nil {
		logger.Errorf("server error: %v", err)
		os.Exit(1)
	}
}
