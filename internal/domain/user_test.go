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
		wantErr  bool
	}{
		{"valid username", "john_doe", false},
		{"too short", "jo", true},
		{"too long", "thisusernameiswaytoolongandshouldnotbeallowedewewrq", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				ID:        uuid.Must(uuid.NewV4()),
				Username:  tt.username,
				CreatedAt: time.Now(),
			}
			err := user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
