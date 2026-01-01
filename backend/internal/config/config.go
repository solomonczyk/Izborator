package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Environment string
	LogLevel    string

	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
	Meili  MeilisearchConfig
	Queue  QueueConfig
	Google GoogleConfig
	OpenAI OpenAIConfig
	QualityGates QualityGatesConfig
}

// ServerConfig конфигурация HTTP сервера
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// DBConfig конфигурация PostgreSQL
type DBConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	MaxConnections  int
	MinConnections  int
	MaxIdleTime     time.Duration
	ConnMaxLifetime time.Duration
}

// RedisConfig конфигурация Redis
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// MeilisearchConfig конфигурация Meilisearch
type MeilisearchConfig struct {
	Host   string
	Port   int
	APIKey string
}

// QueueConfig конфигурация очереди сообщений
type QueueConfig struct {
	Type       string // "kafka", "rabbitmq"
	Brokers    []string
	Topic      string
	GroupID    string
	MaxWorkers int
}

// GoogleConfig конфигурация Google API
type GoogleConfig struct {
	APIKey string // Google Cloud API Key
	CX     string // Custom Search Engine ID
}

// OpenAIConfig конфигурация OpenAI API
type OpenAIConfig struct {
	APIKey string // OpenAI API Key
	Model  string // Модель для использования (по умолчанию gpt-4o-mini)
}

type QualityGateThresholds struct {
	ValidRateMin    float64
	QualityScoreMin float64
}

type QualityGatesConfig struct {
	Goods    QualityGateThresholds
	Services QualityGateThresholds
}

// Load загружает конфигурацию из переменных окружения
func Load() (*Config, error) {
	cfg := &Config{
		Environment: getEnv("ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		Server: ServerConfig{
			Port:         getEnvAsInt("SERVER_PORT", 8080),
			ReadTimeout:  getEnvAsDuration("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getEnvAsDuration("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getEnvAsDuration("SERVER_IDLE_TIMEOUT", 60*time.Second),
		},

		DB: DBConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvAsInt("DB_PORT", 5432),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", ""),
			Database:        getEnv("DB_NAME", "izborator"),
			MaxConnections:  getEnvAsInt("DB_MAX_CONNECTIONS", 50),
			MinConnections:  getEnvAsInt("DB_MIN_CONNECTIONS", 5),
			MaxIdleTime:     getEnvAsDuration("DB_MAX_IDLE_TIME", 30*time.Minute),
			ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 1*time.Hour),
		},

		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvAsInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},

		Meili: MeilisearchConfig{
			Host:   getEnv("MEILISEARCH_HOST", "localhost"),
			Port:   getEnvAsInt("MEILISEARCH_PORT", 7700),
			APIKey: getEnv("MEILISEARCH_API_KEY", ""),
		},

		Queue: QueueConfig{
			Type:       getEnv("QUEUE_TYPE", "rabbitmq"),
			Brokers:    getEnvAsSlice("QUEUE_BROKERS", []string{"localhost:5672"}),
			Topic:      getEnv("QUEUE_TOPIC", "scraping_tasks"),
			GroupID:    getEnv("QUEUE_GROUP_ID", "izborator_workers"),
			MaxWorkers: getEnvAsInt("QUEUE_MAX_WORKERS", 10),
		},

		Google: GoogleConfig{
			APIKey: getEnv("GOOGLE_API_KEY", ""),
			CX:     getEnv("GOOGLE_CX", ""),
		},

		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
			Model:  getEnv("OPENAI_MODEL", ""), // Пустое = gpt-4o-mini по умолчанию
		},
		QualityGates: QualityGatesConfig{
			Goods: QualityGateThresholds{
				ValidRateMin:    0.95,
				QualityScoreMin: 0.85,
			},
			Services: QualityGateThresholds{
				ValidRateMin:    0.80,
				QualityScoreMin: 0.70,
			},
		},

	}

	return cfg, nil
}

// Helper функции для работы с переменными окружения
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Простой парсинг через запятую
	// TODO: можно улучшить для более сложных случаев
	return []string{valueStr}
}

// DSN возвращает строку подключения к PostgreSQL
func (c *DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable client_encoding=UTF8",
		c.Host, c.Port, c.User, c.Password, c.Database,
	)
}

// Address возвращает адрес Redis
func (c *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// Address возвращает адрес Meilisearch
func (c *MeilisearchConfig) Address() string {
	return fmt.Sprintf("http://%s:%d", c.Host, c.Port)
}
