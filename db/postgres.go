package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/mproyyan/grpc-shipping-microservice/config"
)

type PostgreSQL struct {
	env config.Environment
}

func NewPostgreSQL(env config.Environment) *PostgreSQL {
	return &PostgreSQL{
		env: env,
	}
}

func (pq *PostgreSQL) Connect() (*sql.DB, error) {
	db, err := sql.Open(
		"postgres",
		fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			pq.env.DBUsername,
			pq.env.DBPassword,
			pq.env.DBHost,
			pq.env.DBPort,
			pq.env.DBName,
		),
	)

	if err != nil {
		return nil, err
	}

	err = db.Ping()
	return db, err
}
