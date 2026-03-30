package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gofrs/uuid/v5"
	"github.com/matveevaolga/feature-vote/internal/domain"
	"github.com/matveevaolga/feature-vote/internal/service"
	"github.com/matveevaolga/feature-vote/internal/service/mocks"
	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupGroupTest(t *testing.T) (*service.GroupService, *mocks.MockGroupRepository, *service.UserService, *mocks.MockUserRepository, *GroupHandler) {
	mockGroupRepo := new(mocks.MockGroupRepository)
	mockUserRepo := new(mocks.MockUserRepository)
	mockJWTService := new(mocks.MockJWTService)

	userService := service.NewUserService(mockUserRepo, mockJWTService)
	groupService := service.NewGroupService(mockGroupRepo, mockUserRepo)

	handler := NewGroupHandler(groupService, userService)

	return groupService, mockGroupRepo, userService, mockUserRepo, handler
}

func TestGroupHandler_CreateGroup_Success(t *testing.T) {
	_, mockGroupRepo, _, mockUserRepo, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupName := "New Group"

	mockUserRepo.On("GetUserByID", mock.Anything, userID.String()).Return(&domain.User{ID: userID}, nil)
	mockGroupRepo.On("CreateGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
	mockGroupRepo.On("AddMember", mock.Anything, mock.AnythingOfType("*domain.GroupMember")).Return(nil)

	reqBody := dto.CreateGroupRequest{Name: groupName}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateGroup)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var response dto.GroupResponse
	ParseJSONResponse(t, rr, &response)
	assert.Equal(t, groupName, response.Name)
	assert.Equal(t, userID.String(), response.OwnerID)

	mockGroupRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestGroupHandler_CreateGroup_InvalidName(t *testing.T) {
	_, _, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	reqBody := dto.CreateGroupRequest{Name: "ab"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/groups", bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rr := ExecuteRequest(req, handler.CreateGroup)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGroupHandler_GetGroup_Success(t *testing.T) {
	_, mockGroupRepo, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())
	group := &domain.Group{
		ID:      groupID,
		Name:    "Test Group",
		OwnerID: userID,
	}

	mockGroupRepo.On("GetGroupByID", mock.Anything, groupID.String()).Return(group, nil)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID.String(), nil)
	req = req.WithContext(TestContext(userID.String()))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.GetGroup)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response dto.GroupResponse
	ParseJSONResponse(t, rr, &response)
	assert.Equal(t, group.Name, response.Name)
}

func TestGroupHandler_GetGroup_NotFound(t *testing.T) {
	_, mockGroupRepo, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetGroupByID", mock.Anything, groupID.String()).Return(nil, domain.ErrGroupNotFound)

	req := httptest.NewRequest(http.MethodGet, "/groups/"+groupID.String(), nil)
	req = req.WithContext(TestContext(userID.String()))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.GetGroup)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGroupHandler_UpdateGroup_Success(t *testing.T) {
	_, mockGroupRepo, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())
	newName := "Updated Group Name"

	group := &domain.Group{
		ID:      groupID,
		Name:    "Old Name",
		OwnerID: userID,
	}

	mockGroupRepo.On("GetGroupByID", mock.Anything, groupID.String()).Return(group, nil)
	mockGroupRepo.On("UpdateGroup", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)

	reqBody := dto.UpdateGroupRequest{Name: newName}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/groups/"+groupID.String(), bytes.NewReader(body))
	req = req.WithContext(TestContext(userID.String()))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.UpdateGroup)

	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Equal(t, newName, group.Name)
}

func TestGroupHandler_DeleteGroup_Success(t *testing.T) {
	_, mockGroupRepo, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	group := &domain.Group{
		ID:      groupID,
		Name:    "Test Group",
		OwnerID: userID,
	}

	mockGroupRepo.On("GetGroupByID", mock.Anything, groupID.String()).Return(group, nil)
	mockGroupRepo.On("DeleteGroup", mock.Anything, groupID.String()).Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/groups/"+groupID.String(), nil)
	req = req.WithContext(TestContext(userID.String()))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.DeleteGroup)

	assert.Equal(t, http.StatusNoContent, rr.Code)
}

func TestGroupHandler_InviteMember_Success(t *testing.T) {
	_, mockGroupRepo, _, mockUserRepo, handler := setupGroupTest(t)

	ownerID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())
	inviteeID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), ownerID.String()).Return(domain.RoleOwner, nil)
	mockUserRepo.On("GetUserByID", mock.Anything, inviteeID.String()).Return(&domain.User{ID: inviteeID}, nil)
	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), inviteeID.String()).Return(domain.Role(""), domain.ErrNotGroupMember)
	mockGroupRepo.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.Invitation")).Return(nil)

	reqBody := dto.InviteMemberRequest{UserID: inviteeID.String()}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID.String()+"/invite", bytes.NewReader(body))
	req = req.WithContext(TestContext(ownerID.String()))
	req.Header.Set("Content-Type", "application/json")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.InviteMember)

	assert.Equal(t, http.StatusCreated, rr.Code)
}

func TestGroupHandler_LeaveGroup_Success(t *testing.T) {
	_, mockGroupRepo, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleMember, nil)
	mockGroupRepo.On("RemoveMember", mock.Anything, groupID.String(), userID.String()).Return(nil)

	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID.String()+"/leave", nil)
	req = req.WithContext(TestContext(userID.String()))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.LeaveGroup)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestGroupHandler_LeaveGroup_OwnerCannotLeave(t *testing.T) {
	_, mockGroupRepo, _, _, handler := setupGroupTest(t)

	userID := uuid.Must(uuid.NewV4())
	groupID := uuid.Must(uuid.NewV4())

	mockGroupRepo.On("GetMemberRole", mock.Anything, groupID.String(), userID.String()).Return(domain.RoleOwner, nil)

	req := httptest.NewRequest(http.MethodPost, "/groups/"+groupID.String()+"/leave", nil)
	req = req.WithContext(TestContext(userID.String()))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", groupID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := ExecuteRequest(req, handler.LeaveGroup)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
