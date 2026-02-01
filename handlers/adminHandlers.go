package handlers

import (
	"context"
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
// @Param input body types.AdminDashboardRequest true "Admin dashboard data"
// @Success 200 {object} types.AdminDashboardResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /admin/dashboard [get]
func (h *AdminHandler) GetAdminDashboard(c *gin.Context) {
	adminId := c.Query("adminId")
	dateFilter := c.Query("dateFilter")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &types.AdminDashboardRequest {
		AdminID: adminId,
		DateFilter: dateFilter,
	}
	res, err := h.Service.GetAdminDashboard(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}
	c.JSON(http.StatusOK, res)
}