package auth

import (
	"errors"
	"net/http"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type AuthHandler struct {
	service AuthService
}

func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dtos.RegisterFormRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.Register(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrWeakPassword):
			httputil.ValidationError(w, err.Error(), nil)
		case errors.Is(err, ErrEmailTaken):
			httputil.Conflict(w, "email is already taken")
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginFormRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httputil.Unauthorized(w, err.Error())
		} else {
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req dtos.RefreshFormRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidRefreshToken) || errors.Is(err, ErrRefreshTokenExpired) {
			httputil.Unauthorized(w, err.Error())
		} else {
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	if err := h.service.Logout(r.Context(), userID); err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.NoContent(w)
}
