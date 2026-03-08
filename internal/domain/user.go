package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

func (u *User) Validate() error {
	if len(u.Username) < 5 || len(u.Username) > 50 {
		return ErrInvalidUsername
	}
	return nil
}
