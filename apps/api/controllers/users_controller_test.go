package controllers

import (
	"4ks/apps/api/dtos"
	usersvc "4ks/apps/api/services/user"
	models "4ks/libs/go/models"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func performUserControllerRequest(t *testing.T, handler gin.HandlerFunc, method string, target string, body []byte, setup func(*gin.Context)) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = req
	if setup != nil {
		setup(ctx)
	}

	handler(ctx)
	return rec
}

func TestUserControllerCreateUser(t *testing.T) {
	t.Parallel()

	t.Run("bind failure returns bad request", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{})
		rec := performUserControllerRequest(t, controller.CreateUser, http.MethodPost, "/api/user", []byte("{"), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
			ctx.Set("email", "user@example.com")
		})

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("conflict-like validation errors return bad request", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			createUserFn: func(_ context.Context, userID string, userEmail string, payload *dtos.CreateUser) (*models.User, error) {
				if userID != "user-1" || userEmail != "user@example.com" || payload.Username != "chef-user" {
					t.Fatalf("unexpected create inputs: %q %q %+v", userID, userEmail, payload)
				}
				return nil, usersvc.ErrUsernameInUse
			},
		})

		rec := performUserControllerRequest(t, controller.CreateUser, http.MethodPost, "/api/user", []byte(`{"username":"chef-user","displayName":"Chef User"}`), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
			ctx.Set("email", "user@example.com")
		})

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("success returns created user", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			createUserFn: func(_ context.Context, userID string, userEmail string, payload *dtos.CreateUser) (*models.User, error) {
				return &models.User{
					ID:           userID,
					Username:     payload.Username,
					DisplayName:  payload.DisplayName,
					EmailAddress: userEmail,
				}, nil
			},
		})

		rec := performUserControllerRequest(t, controller.CreateUser, http.MethodPost, "/api/user", []byte(`{"username":"chef-user","displayName":"Chef User"}`), func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
			ctx.Set("email", "user@example.com")
		})

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var user models.User
		if err := json.NewDecoder(rec.Body).Decode(&user); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if user.Username != "chef-user" || user.EmailAddress != "user@example.com" {
			t.Fatalf("unexpected user payload: %+v", user)
		}
	})
}

func TestUserControllerHeadAuthenticatedUser(t *testing.T) {
	t.Parallel()

	t.Run("missing user returns no content", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, usersvc.ErrUserNotFound
			},
		})

		rec := performUserControllerRequest(t, controller.HeadAuthenticatedUser, http.MethodHead, "/api/user", nil, func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d", rec.Code)
		}
	})

	t.Run("lookup errors return internal server error", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			getUserByIDFn: func(context.Context, string) (*models.User, error) {
				return nil, errors.New("boom")
			},
		})

		rec := performUserControllerRequest(t, controller.HeadAuthenticatedUser, http.MethodHead, "/api/user", nil, func(ctx *gin.Context) {
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500, got %d", rec.Code)
		}
	})
}

func TestUserControllerRemoveUserEvent(t *testing.T) {
	t.Parallel()

	t.Run("invalid event id returns bad request", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{})
		rec := performUserControllerRequest(t, controller.RemoveUserEvent, http.MethodDelete, "/api/user/events/bad", nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: "bad-uuid"}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("missing event returns not found", func(t *testing.T) {
		t.Parallel()

		eventID := uuid.New()
		controller := NewUserController(stubUserService{
			removeUserEventFn: func(_ context.Context, userID string, got uuid.UUID) error {
				if userID != "user-1" || got != eventID {
					t.Fatalf("unexpected remove inputs: %q %s", userID, got)
				}
				return usersvc.ErrUserEventNotFound
			},
		})

		rec := performUserControllerRequest(t, controller.RemoveUserEvent, http.MethodDelete, "/api/user/events/"+eventID.String(), nil, func(ctx *gin.Context) {
			ctx.Params = gin.Params{{Key: "id", Value: eventID.String()}}
			ctx.Set("id", "user-1")
		})

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}

func TestUserControllerTestUsername(t *testing.T) {
	t.Parallel()

	t.Run("empty username is rejected", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{})
		rec := performUserControllerRequest(t, controller.TestUsername, http.MethodPost, "/api/users/username", []byte(`{"username":""}`), nil)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", rec.Code)
		}
	})

	t.Run("reserved usernames return reserved message", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			testNameFn: func(context.Context, string) error { return usersvc.ErrReservedWord },
		})

		rec := performUserControllerRequest(t, controller.TestUsername, http.MethodPost, "/api/users/username", []byte(`{"username":"admin"}`), nil)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp dtos.TestUsernameResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Message != "reserved" || resp.Available || resp.Valid {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("in-use usernames stay valid but unavailable", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			testNameFn: func(context.Context, string) error { return usersvc.ErrUsernameInUse },
		})

		rec := performUserControllerRequest(t, controller.TestUsername, http.MethodPost, "/api/users/username", []byte(`{"username":"chef-user"}`), nil)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp dtos.TestUsernameResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !resp.Valid || resp.Available || resp.Message != "in use" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})

	t.Run("available usernames return valid and available", func(t *testing.T) {
		t.Parallel()

		controller := NewUserController(stubUserService{
			testNameFn: func(context.Context, string) error { return nil },
		})

		rec := performUserControllerRequest(t, controller.TestUsername, http.MethodPost, "/api/users/username", []byte(`{"username":"chef-user"}`), nil)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}

		var resp dtos.TestUsernameResponse
		if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !resp.Valid || !resp.Available || resp.Message != "" {
			t.Fatalf("unexpected response: %+v", resp)
		}
	})
}
