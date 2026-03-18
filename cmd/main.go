package main

import (
	"fileServer/configs"
	"fileServer/internal"
	"os"
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

	srv, err := internal.NewServer(cfg)
	if err != nil {
		internal.Errorf("failed to initialize server: %v", err)
		os.Exit(1)
	}

	if err := srv.Run(); err != nil {
		internal.Errorf("server error: %v", err)
		os.Exit(1)
	}
}
