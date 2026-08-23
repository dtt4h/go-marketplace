package products

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// --- Mock for handler tests ---

type mockProductServiceForHandler struct {
	listCategories       func(ctx context.Context) ([]dtos.CategoryResponse, error)
	listProducts         func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]dtos.ProductListItem, int64, error)
	listProductsByStore  func(ctx context.Context, storeID int64, page, limit int) ([]dtos.ProductListItem, int64, error)
	getProduct           func(ctx context.Context, id int64) (dtos.ProductResponse, error)
	createProduct        func(ctx context.Context, userID int64, req dtos.CreateProductRequest) (dtos.ProductResponse, error)
	updateProduct        func(ctx context.Context, userID, id int64, req dtos.UpdateProductRequest) (dtos.ProductResponse, error)
	deleteProduct        func(ctx context.Context, userID, id int64) error
	moderateProduct      func(ctx context.Context, id int64, req dtos.ModerateProductRequest) (dtos.ProductResponse, error)
	listProductsByStatus func(ctx context.Context, status string, page, limit int) ([]dtos.ProductListItem, int64, error)
}

func (m *mockProductServiceForHandler) ListCategories(ctx context.Context) ([]dtos.CategoryResponse, error) {
	return m.listCategories(ctx)
}

func (m *mockProductServiceForHandler) ListProducts(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]dtos.ProductListItem, int64, error) {
	return m.listProducts(ctx, storeID, categoryID, minPrice, maxPrice, search, sort, page, limit)
}

func (m *mockProductServiceForHandler) ListProductsByStore(ctx context.Context, storeID int64, page, limit int) ([]dtos.ProductListItem, int64, error) {
	return m.listProductsByStore(ctx, storeID, page, limit)
}

func (m *mockProductServiceForHandler) GetProduct(ctx context.Context, id int64) (dtos.ProductResponse, error) {
	return m.getProduct(ctx, id)
}

func (m *mockProductServiceForHandler) CreateProduct(ctx context.Context, userID int64, req dtos.CreateProductRequest) (dtos.ProductResponse, error) {
	return m.createProduct(ctx, userID, req)
}

func (m *mockProductServiceForHandler) UpdateProduct(ctx context.Context, userID, id int64, req dtos.UpdateProductRequest) (dtos.ProductResponse, error) {
	return m.updateProduct(ctx, userID, id, req)
}

func (m *mockProductServiceForHandler) DeleteProduct(ctx context.Context, userID, id int64) error {
	return m.deleteProduct(ctx, userID, id)
}

func (m *mockProductServiceForHandler) ModerateProduct(ctx context.Context, id int64, req dtos.ModerateProductRequest) (dtos.ProductResponse, error) {
	return m.moderateProduct(ctx, id, req)
}

func (m *mockProductServiceForHandler) ListProductsByStatus(ctx context.Context, status string, page, limit int) ([]dtos.ProductListItem, int64, error) {
	return m.listProductsByStatus(ctx, status, page, limit)
}

// mockAuthMiddleware simulates JWT auth middleware using the same context key as mw.UserIDFromCtx
func mockAuthMiddleware(userID int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, mw.UserIDKey, userID)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// newTestRouter creates a chi router with auth middleware and a single route
func newTestRouter(route string, handler http.HandlerFunc, userID int64) *chi.Mux {
	r := chi.NewRouter()
	if userID > 0 {
		r.Use(mockAuthMiddleware(userID))
	}
	r.Handle(route, handler)
	return r
}

