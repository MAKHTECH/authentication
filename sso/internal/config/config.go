package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sso/sso/pkg/directories"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string         `json:"env" env-default:"local"`
	AppID       int            `json:"app_id" env-required:"true"`
	StoragePath string         `json:"storage_path"`
	Secret      string         `json:"secret" env_required:"true"`
	Jwt         JwtConfig      `json:"jwt" env-required:"true"`
	GRPC        GRPCConfig     `json:"grpc"`
	Telegram    TelegramConfig `json:"telegram"`
	Redis       RedisConfig    `json:"redis"`
	Kafka       KafkaConfig    `json:"kafka"`
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
	//cfg.TokenTTL = time.Duration(time.Hour * cfg.TokenTTL)
	cfg.GRPC.Timeout = time.Duration(time.Second * cfg.GRPC.Timeout)
	cfg.Jwt.AccessTokenTTL = time.Duration(time.Minute * cfg.Jwt.AccessTokenTTL)
	cfg.Jwt.RefreshTokenTTL = time.Duration(time.Minute * cfg.Jwt.RefreshTokenTTL)

	databaseDirectory := filepath.Join(directories.FindDirectoryName("cmd"), "../storage/sso.db")
	cfg.StoragePath = databaseDirectory

	// проверка ключа на 32 битность, для PASETO
	keyBytes := []byte(cfg.Secret)
	if len(keyBytes) != 32 {
		panic(fmt.Errorf("ключ должен быть длиной 32 байта, текущая длина: %d", len(keyBytes)))
	}

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
