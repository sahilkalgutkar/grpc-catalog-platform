package config

import "os"

type Config struct {
	GRPCPort           string
	HTTPPort           string
	PricingServiceAddr string
}

func Load() Config {
	return Config{
		GRPCPort:           getEnv("GRPC_PORT", "9090"),
		HTTPPort:           getEnv("HTTP_PORT", "8080"),
		PricingServiceAddr: getEnv("PRICING_SERVICE_ADDR", "localhost:9091"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
