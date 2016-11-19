package api

import (
	"github.com/jackc/pgx"
)

type api struct {
	db *pgx.Conn
}

func New(db *pgx.Conn) *api {
	return &api{db: db}
}
