package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// AuthHandler handles HTTP requests for authentication.
type AuthHandler struct {
	service AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(service AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a buyer account and returns access + refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.RegisterRequest  true  "Registration data"
// @Success      201  {object}  dtos.AuthResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      409  {object}  httputil.ErrorResponse
// @Router       /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dtos.RegisterRequest
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
			httputil.ErrorWithDetails(w, http.StatusConflict, "CONFLICT", err.Error(), map[string]string{"field": "email"})
		case errors.Is(err, ErrUsernameTaken):
			httputil.ErrorWithDetails(w, http.StatusConflict, "CONFLICT", err.Error(), map[string]string{"field": "username"})
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	h.setRefreshCookie(w, resp.RefreshToken)
	httputil.JSON(w, http.StatusCreated, resp)
}

// Login godoc
// @Summary      Login
// @Description  Authenticates user and returns access + refresh tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.LoginRequest  true  "Login credentials"
// @Success      200  {object}  dtos.AuthResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dtos.LoginRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidCredentials):
			httputil.Unauthorized(w, err.Error())
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	h.setRefreshCookie(w, resp.RefreshToken)
	httputil.JSON(w, http.StatusOK, resp)
}

// Refresh godoc
// @Summary      Refresh tokens
// @Description  Rotates refresh token from cookie and returns new access token
// @Tags         auth
// @Produce      json
// @Success      200  {object}  dtos.RefreshResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		httputil.Unauthorized(w, "missing refresh token")
		return
	}

	resp, err := h.service.Refresh(r.Context(), refreshToken.Value)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidRefreshToken), errors.Is(err, ErrRefreshTokenExpired):
			h.clearRefreshCookie(w)
			httputil.Unauthorized(w, err.Error())
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	h.setRefreshCookie(w, resp.RefreshToken)
	httputil.JSON(w, http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout
// @Description  Deletes all refresh tokens for the current user
// @Tags         auth
// @Security     BearerAuth
// @Success      204
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /auth/logout [post]
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

	h.clearRefreshCookie(w)
	httputil.NoContent(w)
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(168 * time.Hour.Seconds()),
	})
}

func (h *AuthHandler) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}