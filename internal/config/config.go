package config

import (
	"flag"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Port            string        `env:"SERVER_HOST_PORT" envDefault:"localhost:8080"`
	CacheSize       int           `env:"CACHE_SIZE" envDefault:"10"`
	LogLevel        string        `env:"LOG_LEVEL" envDefault:"WARN"`
	DefaultCacheTTL time.Duration `env:"DEFAULT_CACHE_TTL" envDefault:"1m"`
}

func MustLoad() *Config {
	var cfg Config

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	if err := godotenv.Load(configPath); err != nil {
		log.Fatalf("error loading .env file: %s", err)
	}

	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("cannot parse config: %s", err)
	}

	flag.StringVar(&cfg.Port, "server-host-port", cfg.Port, "Address to run the server (e.g., localhost:8080)")
	flag.IntVar(&cfg.CacheSize, "cache-size", cfg.CacheSize, "Maximum cache size")
	flag.DurationVar(&cfg.DefaultCacheTTL, "default-cache-ttl", cfg.DefaultCacheTTL, "Default TTL for cache entries")
	flag.StringVar(&cfg.LogLevel, "log-level", cfg.LogLevel, "Log level (e.g., DEBUG, INFO, WARN, ERROR)")

	flag.Parse()

	return &cfg
}
