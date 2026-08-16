package dtos

import (
	"strings"
	"time"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateProductRequest struct {
	CategoryID  *int64   `json:"category_id"`
	Title       string   `json:"title"`
	Description *string  `json:"description,omitempty"`
	Price       string   `json:"price"`
	Stock       int      `json:"stock"`
	Images      []string `json:"images,omitempty"`
}

type UpdateProductRequest struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Price       *string `json:"price,omitempty"`
	Stock       *int    `json:"stock,omitempty"`
}

type ModerateProductRequest struct {
	Status db.ProductStatus `json:"status"`
	Reason *string          `json:"reason,omitempty"`
}

type ProductStoreInfo struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type CategoryResponse struct {
	ID       int64              `json:"id"`
	Name     string             `json:"name"`
	Slug     string             `json:"slug"`
	ParentID *int64             `json:"parent_id,omitempty"`
	Children []CategoryResponse `json:"children,omitempty"`
}

type ProductImageResponse struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	Position int32  `json:"position"`
}

type ProductResponse struct {
	ID          int64                  `json:"id"`
	Title       string                 `json:"title"`
	Description *string                `json:"description,omitempty"`
	Price       string                 `json:"price"`
	Stock       int                    `json:"stock"`
	Status      string                 `json:"status"`
	Images      []ProductImageResponse `json:"images,omitempty"`
	Store       *ProductStoreInfo      `json:"store,omitempty"`
	Category    *CategoryResponse      `json:"category,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type ProductListItem struct {
	ID        int64                  `json:"id"`
	Title     string                 `json:"title"`
	Price     string                 `json:"price"`
	Stock     int                    `json:"stock"`
	Images    []ProductImageResponse `json:"images,omitempty"`
	Store     *ProductStoreInfo      `json:"store,omitempty"`
	Category  *CategoryResponse      `json:"category,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

func NumericToStr(n pgtype.Numeric) string {
	if !n.Valid {
		return ""
	}
	if n.Int == nil {
		return "0"
	}

	negative := n.Int.Sign() < 0
	s := n.Int.String()
	if negative {
		s = s[1:]
	}

	if n.Exp >= 0 {
		for i := int32(0); i < n.Exp; i++ {
			s += "0"
		}
	} else {
		exp := int(-n.Exp)
		if len(s) <= exp {
			s = strings.Repeat("0", exp-len(s)+1) + s
			s = "0." + s[1:]
		} else {
			pos := len(s) - exp
			s = s[:pos] + "." + s[pos:]
		}
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}

	if negative {
		s = "-" + s
	}
	return s
}

func BuildCategoryTree(flat []db.ListCategoriesRow) []CategoryResponse {
	childrenMap := make(map[int64][]int64)
	byID := make(map[int64]db.ListCategoriesRow, len(flat))
	var rootIDs []int64

	for _, c := range flat {
		byID[c.ID] = c
		if c.ParentID.Valid {
			childrenMap[c.ParentID.Int64] = append(childrenMap[c.ParentID.Int64], c.ID)
		} else {
			rootIDs = append(rootIDs, c.ID)
		}
	}

	var build func(id int64) CategoryResponse
	build = func(id int64) CategoryResponse {
		c := byID[id]
		cat := CategoryResponse{
			ID:   c.ID,
			Name: c.Name,
			Slug: c.Slug,
		}
		if c.ParentID.Valid {
			cat.ParentID = &c.ParentID.Int64
		}
		for _, childID := range childrenMap[id] {
			cat.Children = append(cat.Children, build(childID))
		}
		return cat
	}

	result := make([]CategoryResponse, 0, len(rootIDs))
	for _, id := range rootIDs {
		result = append(result, build(id))
	}
	return result
}

func toProductImages(images []db.ProductImage) []ProductImageResponse {
	if len(images) == 0 {
		return nil
	}
	result := make([]ProductImageResponse, 0, len(images))
	for _, img := range images {
		result = append(result, ProductImageResponse{
			ID:       img.ID,
			URL:      img.Url,
			Position: img.Position,
		})
	}
	return result
}

func ToProductResponse(row db.GetProductRow, images []db.ProductImage) ProductResponse {
	resp := ProductResponse{
		ID:        row.ID,
		Title:     row.Title,
		Price:     NumericToStr(row.Price),
		Stock:     int(row.Stock),
		Status:    string(row.Status),
		Images:    toProductImages(images),
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}

	if row.Description.Valid {
		resp.Description = &row.Description.String
	}

	if row.StoreName.Valid {
		store := ProductStoreInfo{
			ID:   row.StoreID,
			Name: row.StoreName.String,
		}
		store.Description = pgutil.TextToPtr(row.StoreDescription)
		resp.Store = &store
	}

	if row.CategoryName.Valid {
		cat := CategoryResponse{
			Name: row.CategoryName.String,
		}
		if row.CategoryID.Valid {
			cat.ID = row.CategoryID.Int64
		}
		if row.CategorySlug.Valid {
			cat.Slug = row.CategorySlug.String
		}
		resp.Category = &cat
	}

	return resp
}

func ToProductListItem(row db.ListProductsRow, images []db.ProductImage) ProductListItem {
	item := ProductListItem{
		ID:        row.ID,
		Title:     row.Title,
		Price:     NumericToStr(row.Price),
		Stock:     int(row.Stock),
		Images:    toProductImages(images),
		CreatedAt: row.CreatedAt.Time,
	}

	if row.StoreName.Valid {
		item.Store = &ProductStoreInfo{
			ID:   row.StoreID,
			Name: row.StoreName.String,
		}
	}

	if row.CategoryName.Valid {
		cat := CategoryResponse{
			Name: row.CategoryName.String,
		}
		if row.CategoryID.Valid {
			cat.ID = row.CategoryID.Int64
		}
		if row.CategorySlug.Valid {
			cat.Slug = row.CategorySlug.String
		}
		item.Category = &cat
	}

	return item
}

func ToProductListItemFromStoreRow(row db.ListProductsByStoreIDRow, images []db.ProductImage) ProductListItem {
	item := ProductListItem{
		ID:        row.ID,
		Title:     row.Title,
		Price:     NumericToStr(row.Price),
		Stock:     int(row.Stock),
		Images:    toProductImages(images),
		CreatedAt: row.CreatedAt.Time,
	}

	if row.StoreName.Valid {
		item.Store = &ProductStoreInfo{
			ID:   row.StoreID,
			Name: row.StoreName.String,
		}
	}

	if row.CategoryName.Valid {
		cat := CategoryResponse{
			Name: row.CategoryName.String,
		}
		if row.CategoryID.Valid {
			cat.ID = row.CategoryID.Int64
		}
		if row.CategorySlug.Valid {
			cat.Slug = row.CategorySlug.String
		}
		item.Category = &cat
	}

	return item
}
