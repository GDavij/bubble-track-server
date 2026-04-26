package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	Qdrant   QdrantConfig
	Redis    RedisConfig
	GenAI    GenAIConfig
	Ollama   OllamaConfig
}

type ServerConfig struct {
	Port          string
	Env           string
	APIURL        string
	WsURL         string
	DefaultUserID string
	JWTSecret     string
}

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
	SSLMode  string
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DB, c.SSLMode,
	)
}

type QdrantConfig struct {
	Host        string
	Port        int
	Collection string
	APIKey      string
	UseTLS      bool
}

func (c QdrantConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func (c RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type GenAIConfig struct {
	APIKey  string
	Model   string
	Backend string
}

type OllamaConfig struct {
	Host           string
	Port           int
	Model          string
	EmbeddingModel string
}

func (c OllamaConfig) URL() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}

func Load() *Config {
	// Load .env file if exists (doesn't error if missing)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port:          envOr("SERVER_PORT", "8080"),
			Env:           envOr("APP_ENV", "development"),
			APIURL:        envOr("API_URL", "http://localhost:8080"),
			WsURL:         envOr("WS_URL", "ws://localhost:8080/ws"),
			DefaultUserID:  envOr("DEFAULT_USER_ID", "default-user"),
			JWTSecret:      envOr("JWT_SECRET", "change-me-in-production"),
		},
		Postgres: PostgresConfig{
			Host:     envOr("POSTGRES_HOST", "postgres.database"),
			Port:     envIntOr("POSTGRES_PORT", 5432),
			User:     envOr("POSTGRES_USER", "bubbletrack"),
			Password: envOr("POSTGRES_PASSWORD", ""),
			DB:       envOr("POSTGRES_DB", "bubbletrack"),
			SSLMode:  envOr("POSTGRES_SSLMODE", "disable"),
		},
		Qdrant: QdrantConfig{
			Host:       envOr("QDRANT_HOST", "qdrant.database"),
			Port:       envIntOr("QDRANT_PORT", 6334),
			Collection: envOr("QDRANT_COLLECTION", "bubble_interactions"),
			APIKey:     envOr("QDRANT_API_KEY", ""),
			UseTLS:     envBoolOr("QDRANT_USE_TLS", false),
		},
		Redis: RedisConfig{
			Host:     envOr("REDIS_HOST", "redis-stack.stack"),
			Port:     envIntOr("REDIS_PORT", 6379),
			Password: envOr("REDIS_PASSWORD", ""),
			DB:       envIntOr("REDIS_DB", 0),
		},
		GenAI: GenAIConfig{
			APIKey:  envOr("GEMINI_API_KEY", ""),
			Model:   envOr("GENAI_MODEL", "gemini-2.5-flash"),
			Backend: envOr("GENAI_BACKEND", "gemini"),
		},
		Ollama: OllamaConfig{
			Host:           envOr("OLLAMA_HOST", "ollama"),
			Port:           envIntOr("OLLAMA_PORT", 11434),
			Model:          envOr("OLLAMA_MODEL", "gemma4:e4b"),
			EmbeddingModel: envOr("OLLAMA_EMBEDDING_MODEL", "nomic-embed-text"),
		},
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envIntOr(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envBoolOr(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}
