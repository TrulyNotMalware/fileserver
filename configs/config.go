package configs

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//go:embed default.yaml
var defaultConfig []byte

type Config struct {
	Host            string
	Port            string
	StaticDir       string
	Permission      fs.FileMode
	Mode            Mode
	LogLevel        string
	PrivateKeyPath  string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	AdminPassword   string
	GuestPassword   string
}

func (c *Config) Print() {
	fmt.Println("---------- Configuration ----------")
	fmt.Printf("  Host            : %s\n", c.Host)
	fmt.Printf("  Port            : %s\n", c.Port)
	fmt.Printf("  StaticDir       : %s\n", c.StaticDir)
	fmt.Printf("  Permission      : %04o\n", c.Permission)
	fmt.Printf("  Mode            : %s\n", c.Mode)
	fmt.Printf("  LogLevel        : %s\n", c.LogLevel)
	fmt.Printf("  PrivateKeyPath  : %s\n", c.PrivateKeyPath)
	fmt.Printf("  AccessTokenTTL  : %s\n", c.AccessTokenTTL)
	fmt.Printf("  RefreshTokenTTL : %s\n", c.RefreshTokenTTL)
	fmt.Println("-----------------------------------")
}

func Load() (*Config, error) {
	v := viper.New()

	if err := loadDefaults(v); err != nil {
		return nil, err
	}
	bindEnv(v)
	if err := loadConfigFile(v); err != nil {
		return nil, err
	}

	return parseConfig(v)
}

func loadDefaults(v *viper.Viper) error {
	v.SetConfigType("yaml")
	return v.ReadConfig(bytes.NewReader(defaultConfig))
}

func bindEnv(v *viper.Viper) {
	v.SetEnvPrefix("FS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

func loadConfigFile(v *viper.Viper) error {
	pflag.String("config", "", "path to config file (e.g. --config ./config.yaml)")
	pflag.Parse()
	_ = v.BindPFlags(pflag.CommandLine)

	configPath := v.GetString("config")
	if configPath == "" {
		return nil
	}

	v.SetConfigFile(configPath)
	return v.MergeInConfig()
}

func parseConfig(v *viper.Viper) (*Config, error) {
	perm, err := parsePermission(v.GetString("server.permission"))
	if err != nil {
		return nil, err
	}

	mode, err := parseMode(v.GetString("server.mode"))
	if err != nil {
		return nil, err
	}

	accessTTL, err := time.ParseDuration(v.GetString("auth.access_token_ttl"))
	if err != nil {
		return nil, fmt.Errorf("invalid access_token_ttl: %w", err)
	}

	refreshTTL, err := time.ParseDuration(v.GetString("auth.refresh_token_ttl"))
	if err != nil {
		return nil, fmt.Errorf("invalid refresh_token_ttl: %w", err)
	}

	return &Config{
		Host:            v.GetString("server.host"),
		Port:            v.GetString("server.port"),
		StaticDir:       v.GetString("server.static_dir"),
		Permission:      perm,
		Mode:            mode,
		LogLevel:        v.GetString("logging.level"),
		PrivateKeyPath:  v.GetString("auth.private_key_path"),
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
		AdminPassword:   v.GetString("auth.admin_password"),
		GuestPassword:   v.GetString("auth.guest_password"),
	}, nil
}

func parsePermission(s string) (fs.FileMode, error) {
	perm, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid permission %q: must be octal (e.g. 0755): %w", s, err)
	}
	return fs.FileMode(perm), nil
}
