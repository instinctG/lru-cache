// Package config предоставляет функциональность для загрузки и парсинга конфигурации приложения.
// Поддерживается загрузка значений конфигурации из переменных окружения и файлов .env.
package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

// Config содержит значения конфигурации приложения.
type Config struct {
	Port            string        `env:"SERVER_HOST_PORT" envDefault:"localhost:8080"` // Порт, на котором будет запущен сервер.
	CacheSize       int           `env:"CACHE_SIZE" envDefault:"10"`                   // Максимальное количество элементов в кэше.
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"INFO"`                  // Уровень логирования приложения.
	DefaultCacheTTL time.Duration `env:"DEFAULT_CACHE_TTL" envDefault:"1m"`            // Время жизни записей в кэше по умолчанию.
}

// MustLoad загружает конфигурацию приложения.
// Алгоритм загрузки:
// 1. Проверяется наличие переменной окружения CONFIG_PATH, которая указывает путь до файла конфигурации.
// 2. Если путь указан, проверяется существование файла.
// 3. Загружаются переменные окружения из указанного файла .env (если файл существует).
// Политика конфигурирования (общая для всех параметров):
// Если для параметра определен флаг запуска, используется он
// Если флаг не определен, используется переменная окружения
// Если не определены ни флаг, ни переменная окружения, используется значение по умолчанию
// 4. Загружаются значения из переменных окружения и флагов командной строки.
// Если переменная окружения CONFIG_PATH не указан, приложение завершает работу с ошибкой.
func MustLoad() *Config {
	var cfg Config

	// Проверяем путь до конфигурационного файла из переменной окружения
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// Проверяем существование файла конфигурации
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("config file does not exist: %s", configPath)
	}

	// Загружаем переменные окружения из файла конфигурации (если он существует)
	if err := godotenv.Load(configPath); err != nil {
		log.Printf("error loading .env file: %s", err)
		log.Printf("config will be set by default")
	}

	// Парсим переменные окружения в структуру конфигурации
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("cannot parse config: %s", err)
	}

	// Парсим флаги командной строки
	flag.StringVar(&cfg.Port, "server-host-port", cfg.Port, "Address to run the server (e.g., localhost:8080)")
	flag.IntVar(&cfg.CacheSize, "cache-size", cfg.CacheSize, "Maximum cache size")
	flag.DurationVar(&cfg.DefaultCacheTTL, "default-cache-ttl", cfg.DefaultCacheTTL, "Default TTL for cache entries ms,s,m,...")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (e.g., DEBUG, INFO, WARN, ERROR)")

	// Применяем значения флагов
	flag.Parse()

	return &cfg
}
