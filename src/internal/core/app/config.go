package app

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
)

type Env string

const (
	EnvUnknown Env = "unknown"
	EnvLocal   Env = "local"
	EnvDev     Env = "dev"
	EnvProd    Env = "prod"
)

func NewEnv(env string) Env {
	switch env {
	case "local":
		return EnvLocal
	case "dev":
		return EnvDev
	case "prod":
		return EnvProd
	default:
		return EnvUnknown
	}
}

func init() {
	err := godotenv.Load()
	if err != nil {
		slog.Warn("Unable to load .env file")
	}
}

type Config struct {
	Env                 Env
	Port                string
	DbPath              string
	JwtSecret           string
	SpotifyTokenSecret  string
	Host                string
	StateCode           string
	SpotifyClientId     string
	SpotifyClientSecret string
	DiscogsKey          string
	DiscogsSecret       string
	AppName             string
	AppVersion          string
	ContactEmail        string
}

func LoadConfig() *Config {
	env := NewEnv(GetEnvWithPanic("ENV"))
	port := GetEnvWithDefault("PORT", "8080")
	host := GetEnvWithConditionalPanic("HOST", fmt.Sprintf("http://127.0.0.1:%s", port), env != EnvLocal)

	return &Config{
		Env:                 env,
		Port:                port,
		DbPath:              GetEnvWithDefault("DB_PATH", "./tmp/db.sql"),
		JwtSecret:           GetEnvWithConditionalPanic("JWT_SECRET", "secret", env != EnvLocal),
		SpotifyTokenSecret:  GetEnvWithConditionalPanic("SPOTIFY_TOKEN_SECRET", "f9726448847c4509f42a7e7dd3ea24e399f7fb57f3c9def4b4486ebe9f659b47", env != EnvLocal),
		Host:                host,
		StateCode:           GetEnvWithDefault("STATE_CODE", "state"),
		SpotifyClientId:     GetEnvWithPanic("SPOTIFY_ID"),
		SpotifyClientSecret: GetEnvWithPanic("SPOTIFY_SECRET"),
		DiscogsKey:          GetEnvWithPanic("DISCOGS_ID"),
		DiscogsSecret:       GetEnvWithPanic("DISCOGS_SECRET"),
		AppName:             GetEnvWithDefault("APP_NAME", "wax"),
		AppVersion:          GetEnvWithDefault("APP_VERSION", "0.0.0"),
		ContactEmail:        GetEnvWithDefault("CONTACT_EMAIL", "support@wax.com"),
	}
}

func GetEnvWithPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s not set", key))
	}
	return value
}

func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvWithConditionalPanic(key, defaultValue string, condition bool) string {
	if condition {
		return GetEnvWithPanic(key)
	}
	return GetEnvWithDefault(key, defaultValue)
}
