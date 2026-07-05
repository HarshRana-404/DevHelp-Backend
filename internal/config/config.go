// Package config provides application configuration management using Viper.
// It supports multiple environments (development, production) loaded from YAML files.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration values.
type Config struct {
	App         AppConfig         `mapstructure:"app"`
	RateLimiter RateLimiterConfig `mapstructure:"rate_limiter"`
	CORS        CORSConfig        `mapstructure:"cors"`
	Log         LogConfig         `mapstructure:"log"`
}

// AppConfig contains core application settings.
type AppConfig struct {
	Name           string `mapstructure:"name"`
	Env            string `mapstructure:"env"`
	Port           int    `mapstructure:"port"`
	Version        string `mapstructure:"version"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

// RateLimiterConfig controls IP-based rate limiting behaviour.
type RateLimiterConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
}

// CORSConfig holds Cross-Origin Resource Sharing settings.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
}

// LogConfig controls logging behaviour.
type LogConfig struct {
	Level       string `mapstructure:"level"`
	FileEnabled bool   `mapstructure:"file_enabled"`
	FileDir     string `mapstructure:"file_dir"`
}

// Load reads the configuration file for the given environment from the configs/ directory.
// env should be either "development" or "production".
func Load(env string) (*Config, error) {
	v := viper.New()

	v.SetConfigName(strings.ToLower(env))
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	// Allow environment variables to override config values.
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: reading config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshalling config: %w", err)
	}

	return &cfg, nil
}

// IsDevelopment returns true when the app is running in development mode.
func (c *Config) IsDevelopment() bool {
	return strings.EqualFold(c.App.Env, "development")
}

// IsProduction returns true when the app is running in production mode.
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

// Address returns the TCP address string (e.g. ":8080") the server should listen on.
func (c *Config) Address() string {
	return fmt.Sprintf(":%d", c.App.Port)
}
