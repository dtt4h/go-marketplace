package orders

import (
	"net/http"

	db "github.com/dtt4h/go-marketplace/internal/database/sqlc"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

// AdminStatsHandler handles admin HTTP requests for statistics.
type AdminStatsHandler struct {
	service AdminStatsService
}

// NewAdminStatsHandler creates a new AdminStatsHandler.
func NewAdminStatsHandler(service AdminStatsService) *AdminStatsHandler {
	return &AdminStatsHandler{service: service}
}

// GetStats godoc
// @Summary      Get platform statistics (admin)
// @Description  Returns platform-wide statistics (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  dtos.AdminStatsResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      403  {object}  httputil.ErrorResponse
// @Router       /admin/stats [get]
func (h *AdminStatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	role, ok := mw.RoleFromCtx(r.Context())
	if !ok || role != string(db.UserRoleAdmin) {
		httputil.Forbidden(w, "insufficient permissions")
		return
	}

	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		httputil.InternalError(w, r, err.Error())
		return
	}

	httputil.JSON(w, http.StatusOK, stats)
}
