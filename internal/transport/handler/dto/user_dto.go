package dto

import "time"

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=5,max=50"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}
