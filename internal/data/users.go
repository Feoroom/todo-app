package data

import (
	"database/sql"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"password"`
	Activated bool      `json:"activated"`
	Version   int       `json:"version"`
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB sql.DB
}

func (p *password) Set(textPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(textPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &textPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(textPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(textPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	
	return true, nil

}
