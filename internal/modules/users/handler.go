package users

import (
	"errors"
	"net/http"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// UserHandler handles HTTP requests for user profile operations.
type UserHandler struct {
	service UserService
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service UserService) *UserHandler {
	return &UserHandler{service: service}
}

// GetProfile godoc
// @Summary      Get current user profile
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.ProfileResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /users/me [get]
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	resp, err := h.service.GetProfile(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			httputil.NotFound(w, err.Error())
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// UpdateProfile godoc
// @Summary      Update current user profile
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.UpdateProfileRequest  true  "Profile fields to update"
// @Success      200  {object}  dtos.ProfileResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Failure      409  {object}  httputil.ErrorResponse
// @Router       /users/me [patch]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.UpdateProfileRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.UpdateProfile(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrUsernameTaken):
			httputil.Conflict(w, err.Error())
		case errors.Is(err, ErrUsernameRequired):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// GetStore godoc
// @Summary      Get current user's store
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.StoreResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /users/me/store [get]
func (h *UserHandler) GetStore(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	resp, err := h.service.GetStore(r.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrStoreNotFound):
			httputil.NotFound(w, err.Error())
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// UpdateStore godoc
// @Summary      Update store
// @Description  Updates store details (name, description, logo)
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.UpdateStoreRequest  true  "Store fields to update"
// @Success      200  {object}  dtos.StoreResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /users/me/store [patch]
func (h *UserHandler) UpdateStore(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.UpdateStoreRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.UpdateStore(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrStoreNotFound):
			httputil.NotFound(w, err.Error())
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}