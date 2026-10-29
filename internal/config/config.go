package config

import "os"

type Config struct {
    Port        string
    DatabaseURL string
    RedisURL    string
    LogLevel    string
}

func Load() Config {
    port := os.Getenv("PORT")
    if port=="" { port="8080" }
    return Config{
        Port: port,
        DatabaseURL: os.Getenv("DATABASE_URL"),
        RedisURL: os.Getenv("REDIS_URL"),
        LogLevel: os.Getenv("LOG_LEVEL"),
    }
}
