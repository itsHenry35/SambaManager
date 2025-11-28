package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Admin struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"admin"`
	HomeDir string `yaml:"home_dir"`
	Samba   struct {
		ConfigPath string `yaml:"config_path"`
	} `yaml:"samba"`
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	JWT struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
}

var AppConfig *Config

// generateRandomSecret generates a random 32-byte secret key
func generateRandomSecret() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to a timestamp-based secret if random fails
		return fmt.Sprintf("fallback-secret-%d", os.Getpid())
	}
	return hex.EncodeToString(bytes)
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(configPath string) error {
	defaultConfig := Config{}
	defaultConfig.Admin.Username = "admin"
	defaultConfig.Admin.Password = "admin"
	defaultConfig.HomeDir = "/home/samba"
	defaultConfig.Samba.ConfigPath = "/etc/samba/smb.conf"
	defaultConfig.Server.Port = "8080"
	defaultConfig.Server.Host = "0.0.0.0"
	defaultConfig.JWT.Secret = generateRandomSecret()

	data, err := yaml.Marshal(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	fmt.Printf("Created default configuration file at %s\n", configPath)
	fmt.Println("IMPORTANT: Please change the default admin password!")
	return nil
}

func Load(configPath string) error {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Printf("Configuration file not found. Creating default configuration...\n")
		if err := createDefaultConfig(configPath); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate JWT secret exists
	if cfg.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is missing in config file. Please regenerate config or add a secret")
	}

	AppConfig = &cfg
	return nil
}

// GetJWTSecret returns the JWT secret key
func GetJWTSecret() []byte {
	if AppConfig == nil || AppConfig.JWT.Secret == "" {
		// This should never happen if Load() was called successfully
		panic("JWT secret not initialized")
	}
	return []byte(AppConfig.JWT.Secret)
}
