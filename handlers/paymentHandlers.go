package handlers

import (
	"context"
	"errors"
	"handworks-api/types"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// MakeQuotation godoc
// @Summary Create a quotation
// @Description Generate a new quotation for a customer
// @Security BearerAuth
// @Tags Payment
// @Accept json
// @Produce json
// @Param input body types.QuoteRequest true "Quote details"
// @Success 200 {object} types.QuoteResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/quote [post]
func (h *PaymentHandler) MakeQuotation(c *gin.Context) {
	var req types.QuoteRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.MakeQuotation(ctx, req)
	if err != nil {
		// Check for validation errors (like hours exceeded)
		if strings.Contains(err.Error(), "exceed maximum allowed limit") ||
			strings.Contains(err.Error(), "validation failed") ||
			strings.Contains(err.Error(), "areas above 100 SQM") ||
			strings.Contains(err.Error(), "daily limit") {
			// Return 400 Bad Request for validation errors
			c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
			return
		}
		// Return 500 for actual server errors
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// MakePublicQuotation godoc
// @Summary Create a quotation
// @Description Generate a new quotation
// @Tags Payment
// @Accept json
// @Produce json
// @Param input body types.QuoteRequest true "Quote details"
// @Success 200 {object} types.QuoteResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/quote/preview [post]
func (h *PaymentHandler) MakePublicQuotation(c *gin.Context) {
	var req types.QuoteRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.MakePublicQuotation(ctx, req)
	if err != nil {
		// Check for validation errors
		if strings.Contains(err.Error(), "exceed maximum allowed limit") ||
			strings.Contains(err.Error(), "validation failed") ||
			strings.Contains(err.Error(), "areas above 100 SQM") ||
			strings.Contains(err.Error(), "daily limit") {
			// Return 400 Bad Request for validation errors
			c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
			return
		}
		// Return 500 for actual server errors
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetAllQuotesFromCustomer godoc
// @Summary Get all quotations for a customer
// @Security BearerAuth
// @Description Retrieve all quotations associated with a specific customer with optional date filtering and pagination.
// @Tags Payment
// @Accept json
// @Produce json
// @Param customerId query string true "Customer ID"
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number (starting at 0)" default(0)
// @Param limit query int false "Number of quotes per page" default(10)
// @Success 200 {object} types.FetchAllQuotesResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/customer/quotes [get]
func (h *PaymentHandler) GetAllQuotesFromCustomer(c *gin.Context) {
	customerId := c.Query("customerId")
	if customerId == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("customerId is required")))
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid page")))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid limit")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetAllQuotesFromCustomer(ctx, customerId, startDate, endDate, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetAllQuotes godoc
// @Summary Get all quotations
// @Security BearerAuth
// @Description Retrieve all quotations with optional date filtering and pagination.
// @Tags Payment
// @Accept json
// @Produce json
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number (starting at 0)" default(0)
// @Param limit query int false "Number of quotes per page" default(10)
// @Success 200 {object} types.FetchAllQuotesResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/quote/quotes [get]
func (h *PaymentHandler) GetAllQuotes(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid page")))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid limit")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetAllQuotes(ctx, startDate, endDate, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// Keep Swagger annotation as-is
// GetQuoteByIDForCustomer godoc
// @Summary Get quote by ID for a specific customer
// @Security BearerAuth
// @Description Retrieves a quote by ID that belongs to a specific customer
// @Tags Payment
// @Accept json
// @Produce json
// @Param quoteId query string true "Quote ID"
// @Param customerId query string true "Customer ID"
// @Success 200 {object} types.Quote
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/quote [get]
func (h *PaymentHandler) GetQuoteByIDForCustomer(c *gin.Context) {
	quoteId := c.Query("quoteId")
	customerId := c.Query("customerId")

	if quoteId == "" {
		h.Logger.Info("quoteId is empty")
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("quoteId is required")))
		return
	}

	if customerId == "" {
		h.Logger.Info("customerId is empty")
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("customerId is required")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	quote, err := h.Service.GetQuoteByIDForCustomer(ctx, quoteId, customerId)
	if err != nil {
		h.Logger.Info("Service error: %v", err)
		if err.Error() == "quote not found for this customer" {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, quote)
}

// CreateOrder godoc
// @Summary Create an order
// @Description Create a new order from an accepted quotation
// @Security BearerAuth
// @Tags Payment
// @Accept json
// @Produce json
// @Param input body types.CreateOrderRequest true "Order details"
// @Success 200 {object} types.CreateOrderResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/order [post]
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req types.CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.CreateOrder(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetOrder godoc
// @Summary Get order by ID
// @Security BearerAuth
// @Description Retrieve a single order by its ID
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} types.Order
// @Failure 400 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/order [get]
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	orderID := c.Query("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("order id is required")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetOrder(ctx, orderID)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		} else {
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		}
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetOrders godoc
// @Summary Get all orders
// @Security BearerAuth
// @Description Retrieve all orders with pagination
// @Tags Payment
// @Accept json
// @Produce json
// @Param page query int false "Page number (starting at 0)" default(0)
// @Param limit query int false "Number of orders per page" default(10)
// @Success 200 {object} types.GetOrdersResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/order/orders [get]
func (h *PaymentHandler) GetOrders(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid page")))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid limit")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetOrders(ctx, page, limit, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetOrderByCustomerId godoc
// @Summary Get orders by customer ID
// @Security BearerAuth
// @Description Retrieve all orders for a specific customer with pagination
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param page query int false "Page number (starting at 0)" default(0)
// @Param limit query int false "Number of orders per page" default(10)
// @Success 200 {object} types.GetOrdersResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/order/customer [get]
func (h *PaymentHandler) GetOrderByCustomer(c *gin.Context) {
	customerID := c.Query("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("customer id is required")))
		return
	}
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid page")))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid limit")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetOrdersByCustomer(ctx, page, limit, startDate, endDate, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetPaymentsByOrderID godoc
// @Summary Get payments by order ID
// @Security BearerAuth
// @Description Retrieve all payments for a specific order with pagination
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param page query int false "Page number (starting at 0)" default(0)
// @Param limit query int false "Number of payments per page" default(10)
// @Success 200 {object} types.GetPaymentsResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/payments/order [get]
func (h *PaymentHandler) GetPaymentsByOrderID(c *gin.Context) {
	orderID := c.Query("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("order id is required")))
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid page")))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid limit")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetPaymentsByOrderID(ctx, page, limit, startDate, endDate, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// GetPaymentsByCustomerID godoc
// @Summary Get payments by customer ID
// @Security BearerAuth
// @Description Retrieve all payments for a specific customer with pagination
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param page query int false "Page number (starting at 0)" default(0)
// @Param limit query int false "Number of payments per page" default(10)
// @Success 200 {object} types.GetPaymentsResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/customer [get]
func (h *PaymentHandler) GetPaymentsByCustomerID(c *gin.Context) {
	customerID := c.Query("id")
	if customerID == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("customer id is required")))
		return
	}

	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid page")))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("invalid limit")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetPaymentsByCustomerID(ctx, page, limit, startDate, endDate, customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// PayDownpayment godoc
// @Summary Create downpayment payment intent
// @Security BearerAuth
// @Description Create a PayMongo payment intent for the order's downpayment amount
// @Tags Payment
// @Accept json
// @Produce json
// @Param id query string true "Order ID"
// @Success 200 {object} types.PaymentIntentResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/intent/downpayment/{id} [post]
func (h *PaymentHandler) CreateDownpaymentIntent(c *gin.Context) {
	orderId := c.Param("id")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("order id is required")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := h.Service.CreateDownpaymentIntent(ctx, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// PayFullPayment godoc
// @Summary Create full payment payment intent
// @Security BearerAuth
// @Description Create a PayMongo payment intent for the order's full remaining balance
// @Tags Payment
// @Accept json
// @Produce json
// @Param id query string true "Order ID"
// @Success 200 {object} types.PaymentIntentResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /payment/intent/fullpayment/{id} [post]
func (h *PaymentHandler) CreateFullPaymentIntent(c *gin.Context) {
	orderId := c.Param("id")
	if orderId == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("order id is required")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res, err := h.Service.CreateFullPaymentIntent(ctx, orderId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *PaymentHandler) HandlePaymongoWebhook(c *gin.Context) {
	var webhook types.WebhookEvent
	if err := c.ShouldBindJSON(&webhook); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	switch webhook.Data.Type {
	case "payment.paid":
		if err := h.Service.HandlePaymentPaid(ctx, webhook.Data); err != nil {
			h.Logger.Error("failed to handle payment.paid webhook: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
			return
		}
	case "payment.failed":
		if err := h.Service.HandlePaymentFailed(ctx, webhook.Data); err != nil {
			h.Logger.Error("failed to handle payment.failed webhook: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process webhook"})
			return
		}
	default:
		h.Logger.Info("unhandled webhook event type: %s", webhook.Data.Type)
	}
	res := gin.H{"status": "success"}
	c.JSON(http.StatusOK, res)
}
