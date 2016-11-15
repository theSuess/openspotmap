package api

import (
	"database/sql"
)

type api struct {
	db *sql.DB
}

func New(db *sql.DB) *api {
	return &api{db: db}
}
