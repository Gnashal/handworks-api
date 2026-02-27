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

// SignUpCustomer godoc
// @Summary Sign up a new customer
// @Description Create a new customer account
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.SignUpCustomerRequest true "Customer signup data"
// @Success 200 {object} types.SignUpCustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/customer/signup [post]
func (h *AccountHandler) SignUpCustomer(c *gin.Context) {
	var req types.SignUpCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.SignUpCustomer(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SignUpAdmin godoc
// @Summary Sign up a new admin
// @Description Create a new admin account
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.SignUpAdminRequest true "Admin signup data"
// @Success 200 {object} types.SignUpAdminResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/admin/signup [post]
func (h *AccountHandler) SignUpAdmin(c *gin.Context) {
	var req types.SignUpAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.SignUpAdmin(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCustomer godoc
// @Summary Get a customer by ID
// @Description Retrieve customer info
// @Security BearerAuth
// @Tags Account
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} types.GetCustomerResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /account/customer/{id} [get]
func (h *AccountHandler) GetCustomer(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.GetCustomer(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetCustomers godoc
// @Summary Get all customers
// @Description Retrieve all customer info
// @Security BearerAuth
// @Tags Account
// @Produce json
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} types.GetAllCustomersResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /account/customer [get]
func (h *AccountHandler) GetCustomers(c *gin.Context) {
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
	resp, err := h.Service.GetCustomers(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateCustomer godoc
// @Summary Update a customer
// @Description Update customer information
// @Security BearerAuth
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param input body types.UpdateCustomerRequest true "Updated customer info"
// @Success 200 {object} types.UpdateCustomerResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/customer/{id} [put]
func (h *AccountHandler) UpdateCustomer(c *gin.Context) {
	var req types.UpdateCustomerRequest
	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.UpdateCustomer(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteCustomer godoc
// @Summary Delete a customer
// @Description Remove a customer by ID
// @Security BearerAuth
// @Tags Account
// @Produce json
// @Param id path string true "Customer ID"
// @Param accId path string true "Account ID"
// @Success 200 {object} types.DeleteCustomerResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/customer/{id}/{accId} [delete]
func (h *AccountHandler) DeleteCustomer(c *gin.Context) {
	id := c.Param("id")
	accId := c.Param("accId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.DeleteCustomer(ctx, id, accId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SignUpEmployee godoc
// @Summary Sign up a new employee
// @Description Create a new employee account
// @Tags Account
// @Accept json
// @Produce json
// @Param input body types.SignUpEmployeeRequest true "Employee signup data"
// @Success 200 {object} types.SignUpEmployeeResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/employee/signup [post]
func (h *AccountHandler) SignUpEmployee(c *gin.Context) {
	var req types.SignUpEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.SignUpEmployee(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetEmployee godoc
// @Summary Get an employee by ID
// @Description Retrieve employee info
// @Security BearerAuth
// @Tags Account
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} types.GetEmployeeResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /account/employee/{id} [get]
func (h *AccountHandler) GetEmployee(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.GetEmployee(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetEmployees godoc
// @Summary Get all customers
// @Description Retrieve all customer info
// @Security BearerAuth
// @Tags Account
// @Produce json
// @Param page query int false "Page number" default(0)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} types.GetAllCustomersResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /account/employee [get]
func (h *AccountHandler) GetEmployees(c *gin.Context) {
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
	resp, err := h.Service.GetEmployees(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateEmployee godoc
// @Summary Update an employee
// @Description Update employee information
// @Security BearerAuth
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param input body types.UpdateEmployeeRequest true "Updated employee data"
// @Success 200 {object} types.UpdateEmployeeResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/employee/{id} [put]
func (h *AccountHandler) UpdateEmployee(c *gin.Context) {
	var req types.UpdateEmployeeRequest
	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.UpdateEmployee(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateEmployeePerformanceScore godoc
// @Summary Update employee performance score
// @Description Adjust performance score
// @Security BearerAuth
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param score body types.UpdatePerformanceScoreRequest true "New score"
// @Success 200 {object} types.UpdatePerformanceScoreResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/employee/{id}/performance [patch]
func (h *AccountHandler) UpdateEmployeePerformanceScore(c *gin.Context) {
	var req types.UpdatePerformanceScoreRequest
	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.UpdateEmployeePerformanceScore(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateEmployeeStatus godoc
// @Summary Update employee status
// @Description Set employee ACTIVE/ONDUTY/INACTIVE
// @Security BearerAuth
// @Tags Account
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param input body types.UpdateEmployeeStatusRequest true "New status"
// @Success 200 {object} types.UpdateEmployeeStatusResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/employee/{id}/status [patch]
func (h *AccountHandler) UpdateEmployeeStatus(c *gin.Context) {
	var req types.UpdateEmployeeStatusRequest
	req.ID = c.Param("id")

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.UpdateEmployeeStatus(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteEmployee godoc
// @Summary Delete an employee
// @Description Remove employee by ID
// @Security BearerAuth
// @Tags Account
// @Produce json
// @Param id path string true "Employee ID"
// @Param accId path string true "Account ID"
// @Success 200 {object} types.DeleteEmployeeResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /account/employee/{id}/{empId} [delete]
func (h *AccountHandler) DeleteEmployee(c *gin.Context) {
	id := c.Param("id")
	empId := c.Param("empId")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.DeleteEmployee(ctx, id, empId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}
