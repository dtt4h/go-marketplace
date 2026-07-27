package users

import (
	"errors"
	"net/http"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type UserHandler struct {
	service UserService
}

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
			httputil.InternalError(w, err.Error())
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
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// CreateStore godoc
// @Summary      Register as seller (create store)
// @Description  Creates a store and upgrades user role to seller
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CreateStoreRequest  true  "Store data"
// @Success      201  {object}  dtos.StoreResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      409  {object}  httputil.ErrorResponse
// @Router       /users/me/store [post]
func (h *UserHandler) CreateStore(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.CreateStoreRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.CreateStore(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrStoreAlreadyExists):
			httputil.Conflict(w, err.Error())
		case errors.Is(err, ErrStoreNameRequired):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}