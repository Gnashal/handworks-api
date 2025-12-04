package handlers

import (
	"context"
	"fmt"
	"handworks-api/types"
	"net/http"
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

// GetBookingById godoc
// @Summary Get booking by ID
// @Description Retrieve booking information by its database ID
// @Tags Booking
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} map[string]interface{}
// @Router /booking/{id} [get]
func (h *BookingHandler) GetBookingById(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(fmt.Errorf("missing id parameter")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	booking, err := h.Service.GetBookingById(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, booking)
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
func (h *BookingHandler) GetBookingsByUId(c *gin.Context) {
	userId := c.Param("id")

	if userId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "userId is required",
		})
		return
	}

	bookings, err := h.Service.GetBookingsByUId(c.Request.Context(), userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "success",
		"bookings": bookings,
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
