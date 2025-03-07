package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"sso/sso/pkg/directories"
	"time"
)

type Config struct {
	Env         string      `json:"env" env-default:"local"`
	AppID       int         `json:"app_id" env-required:"true"`
	StoragePath string      `json:"storage_path" env_required:"true"`
	Secret      string      `json:"secret" env_required:"true"`
	Jwt         JwtConfig   `json:"jwt" env-required:"true"`
	GRPC        GRPCConfig  `json:"grpc"`
	Redis       RedisConfig `json:"redis"`
}

type GRPCConfig struct {
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

type RedisConfig struct {
	Addr     string `json:"addr"`
	Password string `json:"password"`
	DB       int    `json:"db"`
}

type JwtConfig struct {
	AccessTokenTTL  time.Duration `json:"access_token_expiration_minute"`
	RefreshTokenTTL time.Duration `json:"refresh_token_expiration_minute"`
}

func MustLoad() *Config {
	path := fetchConfigPath()
	if path == "" {
		path = directories.FindDirectoryName("config") + "\\local.json"
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

	databaseDirectory := directories.FindDirectoryName("protos") + "\\..\\sso\\storage\\sso.db"
	cfg.StoragePath = databaseDirectory

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
