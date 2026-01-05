package handlers

import (
	"context"
	"handworks-api/types"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateBooking godoc
// @Summary Create a new booking
// @Description Creates a booking record
// @Tags Booking
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.CreateBookingRequest true "Booking info"
// @Success 200 {object} types.Booking
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /booking [post]
func (h *BookingHandler) CreateBooking(c *gin.Context) {
	var req types.CreateBookingRequest

	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Service.CreateBooking(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetBookings godoc
// @Summary Get all bookings with filters
// @Description Retrieve all bookings with optional date filtering and pagination
// @Tags Booking
// @Accept json
// @Produce json
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} types.FetchAllBookingsResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /booking/bookings [get]
// @Security BearerAuth
func (h *BookingHandler) GetBookings(c *gin.Context) {
	// Parse query parameters
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid page parameter",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid limit parameter",
		})
		return
	}

	// Use context with timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// Call the service
	result, err := h.Service.GetBookings(ctx, startDate, endDate, page, limit)
	if err != nil {
		h.Logger.Error("failed to get bookings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to fetch bookings",
		})
		return
	}

	// Return response
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})

}

// GetBookingByUId godoc
// @Summary Get bookings by user ID
// @Description Retrieve all bookings for a specific user with optional date filtering and pagination
// @Tags Booking
// @Security BearerAuth
// @Accept json
// @Produce json
//
// @Param customerId path string true "Customer/User ID"
// @Param startDate query string false "Start date (YYYY-MM-DD)"
// @Param endDate query string false "End date (YYYY-MM-DD)"
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(10)
//
// @Success 200 {object} types.FetchAllBookingsResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
//
// @Router /booking/{customerId} [get]
func (h *BookingHandler) GetBookingByUId(c *gin.Context) {
	// Read path parameter
	customerId := c.Param("id")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Pagination defaults
	pageStr := c.DefaultQuery("page", "0")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid page parameter",
		})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid limit parameter",
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	result, err := h.Service.GetBookingByUId(ctx, customerId, startDate, endDate, page, limit)
	if err != nil {
		h.Logger.Error("failed to get customer bookings: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "failed to fetch bookings",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   result,
	})
}

// UpdateBooking godoc
// @Summary Update a booking
// @Description Update booking information
// @Tags Booking
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Param input body map[string]interface{} true "Updated booking info"
// @Success 200 {object} map[string]string
// @Router /booking/{id} [put]
func (h *BookingHandler) UpdateBooking(c *gin.Context) {
	_ = h.Service.UpdateBooking(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// DeleteBooking godoc
// @Summary Delete a booking
// @Description Remove booking by ID
// @Tags Booking
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]string
// @Router /booking/{id} [delete]
func (h *BookingHandler) DeleteBooking(c *gin.Context) {
	_ = h.Service.DeleteBooking(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
