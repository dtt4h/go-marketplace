package products

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.ListCategories(r.Context())
	if err != nil {
		httputil.InternalError(w, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	storeID := parseOptionalInt64(r.URL.Query().Get("store_id"))
	categoryID := parseOptionalInt64(r.URL.Query().Get("category_id"))
	minPrice := parseOptionalString(r.URL.Query().Get("min_price"))
	maxPrice := parseOptionalString(r.URL.Query().Get("max_price"))
	search := parseOptionalString(r.URL.Query().Get("search"))
	sort := r.URL.Query().Get("sort")

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListProducts(r.Context(), storeID, categoryID, minPrice, maxPrice, search, sort, page, limit)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidSort):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
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

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid product id", nil)
		return
	}

	resp, err := h.service.GetProduct(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			httputil.NotFound(w, err.Error())
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.CreateProductRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.CreateProduct(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrStoreNotFound):
			httputil.Forbidden(w, err.Error())
		case errors.Is(err, ErrInvalidPrice), errors.Is(err, ErrInvalidStock),
			errors.Is(err, ErrTitleRequired), errors.Is(err, ErrTitleTooLong):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid product id", nil)
		return
	}

	var req dtos.UpdateProductRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.UpdateProduct(r.Context(), userID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrForbidden):
			httputil.Forbidden(w, err.Error())
		case errors.Is(err, ErrInvalidPrice), errors.Is(err, ErrInvalidStock),
			errors.Is(err, ErrTitleRequired), errors.Is(err, ErrTitleTooLong):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid product id", nil)
		return
	}

	err = h.service.DeleteProduct(r.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrForbidden):
			httputil.Forbidden(w, err.Error())
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.NoContent(w)
}

func (h *ProductHandler) ModerateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid product id", nil)
		return
	}

	var req dtos.ModerateProductRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.ModerateProduct(r.Context(), id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrInvalidStatus):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

func parseOptionalInt64(s string) *int64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseOptionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
