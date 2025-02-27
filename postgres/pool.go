package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
	*pgxpool.Pool
}

func NewPool(config *Config) *Pool {
	pool, err := pgxpool.New(context.Background(), config.ConnectionString())
	if err != nil {
		panic(err)
	}

	return &Pool{Pool: pool}
}
