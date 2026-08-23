package products

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/internal/cache"
	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// ProductHandler handles HTTP requests for product operations.
type ProductHandler struct {
	service ProductService
	cache   *cache.CatalogCache
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(service ProductService, cache *cache.CatalogCache) *ProductHandler {
	return &ProductHandler{service: service, cache: cache}
}

// ListCategories godoc
// @Summary      List product categories
// @Description  Returns a tree of categories and subcategories
// @Tags         products
// @Produce      json
// @Success      200  {array}  dtos.CategoryResponse
// @Router       /products/categories [get]
func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	resp, err := h.service.ListCategories(r.Context())
	if err != nil {
		httputil.InternalError(w, r, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// ListProducts godoc
// @Summary      List products with filters
// @Description  Returns paginated list of active products
// @Tags         products
// @Produce      json
// @Param        page         query  int     false  "Page number (default 1)"
// @Param        limit        query  int     false  "Items per page (default 20, max 100)"
// @Param        store_id     query  int     false  "Filter by store ID"
// @Param        category_id  query  int     false  "Filter by category ID"
// @Param        min_price    query  string  false  "Minimum price"
// @Param        max_price    query  string  false  "Maximum price"
// @Param        search       query  string  false  "Search by title"
// @Param        sort         query  string  false  "Sort: price_asc, price_desc, created_desc (default)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Router       /products [get]
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

// GetProduct godoc
// @Summary      Get product by ID
// @Tags         products
// @Produce      json
// @Param        id   path  int  true  "Product ID"
// @Success      200  {object}  dtos.ProductResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/{id} [get]
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
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// CreateProduct godoc
// @Summary      Create a product
// @Description  Creates a product with status 'pending' (requires seller role)
// @Tags         products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CreateProductRequest  true  "Product data"
// @Success      201  {object}  dtos.ProductResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Router       /products [post]
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

	// Validate input
	errs := dtos.ValidateNonEmptyString(req.Title, "title")
	errs.AddMap(dtos.ValidatePrice(req.Price, "price"))
	errs.AddMap(dtos.ValidatePositiveInt(req.Stock, "stock"))
	if !errs.IsEmpty() {
		httputil.ValidationError(w, "validation failed", errs)
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
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

// UpdateProduct godoc
// @Summary      Update a product
// @Description  Updates product fields (only store owner)
// @Tags         products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path  int                         true  "Product ID"
// @Param        request  body  dtos.UpdateProductRequest   true  "Fields to update"
// @Success      200  {object}  dtos.ProductResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/{id} [patch]
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
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// DeleteProduct godoc
// @Summary      Delete a product
// @Description  Deletes a product (only store owner)
// @Tags         products
// @Security     BearerAuth
// @Param        id   path  int  true  "Product ID"
// @Success      204
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/{id} [delete]
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
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.NoContent(w)
}

// ModerateProduct godoc
// @Summary      Moderate a product
// @Description  Approve, reject or archive a product (admin only)
// @Tags         products
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id       path  int                         true  "Product ID"
// @Param        request  body  dtos.ModerateProductRequest  true  "Moderation decision"
// @Success      200  {object}  dtos.ProductResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/{id}/moderate [patch]
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
			httputil.InternalError(w, r, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// ListProductsByStore godoc
// @Summary      List products by store ID
// @Description  Returns paginated list of active products for a specific store
// @Tags         products
// @Produce      json
// @Param        id   path  int  true  "Store ID"
// @Param        page   query  int  false  "Page number (default 1)"
// @Param        limit  query  int  false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Router       /stores/{id}/products [get]
func (h *ProductHandler) ListProductsByStore(w http.ResponseWriter, r *http.Request) {
	storeID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid store id", nil)
		return
	}

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListProductsByStore(r.Context(), storeID, page, limit)
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

// ListProductsByStatus godoc
// @Summary      List products by moderation status (admin only)
// @Description  Returns paginated list of products filtered by status (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Param        status  query  string  true  "Product status: pending, active, rejected, archived"
// @Param        page    query  int     false  "Page number (default 1)"
// @Param        limit   query  int     false  "Items per page (default 20, max 100)"
// @Success      200  {object}  httputil.PaginatedResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Router       /admin/products [get]
func (h *ProductHandler) ListProductsByStatus(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		httputil.ValidationError(w, "status query parameter is required", nil)
		return
	}

	page, limit := httputil.ParsePagination(r)

	items, total, err := h.service.ListProductsByStatus(r.Context(), status, page, limit)
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
