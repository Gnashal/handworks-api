package handlers

import (
	"context"
	"errors"
	"handworks-api/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetAdminDashboard godoc
// @Summary Fetch data for admin dashboard
// @Description Fetch data for admin dashboard
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param adminId query string true "Admin ID"
// @Param dateFilter query string false "Date filter (year, month, week)"
// @Success 200 {object} types.AdminDashboardResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/dashboard [get]
func (h *AdminHandler) GetAdminDashboard(c *gin.Context) {
	adminId := c.Query("adminId")
	dateFilter := c.Query("dateFilter")

	if adminId == "" {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(errors.New("You are not an admin")))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &types.AdminDashboardRequest{
		AdminID:    adminId,
		DateFilter: dateFilter,
	}
	res, err := h.Service.GetAdminDashboard(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// OnboardEmployee godoc
// @Summary Onboard a new employee
// @Description Create a new employee account
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.OnboardEmployeeRequest true "Employee onboard data"
// @Success 200 {object} types.SignUpEmployeeResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/employee/onboard [post]
func (h *AdminHandler) OnboardEmployee(c *gin.Context) {
	var req types.OnboardEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Service.OnboardEmployee(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// AcceptBooking godoc
// @Summary Accept a booking
// @Description Updates the booking review status to SCHEDULED, triggering a notification to assigned employees
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} types.AcceptBookingResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/booking/approve/{id} [post]
func (h *AdminHandler) AcceptBooking(c *gin.Context) {
	bookingId := c.Param("id")
	if bookingId == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("Booking ID is required")))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Service.AcceptBooking(ctx, bookingId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// AssignResourcesToBooking godoc
// @Summary Assign resources to a booking
// @Description Admin override to manually assign resources (supplies) to a booking
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.AssignResourcesToBookingRequest true "Assign resources data"
// @Success 200 {object} types.AssignInventoryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/inventory/assign-resources [post]
func (h *AdminHandler) AssignResourcesToBooking(c *gin.Context) {
	var req types.AssignResourcesToBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Service.AssignResourcesToBooking(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// AssignEquipmentToBooking godoc
// @Summary Assign equipment to a booking
// @Description Admin override to manually assign equipment to a booking
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.AssignEquipmentToBookingRequest true "Assign equipment data"
// @Success 200 {object} types.AssignInventoryResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/inventory/assign-equipment [post]
func (h *AdminHandler) AssignEquipmentToBooking(c *gin.Context) {
	var req types.AssignEquipmentToBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := h.Service.AssignEquipmentToBooking(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}

// GetCalendarBookings godoc
// @Summary Fetch current month calendar bookings
// @Description Returns calendar booking cards for the current calendar month
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.CalendarBookingResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/booking/calendar [get]
func (h *AdminHandler) GetCalendarBookings(c *gin.Context) {
	bookings, err := h.Service.GetCalendarBookings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, bookings)
}
