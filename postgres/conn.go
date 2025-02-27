package postgres

import (
	"context"
	"github.com/jackc/pgx/v5"
)

type Connection struct {
	*pgx.Conn
}

func NewConnection(config *Config) *Connection {
	conn, err := pgx.Connect(context.Background(), config.ConnectionString())
	if err != nil {
		panic(err)
	}
	return &Connection{Conn: conn}
}
