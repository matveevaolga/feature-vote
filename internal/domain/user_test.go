package domain

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
)

func TestUser_Validate(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		wantErr  bool
	}{
		{"valid user", "john_doe", "john@example.com", false},
		{"too short username", "jo", "john@example.com", true},
		{"too long username", "thisusernameiswaytoolongandshouldnotbeallowedxxxxxxxxxxxxxxxx", "john@example.com", true},
		{"empty username", "", "john@example.com", true},
		{"empty email", "john_doe", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:           uuid.Must(uuid.NewV4()),
				Username:     tt.username,
				Email:        tt.email,
				PasswordHash: "hashedpassword123",
				CreatedAt:    time.Now(),
			}
			err := user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