func TestHandlerListCategories(t *testing.T) {
	tests := []struct {
		name       string
		mockResp   []dtos.CategoryResponse
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			mockResp:   []dtos.CategoryResponse{{ID: 1, Name: "Electronics"}},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty",
			mockResp:   []dtos.CategoryResponse{},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "error",
			mockResp:   nil,
			mockErr:    ErrProductNotFound,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				listCategories: func(ctx context.Context) ([]dtos.CategoryResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			req := httptest.NewRequest(http.MethodGet, "/products/categories", nil)
			w := httptest.NewRecorder()

			h.ListCategories(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerListProducts(t *testing.T) {
	tests := []struct {
		name       string
		mockItems  []dtos.ProductListItem
		mockTotal  int64
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success with items",
			mockItems:  []dtos.ProductListItem{{ID: 1, Title: "Test"}},
			mockTotal:  1,
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty",
			mockItems:  []dtos.ProductListItem{},
			mockTotal:  0,
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid sort",
			mockItems:  nil,
			mockTotal:  0,
			mockErr:    ErrInvalidSort,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]dtos.ProductListItem, int64, error) {
					return tt.mockItems, tt.mockTotal, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			req := httptest.NewRequest(http.MethodGet, "/products?sort=created_desc", nil)
			w := httptest.NewRecorder()

			h.ListProducts(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerGetProduct(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		mockResp   dtos.ProductResponse
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			productID:  "1",
			mockResp:   dtos.ProductResponse{ID: 1, Title: "Test Product"},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			productID:  "1",
			mockResp:   dtos.ProductResponse{},
			mockErr:    ErrProductNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			productID:  "abc",
			mockResp:   dtos.ProductResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				getProduct: func(ctx context.Context, id int64) (dtos.ProductResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			r := newTestRouter("/products/{id}", http.HandlerFunc(h.GetProduct), 0)

			req := httptest.NewRequest(http.MethodGet, "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerCreateProduct(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		body       string
		mockResp   dtos.ProductResponse
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			userID:     1,
			body:       `{"title":"Test","price":"100","stock":10}`,
			mockResp:   dtos.ProductResponse{ID: 1, Title: "Test"},
			mockErr:    nil,
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing auth",
			userID:     0,
			body:       `{"title":"Test","price":"100","stock":10}`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid json",
			userID:     1,
			body:       `{invalid`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty title",
			userID:     1,
			body:       `{"title":"","price":"100","stock":10}`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    ErrTitleRequired,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				createProduct: func(ctx context.Context, userID int64, req dtos.CreateProductRequest) (dtos.ProductResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			r := newTestRouter("/products", http.HandlerFunc(h.CreateProduct), tt.userID)

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerUpdateProduct(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		productID  string
		body       string
		mockResp   dtos.ProductResponse
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			userID:     1,
			productID:  "1",
			body:       `{"title":"Updated"}`,
			mockResp:   dtos.ProductResponse{ID: 1, Title: "Updated"},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing auth",
			userID:     0,
			productID:  "1",
			body:       `{"title":"Updated"}`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "forbidden",
			userID:     1,
			productID:  "1",
			body:       `{"title":"Updated"}`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    ErrForbidden,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				updateProduct: func(ctx context.Context, userID, id int64, req dtos.UpdateProductRequest) (dtos.ProductResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			r := newTestRouter("/products/{id}", http.HandlerFunc(h.UpdateProduct), tt.userID)

			req := httptest.NewRequest(http.MethodPatch, "/products/"+tt.productID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerDeleteProduct(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		productID  string
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			userID:     1,
			productID:  "1",
			mockErr:    nil,
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing auth",
			userID:     0,
			productID:  "1",
			mockErr:    nil,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "not found",
			userID:     1,
			productID:  "1",
			mockErr:    ErrProductNotFound,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				deleteProduct: func(ctx context.Context, userID, id int64) error {
					return tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			r := newTestRouter("/products/{id}", http.HandlerFunc(h.DeleteProduct), tt.userID)

			req := httptest.NewRequest(http.MethodDelete, "/products/"+tt.productID, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerModerateProduct(t *testing.T) {
	tests := []struct {
		name       string
		productID  string
		body       string
		mockResp   dtos.ProductResponse
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success approve",
			productID:  "1",
			body:       `{"status":"active"}`,
			mockResp:   dtos.ProductResponse{ID: 1, Status: "active"},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "success reject",
			productID:  "1",
			body:       `{"status":"rejected","reason":"Bad quality"}`,
			mockResp:   dtos.ProductResponse{ID: 1, Status: "rejected"},
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid status",
			productID:  "1",
			body:       `{"status":"invalid"}`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    ErrInvalidStatus,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid id",
			productID:  "abc",
			body:       `{"status":"active"}`,
			mockResp:   dtos.ProductResponse{},
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				moderateProduct: func(ctx context.Context, id int64, req dtos.ModerateProductRequest) (dtos.ProductResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			r := newTestRouter("/products/{id}/moderate", http.HandlerFunc(h.ModerateProduct), 0)

			req := httptest.NewRequest(http.MethodPatch, "/products/"+tt.productID+"/moderate", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerListProductsByStatus(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		mockItems  []dtos.ProductListItem
		mockTotal  int64
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success pending",
			query:      "?status=pending",
			mockItems:  []dtos.ProductListItem{{ID: 1, Title: "Pending Product"}},
			mockTotal:  1,
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing status",
			query:      "",
			mockItems:  nil,
			mockTotal:  0,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid status",
			query:      "?status=invalid",
			mockItems:  nil,
			mockTotal:  0,
			mockErr:    ErrInvalidStatus,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				listProductsByStatus: func(ctx context.Context, status string, page, limit int) ([]dtos.ProductListItem, int64, error) {
					return tt.mockItems, tt.mockTotal, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			req := httptest.NewRequest(http.MethodGet, "/admin/products"+tt.query, nil)
			w := httptest.NewRecorder()

			h.ListProductsByStatus(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

func TestHandlerListProductsByStore(t *testing.T) {
	tests := []struct {
		name       string
		storeID    string
		mockItems  []dtos.ProductListItem
		mockTotal  int64
		mockErr    error
		wantStatus int
	}{
		{
			name:       "success",
			storeID:    "1",
			mockItems:  []dtos.ProductListItem{{ID: 1, Title: "Store Product"}},
			mockTotal:  1,
			mockErr:    nil,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid store ID",
			storeID:    "abc",
			mockItems:  nil,
			mockTotal:  0,
			mockErr:    nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &mockProductServiceForHandler{
				listProductsByStore: func(ctx context.Context, storeID int64, page, limit int) ([]dtos.ProductListItem, int64, error) {
					return tt.mockItems, tt.mockTotal, tt.mockErr
				},
			}
			h := NewProductHandler(svc, nil)

			r := newTestRouter("/stores/{id}/products", http.HandlerFunc(h.ListProductsByStore), 0)

			req := httptest.NewRequest(http.MethodGet, "/stores/"+tt.storeID+"/products", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d; body: %s", w.Code, tt.wantStatus, w.Body.String())
			}
		})
	}
}

// TestJSONResponse verifies the JSON structure of responses
func TestJSONResponse(t *testing.T) {
	svc := &mockProductServiceForHandler{
		listProducts: func(ctx context.Context, storeID, categoryID *int64, minPrice, maxPrice *string, search *string, sort string, page, limit int) ([]dtos.ProductListItem, int64, error) {
			return []dtos.ProductListItem{{ID: 1, Title: "Test", Price: "100"}}, 1, nil
		},
	}
	h := NewProductHandler(svc, nil)

	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	w := httptest.NewRecorder()

	h.ListProducts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp httputil.PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("expected total 1, got %d", resp.Total)
	}

	// Verify Items is a slice
	itemsSlice, ok := resp.Items.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", resp.Items)
	}
	if len(itemsSlice) != 1 {
		t.Errorf("expected 1 item, got %d", len(itemsSlice))
	}

	// Verify first item has expected fields
	firstItem, ok := itemsSlice[0].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", itemsSlice[0])
	}
	if firstItem["id"] != float64(1) {
		t.Errorf("expected id 1, got %v", firstItem["id"])
	}
	if firstItem["title"] != "Test" {
		t.Errorf("expected title 'Test', got %v", firstItem["title"])
	}
}
