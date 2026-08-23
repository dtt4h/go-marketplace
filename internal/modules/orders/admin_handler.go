package orders

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// AdminOrderHandler handles admin HTTP requests for order operations.
type AdminOrderHandler struct {
	service AdminOrderService
}

// NewAdminOrderHandler creates a new AdminOrderHandler.
func NewAdminOrderHandler(service AdminOrderService) *AdminOrderHandler {
	return &AdminOrderHandler{service: service}
}

// ListAllOrders godoc
// @Summary      List all orders (admin)
// @Description  Returns all orders across the platform (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default 1)"
// @Param        limit  query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Router       /admin/orders [get]
func (h *AdminOrderHandler) ListAllOrders(w http.ResponseWriter, r *http.Request) {
	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListAllOrders(r.Context(), page, limit)
	if err != nil {
		httputil.InternalError(w, r, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, httputil.PaginatedResponse{
		Items: items,
		Total: int(total),
		Page:  page,
		Limit: limit,
	})
}

// ListOrdersByStatus godoc
// @Summary      List orders by status (admin)
// @Description  Returns orders filtered by status (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        status  query  string  true  "Order status"
// @Param        page    query  int     false  "Page number (default 1)"
// @Param        limit   query  int     false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Router       /admin/orders [get]
func (h *AdminOrderHandler) ListOrdersByStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		httputil.ValidationError(w, "status query parameter is required", nil)
		return
	}

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListOrdersByStatus(r.Context(), status, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidStatus):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, httputil.PaginatedResponse{
		Items: items,
		Total: int(total),
		Page:  page,
		Limit: limit,
	})
}

// ListOrdersByUser godoc
// @Summary      List orders by user (admin)
// @Description  Returns all orders for a specific user (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        user_id  query  int  true  "User ID"
// @Param        page     query  int  false  "Page number (default 1)"
// @Param        limit    query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Router       /admin/orders [get]
func (h *AdminOrderHandler) ListOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		httputil.ValidationError(w, "user_id query parameter is required", nil)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid user_id", nil)
		return
	}

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListOrdersByUser(r.Context(), userID, page, limit)
	if err != nil {
		httputil.InternalError(w, r, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, httputil.PaginatedResponse{
		Items: items,
		Total: int(total),
		Page:  page,
		Limit: limit,
	})
}

// GetOrderDetail godoc
// @Summary      Get order detail (admin)
// @Description  Returns full order detail with items (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  int  true  "Order ID"
// @Success      200  {object}  dtos.OrderResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /admin/orders/{id} [get]
func (h *AdminOrderHandler) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid order id", nil)
		return
	}

	resp, err := h.service.GetOrderDetail(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			httputil.NotFound(w, err.Error())
		default:
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}
