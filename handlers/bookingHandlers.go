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
	if err != nil || page < 1 {
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
// @Summary Get booking by user ID
// @Description Retrieve all bookings for a specific user
// @Tags Booking
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param uid path string true "User ID"
// @Success 200 {array} map[string]interface{}
// @Router /booking/user/{uid} [get]
func (h *BookingHandler) GetBookingByUId(c *gin.Context) {
	_ = h.Service.GetBookingByUId(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"status": "success"})
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
	bookingId := c.Param("id")
	if bookingId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "bookingId is required",
		})
		return
	}

	var req types.UpdateBookingEvent
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "invalid request body: " + err.Error(),
		})
		return
	}

	req.BookingID = bookingId

	updatedBooking, err := h.Service.UpdateBooking(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"booking": updatedBooking,
	})
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
