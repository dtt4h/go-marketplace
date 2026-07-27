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
