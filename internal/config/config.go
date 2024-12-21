package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
	"time"
)

type Config struct {
	Port            string        `env:"SERVER_HOST_PORT" envDefault:"localhost:8080"`
	CacheSize       int           `env:"CACHE_SIZE" envDefault:"10"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"WARN"`
	DefaultCacheTTL time.Duration `env:"DEFAULT_CACHE_TTL" envDefault:"1m"`
	ConfigPath      string        `env:"CONFIG_PATH"`
}

func MustLoad() *Config {
	var cfg Config

	//TODO : корректно доделать конфиг-path
	flag.StringVar(&cfg.ConfigPath, "config-path", "config/local.env", "Path to the config file")

	fmt.Println(cfg.ConfigPath)
	if err := godotenv.Load(cfg.ConfigPath); err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("cannot parse config: %s", err)
	}

	// Определение флагов
	flag.StringVar(&cfg.Port, "server-host-port", cfg.Port, "Address to run the server (e.g., localhost:8080)")
	flag.IntVar(&cfg.CacheSize, "cache-size", cfg.CacheSize, "Maximum cache size")
	flag.DurationVar(&cfg.DefaultCacheTTL, "default-cache-ttl", cfg.DefaultCacheTTL, "Default TTL for cache entries")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (e.g., DEBUG, INFO, WARN, ERROR)")

	// Парсинг флагов
	flag.Parse()

	return &cfg
}
