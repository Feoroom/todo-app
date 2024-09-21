package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
	"library/internal/validation"
	"time"
)

var (
	ErrDuplicateEmail         = errors.New("duplicate email")
	emailUniqueConstraintName = `users_email_key`
)

type User struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  password  `json:"-"`
	Activated bool      `json:"activated"`
	Version   int       `json:"-"`
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

var AnonymousUser = &User{}

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

func ValidateEmail(v *validation.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validation.Matches(email, validation.EmailRX), "email", "must be a valid email address")
}

func ValidatePlaintextPassword(v *validation.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must be not more than 72 bytes long")
}

func ValidateUser(v *validation.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must be not more than 500 bytes long")

	ValidateEmail(v, user.Email)

	if user.Password.plaintext != nil {
		ValidatePlaintextPassword(v, *user.Password.plaintext)
	}

	if user.Password.hash == nil {
		panic("missing password hash")
	}
}

func (u UserModel) Insert(user *User) error {
	q := `
		insert into users (name, email, password_hash, activated)
		values ($1, $2, $3, $4)
		returning id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, q, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)

	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch {
		case pqErr.Constraint == emailUniqueConstraintName:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	q := `
		select id, created_at, name, email, password_hash, activated, version from users
		where email=$1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, q, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u UserModel) Update(user *User) error {
	q := `
		update users
set name=$1, email=$2, password_hash=$3, activated=$4, version = version + 1
		where id = $5 and version=$6
		returning version
	`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.hash,
		user.Activated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, q, args...).Scan(&user.Version)
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch {
		case pqErr.Constraint == emailUniqueConstraintName:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (u UserModel) GetForToken(scope, tokenPlainText string) (*User, error) {

	hash := sha256.Sum256([]byte(tokenPlainText))

	q := `select users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
		from users
		inner join tokens on users.id = tokens.user_id
		where tokens.hash = $1
		and tokens.scope = $2
		and tokens.expiry > $3`

	args := []interface{}{
		hash[:], scope, time.Now(),
	}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, q, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}
