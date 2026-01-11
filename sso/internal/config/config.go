package config

import (
	"crypto/ed25519"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"sso/sso/pkg/directories"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string         `json:"env" env-default:"local"`
	AppID      int            `json:"app_id" env-required:"true"`
	Postgres   PostgresConfig `json:"postgres"`
	PrivateKey string         `json:"private_key" env_required:"true"`
	PublicKey  string         `json:"public_key"` // Вычисляется из private_key
	Jwt        JwtConfig      `json:"jwt" env-required:"true"`
	GRPC       GRPCConfig     `json:"grpc"`
	Telegram   TelegramConfig `json:"telegram"`
	Redis      RedisConfig    `json:"redis"`
	Kafka      KafkaConfig    `json:"kafka"`
}

type PostgresConfig struct {
	Host     string `json:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `json:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `json:"user" env:"POSTGRES_USER" env-default:"sso"`
	Password string `json:"password" env:"POSTGRES_PASSWORD" env-default:"sso_password"`
	DBName   string `json:"dbname" env:"POSTGRES_DB" env-default:"sso"`
	SSLMode  string `json:"sslmode" env-default:"disable"`
}

type GRPCConfig struct {
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

type TelegramConfig struct {
	CallbackPort int    `json:"callback_port"`
	Token        string `json:"token"`
}

type RedisConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	DB   int    `json:"db"`
}

type JwtConfig struct {
	AccessTokenTTL  time.Duration `json:"access_token_expiration_minute"`
	RefreshTokenTTL time.Duration `json:"refresh_token_expiration_minute"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		path = filepath.Join(directories.FindDirectoryName("config"), "local.json")
	}

	return MustLoadByPath(path)
}

func MustLoadByPath(path string) *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	// Переводим в нужные форматы времени
	cfg.GRPC.Timeout = time.Duration(time.Second * cfg.GRPC.Timeout)
	cfg.Jwt.AccessTokenTTL = time.Duration(time.Minute * cfg.Jwt.AccessTokenTTL)
	cfg.Jwt.RefreshTokenTTL = time.Duration(time.Minute * cfg.Jwt.RefreshTokenTTL)

	// Загружаем параметры PostgreSQL из переменных окружения (приоритет над конфигом)
	if envHost := os.Getenv("POSTGRES_HOST"); envHost != "" {
		cfg.Postgres.Host = envHost
	}
	if envPort := os.Getenv("POSTGRES_PORT"); envPort != "" {
		if port, err := strconv.Atoi(envPort); err == nil {
			cfg.Postgres.Port = port
		}
	}
	if envUser := os.Getenv("POSTGRES_USER"); envUser != "" {
		cfg.Postgres.User = envUser
	}
	if envPassword := os.Getenv("POSTGRES_PASSWORD"); envPassword != "" {
		cfg.Postgres.Password = envPassword
	}
	if envDB := os.Getenv("POSTGRES_DB"); envDB != "" {
		cfg.Postgres.DBName = envDB
	}

	// Логируем подключение к базе данных для отладки
	fmt.Printf("PostgreSQL: %s@%s:%d/%s\n", cfg.Postgres.User, cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.DBName)

	// Проверка и валидация Ed25519 приватного ключа
	privateKeyBytes, err := hex.DecodeString(cfg.PrivateKey)
	if err != nil {
		panic(fmt.Errorf("не удалось декодировать приватный ключ (должен быть hex): %w", err))
	}
	if len(privateKeyBytes) != ed25519.PrivateKeySize {
		panic(fmt.Errorf("приватный ключ должен быть %d байт (128 hex символов), получено: %d байт", ed25519.PrivateKeySize, len(privateKeyBytes)))
	}

	// Вычисляем публичный ключ из приватного
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	cfg.PublicKey = hex.EncodeToString(publicKey)

	return &cfg
}

func fetchConfigPath() string {
	var res string

	// --config="path/to/config.yml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
