package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (u *User) Validate() error {
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return ErrInvalidUsername
	}
	if len(u.Email) == 0 {
		return ErrInvalidEmail
	}
	if len(u.PasswordHash) == 0 {
		return ErrInvalidPassword
	}
	return nil
}
