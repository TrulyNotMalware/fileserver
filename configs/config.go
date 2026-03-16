package configs

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/fs"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

//go:embed default.yaml
var defaultConfig []byte

type Config struct {
	Host       string
	Port       string
	StaticDir  string
	Permission fs.FileMode
	Mode       Mode
	LogLevel   string
}

func (c *Config) Print() {
	fmt.Println("---------- Configuration ----------")
	fmt.Printf("  Host       : %s\n", c.Host)
	fmt.Printf("  Port       : %s\n", c.Port)
	fmt.Printf("  StaticDir  : %s\n", c.StaticDir)
	fmt.Printf("  Permission : %04o\n", c.Permission)
	fmt.Printf("  Mode       : %s\n", c.Mode)
	fmt.Printf("  LogLevel   : %s\n", c.LogLevel)
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

// loadDefaults loads the embedded YAML as default values.
func loadDefaults(v *viper.Viper) error {
	v.SetConfigType("yaml")
	return v.ReadConfig(bytes.NewReader(defaultConfig))
}

// bindEnv configures environment variable bindings.
func bindEnv(v *viper.Viper) {
	v.SetEnvPrefix("FS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

// loadConfigFile loads the config file specified by the --config flag.
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

// parseConfig maps viper values into a Config struct.
func parseConfig(v *viper.Viper) (*Config, error) {
	perm, err := parsePermission(v.GetString("server.permission"))
	if err != nil {
		return nil, err
	}

	mode, err := parseMode(v.GetString("server.mode"))
	if err != nil {
		return nil, err
	}

	return &Config{
		Host:       v.GetString("server.host"),
		Port:       v.GetString("server.port"),
		StaticDir:  v.GetString("server.static_dir"),
		Permission: perm,
		Mode:       mode,
		LogLevel:   v.GetString("logging.level"),
	}, nil
}

// parsePermission converts an octal string into fs.FileMode.
func parsePermission(s string) (fs.FileMode, error) {
	perm, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid permission %q: must be octal (e.g. 0755): %w", s, err)
	}
	return fs.FileMode(perm), nil
}
