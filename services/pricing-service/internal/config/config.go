package config

import "os"

type Config struct {
	GRPCPort string
}

func Load() Config {
	return Config{
		GRPCPort: getEnv("GRPC_PORT", "9091"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
