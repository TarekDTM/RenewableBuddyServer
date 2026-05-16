package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/TarekDTM/EnergyOptimizer/internal/validator"
	"golang.org/x/crypto/bcrypt"
)

type UsersModel struct {
	DB *sql.DB
}

var AnonymousUser = &User{}

type User struct {
	ID         int64    `json:"id"`
	Created_at string   `json:"created_at"`
	Name       string   `json:"name"`
	Email      string   `json:"email"`
	Password   Password `json:"-"`
	Activated  bool     `json:"activated"`
	Version    int      `json:"version"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser

}

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

func (m UsersModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id,name,email,password_hash,activated 
	FROM users 
	WHERE email = $1  `
	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated)
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
func (m UsersModel) Insert(user *User) error {
	query := `INSERT INTO users(name, email, password_hash, activated)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version
	`
	args := []interface{}{user.Name, user.Email, user.Password.hash, user.Activated}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.Created_at, &user.Version)

	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
func (m UsersModel) Update(user *User) error {
	query := `
	UPDATE users
	SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
	WHERE id = $5 AND version = $6
	RETURNING version`

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

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "user_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
func (m UsersModel) GetForToken(tokenScope string, tokenPlainText string) (*User, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlainText))
	query := `SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var user User

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Created_at,
		&user.Name,
		&user.Email,
		&user.Password.hash,
		&user.Activated,
		&user.Version)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil

}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email form")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "password must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 500, "name", "must not be more than 500 bytes long")
	ValidateEmail(v, user.Email)
	if user.Password.plainText != nil {
		ValidatePasswordPlaintext(v, *user.Password.plainText)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user ")
	}

}

type Password struct {
	plainText *string
	hash      []byte
}

func (p *Password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}
	p.plainText = &plainTextPassword
	p.hash = hash
	return nil

}

func (p Password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
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
