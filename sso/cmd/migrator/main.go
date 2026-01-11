package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath, migrationsTable string
	var host, port, user, password, dbname, sslmode string

	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.StringVar(&host, "host", "", "postgres host")
	flag.StringVar(&port, "port", "5432", "postgres port")
	flag.StringVar(&user, "user", "", "postgres user")
	flag.StringVar(&password, "password", "", "postgres password")
	flag.StringVar(&dbname, "dbname", "", "postgres database name")
	flag.StringVar(&sslmode, "sslmode", "disable", "postgres sslmode")
	flag.Parse()

	// Приоритет: env переменные > флаги командной строки
	if envHost := os.Getenv("POSTGRES_HOST"); envHost != "" {
		host = envHost
	}
	if envPort := os.Getenv("POSTGRES_PORT"); envPort != "" {
		port = envPort
	}
	if envUser := os.Getenv("POSTGRES_USER"); envUser != "" {
		user = envUser
	}
	if envPassword := os.Getenv("POSTGRES_PASSWORD"); envPassword != "" {
		password = envPassword
	}
	if envDB := os.Getenv("POSTGRES_DB"); envDB != "" {
		dbname = envDB
	}

	if host == "" {
		panic("postgres host is required (use -host flag or POSTGRES_HOST env)")
	}
	if user == "" {
		panic("postgres user is required (use -user flag or POSTGRES_USER env)")
	}
	if password == "" {
		panic("postgres password is required (use -password flag or POSTGRES_PASSWORD env)")
	}
	if dbname == "" {
		panic("postgres dbname is required (use -dbname flag or POSTGRES_DB env)")
	}
	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&x-migrations-table=%s",
		user, password, host, port, dbname, sslmode, migrationsTable,
	)

	m, err := migrate.New(
		"file://"+migrationsPath,
		connStr,
	)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}

		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
