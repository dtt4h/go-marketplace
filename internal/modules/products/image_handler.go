package products

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
	"github.com/dtt4h/go-marketplace/pkg/pgutil"
)

// ImageHandler handles HTTP requests for product image operations.
type ImageHandler struct {
	service ImageUploadService
}

// NewImageHandler creates a new ImageHandler.
func NewImageHandler(service ImageUploadService) *ImageHandler {
	return &ImageHandler{service: service}
}

// UploadImage godoc
// @Summary      Upload an image for a product
// @Description  Uploads an image file and associates it with a product (seller only)
// @Tags         products
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id       path  int                            true  "Product ID"
// @Param        file     formData file                       true  "Image file"
// @Success      201  {object}  dtos.ProductImageResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/{id}/images [post]
func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httputil.InternalError(w, "image storage is not configured")
		return
	}

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

	file, header, err := r.FormFile("file")
	if err != nil {
		httputil.ValidationError(w, "file is required", nil)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := io.ReadFull(file, buf)
		contentType = http.DetectContentType(buf[:n])
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			httputil.InternalError(w, "failed to read file")
			return
		}
	}

	if !isAllowedImageType(contentType) {
		httputil.ValidationError(w, "unsupported file type, only images are allowed", nil)
		return
	}

	limitedReader := io.LimitReader(file, 10*1024*1024)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		httputil.InternalError(w, "failed to read file")
		return
	}

	if len(data) == 0 {
		httputil.ValidationError(w, "file is empty", nil)
		return
	}

	resp, err := h.service.UploadImage(r.Context(), id, userID, bytes.NewReader(data), int64(len(data)), contentType)
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

	httputil.JSON(w, http.StatusCreated, resp)
}

// DeleteImage godoc
// @Summary      Delete a product image
// @Description  Deletes an image from a product (seller only)
// @Tags         products
// @Security     BearerAuth
// @Param        id  path  int  true  "Image ID"
// @Success      204
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/images/{id} [delete]
func (h *ImageHandler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httputil.InternalError(w, "image storage is not configured")
		return
	}

	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid image id", nil)
		return
	}

	err = h.service.DeleteImage(r.Context(), id, userID)
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

// GetPresignedUploadURL godoc
// @Summary      Get presigned upload URL
// @Description  Returns a presigned URL for direct upload to S3 (seller only)
// @Tags         products
// @Security     BearerAuth
// @Param        id        path  int    true  "Product ID"
// @Param        content_type  query  string  false  "Content-Type of the file"
// @Success      200  {object}  map[string]string
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /products/{id}/images/presigned [get]
func (h *ImageHandler) GetPresignedUploadURL(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		httputil.InternalError(w, "image storage is not configured")
		return
	}

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

	contentType := r.URL.Query().Get("content_type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectKey := fmt.Sprintf("products/images/%s", pgutil.GenerateUploadKey(contentType))

	url, err := h.service.GetPresignedUploadURL(r.Context(), id, userID, contentType, objectKey)
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

	httputil.JSON(w, http.StatusOK, map[string]string{
		"upload_url":   url,
		"object_key":   objectKey,
		"content_type": contentType,
	})
}

func isAllowedImageType(contentType string) bool {
	allowed := map[string]bool{
		"image/jpeg":    true,
		"image/png":     true,
		"image/webp":    true,
		"image/gif":     true,
		"image/svg+xml": true,
	}
	return allowed[contentType]
}
