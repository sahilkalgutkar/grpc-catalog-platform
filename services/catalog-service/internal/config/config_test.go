package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skalgutkar/grpc-catalog-platform/services/catalog-service/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := config.Load()

	require.Equal(t, "9090", cfg.GRPCPort)
	require.Equal(t, "8080", cfg.HTTPPort)
	require.Equal(t, "localhost:9091", cfg.PricingServiceAddr)
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("GRPC_PORT", "19090")
	t.Setenv("HTTP_PORT", "18080")
	t.Setenv("PRICING_SERVICE_ADDR", "pricing:9999")

	cfg := config.Load()

	require.Equal(t, "19090", cfg.GRPCPort)
	require.Equal(t, "18080", cfg.HTTPPort)
	require.Equal(t, "pricing:9999", cfg.PricingServiceAddr)
}
