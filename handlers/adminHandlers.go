package handlers

import (
	"context"
	"errors"
	"handworks-api/tasks"
	"handworks-api/types"
	"net/http"
	"strings"
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

// AssignEmployeeToBooking godoc
// @Summary Assign or unassign a cleaner to a booking
// @Description Adds or removes a cleaner from a booking after schedule conflict checks
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.AssignEmployeeToBookingRequest true "Assign employee data"
// @Success 200 {object} types.AssignEmployeeToBookingResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 409 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/employee/assign [post]
func (h *AdminHandler) AssignEmployeeToBooking(c *gin.Context) {
	var req types.AssignEmployeeToBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	req.Action = types.AssignEmployeeAction(strings.ToUpper(string(req.Action)))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.AssignEmployeeToBooking(ctx, &req)
	if err != nil {
		switch {
		case errors.Is(err, tasks.ErrBookingNotFound), errors.Is(err, tasks.ErrEmployeeNotFoundOrInactive):
			c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
			return
		case errors.Is(err, tasks.ErrCleanerAlreadyAssigned), errors.Is(err, tasks.ErrCleanerHasConflict), errors.Is(err, tasks.ErrCleanerNotAssigned):
			c.JSON(http.StatusConflict, types.NewErrorResponse(err))
			return
		default:
			c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
			return
		}
	}

	c.JSON(http.StatusOK, res)
}

// ListAvailableCleaners godoc
// @Summary List available cleaners for a booking
// @Description Returns cleaners with no overlapping non-cancelled responsibilities for the booking schedule
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param bookingId query string true "Booking ID"
// @Success 200 {object} types.AvailableCleanersResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/employee/available [get]
func (h *AdminHandler) ListAvailableCleaners(c *gin.Context) {
	var req types.AvailableCleanersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.GetAvailableCleaners(ctx, &req)
	if err != nil {
		if errors.Is(err, tasks.ErrBookingNotFound) {
			c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
			return
		}
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
// @Summary Fetch calendar bookings by month
// @Description Returns calendar booking cards for the selected month
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param month query string true "Month in YYYY-MM format"
// @Success 200 {object} types.CalendarBookingResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/booking/calendar [get]
func (h *AdminHandler) GetCalendarBookings(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(errors.New("month query param is required (YYYY-MM)")))
		return
	}

	bookings, err := h.Service.GetCalendarBookings(c.Request.Context(), month)
	if err != nil {
		if err.Error() == "month is required" || err.Error() == "invalid month format, expected YYYY-MM" {
			c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// GetBookingTrends godoc
// @Summary Fetch booking trend analytics
// @Description Returns booking trends formatted for charting with weeklyData and monthlyData points
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.BookingTrendsResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/booking-trends [get]
func (h *AdminHandler) GetBookingTrends(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := h.Service.GetBookingTrends(ctx)
	if err != nil {
		h.Logger.Error("failed to get booking trends: %v", err)
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}
