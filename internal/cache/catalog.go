package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	"github.com/redis/go-redis/v9"
)

const (
	productsCacheKey   = "cache:products:list"
	categoriesCacheKey = "cache:categories:list"
	productDetailKey   = "cache:product:detail:"
	defaultTTL         = 60 * time.Second
)

// CatalogCache wraps Redis operations for catalog caching.
type CatalogCache struct {
	rdb *redis.Client
	log *slog.Logger
	ttl time.Duration
}

// NewCatalogCache creates a new catalog cache. Returns nil if Redis is not available.
func NewCatalogCache(rdb *redis.Client, log *slog.Logger, ttl time.Duration) *CatalogCache {
	if rdb == nil {
		return nil
	}
	return &CatalogCache{
		rdb: rdb,
		log: log,
		ttl: ttl,
	}
}

// GetProducts tries to get cached products list. Returns (items, total, true) on hit,
// or (nil, 0, false) on miss/error.
func (c *CatalogCache) GetProducts(ctx context.Context) ([]dtos.ProductListItem, int64, bool) {
	if c == nil {
		return nil, 0, false
	}

	data, err := c.rdb.Get(ctx, productsCacheKey).Bytes()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("cache get products error", slog.String("error", err.Error()))
		}
		return nil, 0, false
	}

	var result struct {
		Items []dtos.ProductListItem `json:"items"`
		Total int64                  `json:"total"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		c.log.Warn("cache unmarshal products error", slog.String("error", err.Error()))
		return nil, 0, false
	}

	c.log.Debug("cache hit", slog.String("key", productsCacheKey))
	return result.Items, result.Total, true
}

// SetProducts caches the products list.
func (c *CatalogCache) SetProducts(ctx context.Context, items []dtos.ProductListItem, total int64) {
	if c == nil {
		return
	}

	data, err := json.Marshal(struct {
		Items []dtos.ProductListItem `json:"items"`
		Total int64                  `json:"total"`
	}{
		Items: items,
		Total: total,
	})
	if err != nil {
		c.log.Error("cache marshal products error", slog.String("error", err.Error()))
		return
	}

	if err := c.rdb.SetEx(ctx, productsCacheKey, data, c.ttl).Err(); err != nil {
		c.log.Error("cache set products error", slog.String("error", err.Error()))
	}
}

// InvalidateProducts removes the products list cache.
func (c *CatalogCache) InvalidateProducts(ctx context.Context) {
	if c == nil {
		return
	}
	_ = c.rdb.Del(ctx, productsCacheKey).Err()
}

// InvalidateProduct removes a single product detail cache.
func (c *CatalogCache) InvalidateProduct(ctx context.Context, productID int64) {
	if c == nil {
		return
	}
	_ = c.rdb.Del(ctx, productDetailKey+fmt.Sprintf("%d", productID)).Err()
}

// GetProductDetail tries to get cached product detail.
func (c *CatalogCache) GetProductDetail(ctx context.Context, productID int64) (dtos.ProductResponse, bool) {
	if c == nil {
		return dtos.ProductResponse{}, false
	}

	key := productDetailKey + fmt.Sprintf("%d", productID)
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("cache get product detail error", slog.String("error", err.Error()))
		}
		return dtos.ProductResponse{}, false
	}

	var resp dtos.ProductResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		c.log.Warn("cache unmarshal product detail error", slog.String("error", err.Error()))
		return dtos.ProductResponse{}, false
	}

	c.log.Debug("cache hit", slog.String("key", key))
	return resp, true
}

// SetProductDetail caches a product detail.
func (c *CatalogCache) SetProductDetail(ctx context.Context, productID int64, resp dtos.ProductResponse) {
	if c == nil {
		return
	}

	key := productDetailKey + fmt.Sprintf("%d", productID)
	data, err := json.Marshal(resp)
	if err != nil {
		c.log.Error("cache marshal product detail error", slog.String("error", err.Error()))
		return
	}

	if err := c.rdb.SetEx(ctx, key, data, c.ttl).Err(); err != nil {
		c.log.Error("cache set product detail error", slog.String("error", err.Error()))
	}
}

// InvalidateCategories removes the categories cache.
func (c *CatalogCache) InvalidateCategories(ctx context.Context) {
	if c == nil {
		return
	}
	_ = c.rdb.Del(ctx, categoriesCacheKey).Err()
}

// GetCategories tries to get cached categories.
func (c *CatalogCache) GetCategories(ctx context.Context) ([]dtos.CategoryResponse, bool) {
	if c == nil {
		return nil, false
	}

	data, err := c.rdb.Get(ctx, categoriesCacheKey).Bytes()
	if err != nil {
		if err != redis.Nil {
			c.log.Warn("cache get categories error", slog.String("error", err.Error()))
		}
		return nil, false
	}

	var items []dtos.CategoryResponse
	if err := json.Unmarshal(data, &items); err != nil {
		c.log.Warn("cache unmarshal categories error", slog.String("error", err.Error()))
		return nil, false
	}

	c.log.Debug("cache hit", slog.String("key", categoriesCacheKey))
	return items, true
}

// SetCategories caches the categories list.
func (c *CatalogCache) SetCategories(ctx context.Context, items []dtos.CategoryResponse) {
	if c == nil {
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		c.log.Error("cache marshal categories error", slog.String("error", err.Error()))
		return
	}

	if err := c.rdb.SetEx(ctx, categoriesCacheKey, data, c.ttl).Err(); err != nil {
		c.log.Error("cache set categories error", slog.String("error", err.Error()))
	}
}
