package config

import (
	"fmt"
	"os"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
}

type DatabaseConfig struct {
	DSN string
}

type ServerConfig struct {
	Port string
}

func LoadConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_URL", "postgres://root:123456@postgres-process:5432/processdb?sslmode=disable&connect_timeout=1&TimeZone=Asia/Shanghai"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", ":8003"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (c *Config) GetDSN() string {
	return c.Database.DSN
}

func (c *Config) GetServerPort() string {
	return c.Server.Port
}

func (c *Config) String() string {
	return fmt.Sprintf("Config{Database: %s, Server: %s}", c.Database.DSN, c.Server.Port)
}
