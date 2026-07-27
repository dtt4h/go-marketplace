package payments

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/dtt4h/go-marketplace/internal/server/dtos"
	mw "github.com/dtt4h/go-marketplace/internal/server/middleware"
	"github.com/dtt4h/go-marketplace/pkg/httputil"
)

type PaymentHandler struct {
	service PaymentService
}

func NewPaymentHandler(service PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

// CreatePayment godoc
// @Summary      Create a payment
// @Description  Initiates a payment for an order (returns confirmation URL)
// @Tags         payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.CreatePaymentRequest  true  "Payment data"
// @Success      201  {object}  dtos.PaymentResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Failure      409  {object}  httputil.ErrorResponse
// @Router       /payments [post]
func (h *PaymentHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	var req dtos.CreatePaymentRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	resp, err := h.service.CreatePayment(r.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrPaymentExists):
			httputil.Conflict(w, err.Error())
		case errors.Is(err, ErrAlreadyPaid):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusCreated, resp)
}

// GetPayment godoc
// @Summary      Get payment by ID
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  int  true  "Payment ID"
// @Success      200  {object}  dtos.PaymentResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /payments/{id} [get]
func (h *PaymentHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid payment id", nil)
		return
	}

	resp, err := h.service.GetPayment(r.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			httputil.NotFound(w, err.Error())
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}

// Webhook godoc
// @Summary      Payment provider webhook
// @Description  Receives payment status updates from provider (no JWT auth)
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        request  body  dtos.WebhookRequest  true  "Webhook payload"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /payments/webhook [post]
func (h *PaymentHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	var req dtos.WebhookRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		httputil.ValidationError(w, "invalid request body", nil)
		return
	}

	if err := h.service.HandleWebhook(r.Context(), req); err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrInvalidWebhook):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RefundPayment godoc
// @Summary      Refund a payment
// @Description  Refunds a succeeded payment (order owner only)
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        id   path  int  true  "Payment ID"
// @Success      200  {object}  dtos.PaymentResponse
// @Failure      400  {object}  httputil.ErrorResponse
// @Failure      401  {object}  httputil.ErrorResponse
// @Failure      404  {object}  httputil.ErrorResponse
// @Router       /payments/{id}/refund [post]
func (h *PaymentHandler) RefundPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := mw.UserIDFromCtx(r.Context())
	if !ok {
		httputil.Unauthorized(w, "not authenticated")
		return
	}

	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		httputil.ValidationError(w, "invalid payment id", nil)
		return
	}

	resp, err := h.service.RefundPayment(r.Context(), userID, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrPaymentNotFound):
			httputil.NotFound(w, err.Error())
		case errors.Is(err, ErrRefundFailed):
			httputil.ValidationError(w, err.Error(), nil)
		default:
			httputil.InternalError(w, err.Error())
		}
		return
	}

	httputil.JSON(w, http.StatusOK, resp)
}
