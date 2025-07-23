package data

import (
	"database/sql"
	"errors"
)

var (
	// ErrRecordNotFound is returned when book record doesn't exist in database.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when there is a data race.
	ErrEditConflict = errors.New("edit conflict")
)

// Models struct is a single container to hold all database models.
type Models struct {
	Books BookModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Books: BookModel{DB: db},
	}
}
