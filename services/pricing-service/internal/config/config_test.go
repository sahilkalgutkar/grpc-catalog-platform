package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/skalgutkar/grpc-catalog-platform/services/pricing-service/internal/config"
)

func TestLoad_Default(t *testing.T) {
	cfg := config.Load()

	require.Equal(t, "9091", cfg.GRPCPort)
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("GRPC_PORT", "19091")

	cfg := config.Load()

	require.Equal(t, "19091", cfg.GRPCPort)
}
