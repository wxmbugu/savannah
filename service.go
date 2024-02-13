package savannah

import (
	"database/sql"
	// "log"
	// "sync"
	// "time"
)

type Service struct {
	service database
}

func NewService(conn *sql.DB) Service {
	db := Newdb(conn)
	return Service{
		service: db,
	}
}
