package migrations

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// PostgresConfig содержит параметры подключения к PostgreSQL
type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ApplyMigrations применяет миграции к базе данных
// migrationsPath - путь к папке с миграциями
// migrationsTable - название таблицы для отслеживания миграций (по умолчанию "migrations")
func ApplyMigrations(cfg PostgresConfig, migrationsPath string, migrationsTable string) error {
	// Проверяем, что путь к миграциям указан
	if migrationsPath == "" {
		fmt.Println("migrations path is empty, skipping migrations")
		return nil
	}

	// Проверяем, существует ли папка с миграциями
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		fmt.Println("migrations directory does not exist, skipping migrations")
		return nil
	}

	// Проверяем, есть ли файлы миграций в папке
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	hasMigrations := false
	for _, entry := range entries {
		if !entry.IsDir() {
			hasMigrations = true
			break
		}
	}

	if !hasMigrations {
		fmt.Println("no migration files found, skipping migrations")
		return nil
	}

	if migrationsTable == "" {
		migrationsTable = "migrations"
	}

	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	if cfg.Port == 0 {
		cfg.Port = 5432
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s&x-migrations-table=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName, cfg.SSLMode, migrationsTable,
	)

	m, err := migrate.New(
		"file://"+migrationsPath,
		connStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return nil
		}
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	fmt.Println("migrations applied successfully")
	return nil
}
