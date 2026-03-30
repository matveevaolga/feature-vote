package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/transport/middleware"
)

func TestContext(userID string) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, middleware.UserIDKey, userID)
}

func ExecuteRequest(req *http.Request, handler http.HandlerFunc) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

func ParseJSONResponse(t *testing.T, rr *httptest.ResponseRecorder, target interface{}) {
	if err := json.NewDecoder(rr.Body).Decode(target); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}
}

func CreateTestUser() *domain.User {
	return &domain.User{
		ID:       uuid.Must(uuid.NewV4()),
		Username: "testuser",
	}
}

func CreateTestGroup(ownerID uuid.UUID) *domain.Group {
	return &domain.Group{
		ID:        uuid.Must(uuid.NewV4()),
		Name:      "Test Group",
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
	}
}
