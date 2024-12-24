package config_test

import (
	"github.com/instinctG/lru-cache/internal/config"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestMustLoad(t *testing.T) {
	// Мокаем загрузку конфигурации, чтобы не зависеть от реальных файлов
	t.Run("TestFlagsWithDefaultValues", func(t *testing.T) {
		os.Args = []string{"cmd"} // Симулируем запуск с дефолтными аргументами (без флагов)

		cfg := config.MustLoad()

		// Проверяем, что дефолтные значения конфигурации корректны
		assert.Equal(t, ":8080", cfg.Port)
		assert.Equal(t, 10, cfg.CacheSize)
		assert.Equal(t, "WARN", cfg.LogLevel)
		assert.Equal(t, time.Minute, cfg.DefaultCacheTTL)
	})

}
