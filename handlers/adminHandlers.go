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
