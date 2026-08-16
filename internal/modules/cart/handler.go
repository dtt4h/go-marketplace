package cart

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type CartHandler struct {
	service CartService
}

func NewCartHandler(service CartService) *CartHandler {
	return &CartHandler{service: service}
}

// GetCart godoc
// @Summary      Get current user's cart
// @Description  Returns all items in the current user's shopping cart
// @Tags         cart
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.CartResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	resp, err := h.service.GetCart(r.Context(), userID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, resp)
}

// AddItem godoc
// @Summary      Add item to cart
// @Description  Adds a product to the current user's cart (or updates quantity if already present)
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.AddCartItemRequest  true  "Product ID and quantity"
// @Success      200  {object}  dtos.CartResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /cart/items [post]
func (h *CartHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.AddCartItemRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.AddItem(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidQuantity):
			httputil.ValidationError(w, err.Error(), nil)
		case errors.Is(err, ErrProductNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrInsufficientStock):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}
	httputil.JSON(w, http.StatusOK, resp)
}

// UpdateItem godoc
// @Summary      Update cart item quantity
// @Description  Updates the quantity of a specific cart item
// @Tags         cart
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        itemID   path  int                      true  "Cart item ID"
// @Param        request  body  dtos.UpdateCartItemRequest true  "New quantity"
// @Success      200  {object}  dtos.CartResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /cart/items/{itemID} [patch]
func (h *CartHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid item id", nil)
		return
	}

	var req dtos.UpdateCartItemRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.UpdateItem(r.Context(), userID, itemID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidQuantity):
			httputil.ValidationError(w, err.Error(), nil)
		case errors.Is(err, ErrProductNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrInsufficientStock):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}
	httputil.JSON(w, http.StatusOK, resp)
}

// RemoveItem godoc
// @Summary      Remove item from cart
// @Description  Removes a specific item from the current user's cart
// @Tags         cart
// @Security     BearerAuth
// @Produce      json
// @Param        itemID  path  int               true  "Cart item ID"
// @Success      200  {object}  dtos.CartResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /cart/items/{itemID} [delete]
func (h *CartHandler) RemoveItem(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid item id", nil)
		return
	}

	resp, err := h.service.RemoveItem(r.Context(), userID, itemID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, resp)
}

// Clear godoc
// @Summary      Clear cart
// @Description  Removes all items from the current user's cart
// @Tags         cart
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.CartResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Router       /cart [delete]
func (h *CartHandler) Clear(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	resp, err := h.service.Clear(r.Context(), userID)
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}
	httputil.JSON(w, http.StatusOK, resp)
}
