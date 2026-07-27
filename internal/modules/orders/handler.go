package orders

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type OrderHandler struct {
	service OrderService
}

func NewOrderHandler(service OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

// CreateOrder godoc
// @Summary      Create an order
// @Description  Creates an order from provided items, decrements stock
// @Tags         orders
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CreateOrderRequest  true  "Order data"
// @Success      201  {object}  dtos.OrderResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.CreateOrderRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.CreateOrder(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderEmpty):
			httputil.ValidationError(w, err.Error(), nil)
		case errors.Is(err, ErrInsufficientStock):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

// GetOrder godoc
// @Summary      Get order by ID
// @Description  Returns order details (buyer or seller of items in order)
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  int  true  "Order ID"
// @Success      200  {object}  dtos.OrderResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /orders/me/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid order id", nil)
		return
	}

	resp, err := h.service.GetOrder(r.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrForbidden):
			httputil.Forbidden(w, err.Error())
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// ListOrdersByUser godoc
// @Summary      List current user's orders
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default 1)"
// @Param        limit  query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /orders/me [get]
func (h *OrderHandler) ListOrdersByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListOrdersByUser(r.Context(), userID, page, limit)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, httputil.PaginatedResponse{
		Items: items,
		Total: int(total),
		Page:  page,
		Limit: limit,
	})
}

// ListOrdersBySeller godoc
// @Summary      List orders for seller's products
// @Description  Returns orders containing products from seller's store (seller role)
// @Tags         orders
// @Security     BearerAuth
// @Produce      json
// @Param        page   query  int  false  "Page number (default 1)"
// @Param        limit  query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Router       /orders/seller [get]
func (h *OrderHandler) ListOrdersBySeller(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListOrdersBySeller(r.Context(), userID, page, limit)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, httputil.PaginatedResponse{
		Items: items,
		Total: int(total),
		Page:  page,
		Limit: limit,
	})
}

// UpdateOrderStatus godoc
// @Summary      Update order status
// @Description  Transitions order status (seller or buyer)
// @Tags         orders
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path  int                            true  "Order ID"
// @Param        request  body  dtos.UpdateOrderStatusRequest  true  "New status"
// @Success      200  {object}  dtos.OrderResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid order id", nil)
		return
	}

	var req dtos.UpdateOrderStatusRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.UpdateOrderStatus(r.Context(), userID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrForbidden):
			httputil.Forbidden(w, err.Error())
		case errors.Is(err, ErrInvalidStatus):
			httputil.ValidationError(w, err.Error(), nil)
		case errors.Is(err, ErrInvalidStatusTrans):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}