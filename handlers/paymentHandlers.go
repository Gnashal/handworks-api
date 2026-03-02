package handlers

import (
	"context"
	"errors"
	"handworks-api/types"
	"net/http"
	"strconv"
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
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// MakeQuotation godoc
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
// @Router /payment/quotes [get]
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
