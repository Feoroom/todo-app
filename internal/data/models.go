package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Events EventModel
	Cards  CardModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Events: EventModel{DB: db},
		Cards:  CardModel{DB: db},
	}
}
