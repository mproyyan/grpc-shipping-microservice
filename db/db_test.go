package db

import (
	"testing"

	"github.com/mproyyan/grpc-shipping-microservice/config"
	"github.com/stretchr/testify/require"
)

func TestPostgresConnection(t *testing.T) {
	env := config.Environment{
		DBUsername: "postgres",
		DBPassword: "ligmaballs",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "grpc_shipping",
	}

	db, err := NewPostgreSQL(env).Connect()
	require.NoError(t, err)
	require.NotNil(t, db)
}
