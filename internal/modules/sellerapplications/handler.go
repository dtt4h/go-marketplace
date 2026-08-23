package sellerapplications

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type SellerApplicationHandler struct {
	service SellerApplicationService
}

func NewSellerApplicationHandler(service SellerApplicationService) *SellerApplicationHandler {
	return &SellerApplicationHandler{service: service}
}

// Create godoc
// @Summary      Submit seller application
// @Description  Submits a seller application for admin review
// @Tags         seller-applications
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body  dtos.CreateSellerApplicationRequest  true  "Application data"
// @Success      201  {object}  dtos.SellerApplicationResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      409  {object}  httputil.ErrorResponse
// @Router       /seller-applications [post]
func (h *SellerApplicationHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.CreateSellerApplicationRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	app, err := h.service.Create(r.Context(), userID, req.StoreName, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, ErrAlreadyPending):
			httputil.Conflict(w, err.Error())
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, dtos.ToSellerApplicationResponse(app))
}

// ListMine godoc
// @Summary      List my seller applications
// @Tags         seller-applications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}  dtos.SellerApplicationResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /seller-applications/me [get]
func (h *SellerApplicationHandler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	apps, err := h.service.ListMine(r.Context(), userID)
	if err != nil {
		httputil.InternalError(w, r, err.Error())
		return
	}

	resp := make([]dtos.SellerApplicationResponse, 0, len(apps))
	for _, app := range apps {
		resp = append(resp, dtos.ToSellerApplicationResponse(app))
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// ListPending godoc
// @Summary      List pending seller applications (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default 1)"
// @Param        limit  query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Router       /admin/seller-applications [get]
func (h *SellerApplicationHandler) ListPending(w http.ResponseWriter, r *http.Request) {
	page, limit := httputil.ParsePagination(r)

	apps, total, err := h.service.ListPending(r.Context(), page, limit)
	if err != nil {
		httputil.InternalError(w, r, err.Error())
		return
	}

	resp := make([]dtos.SellerApplicationResponse, 0, len(apps))
	for _, app := range apps {
		resp = append(resp, dtos.ToSellerApplicationResponse(app))
	}

	httputil.JSON(w, http.StatusOK, httputil.PaginatedResponse{
		Items: resp,
		Total: int(total),
		Page:  page,
		Limit: limit,
	})
}

// UpdateStatus godoc
// @Summary      Approve or reject seller application (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path  int                                  true  "Application ID"
// @Param        request  body  dtos.UpdateApplicationStatusRequest  true  "New status"
// @Success      200  {object}  dtos.SellerApplicationResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /admin/seller-applications/{id} [patch]
func (h *SellerApplicationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid application id", nil)
		return
	}

	var req dtos.UpdateApplicationStatusRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	var app db.SellerApplication
	switch req.Status {
	case "approved":
		app, err = h.service.Approve(r.Context(), id)
	case "rejected":
		app, err = h.service.Reject(r.Context(), id, req.Reason)
	default:
		httputil.ValidationError(w, ErrInvalidStatus.Error(), nil)
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrAlreadyProcessed):
			httputil.Conflict(w, err.Error())
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, dtos.ToSellerApplicationResponse(app))
}
