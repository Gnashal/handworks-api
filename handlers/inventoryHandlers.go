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

// CreateItem godoc
// @Summary Create a new inventory item
// @Description Adds a new item to inventory
// @Security BearerAuth
// @Tags Inventory
// @Accept json
// @Produce json
// @Param input body types.CreateItemRequest true "Item info"
// @Success 200 {object} types.InventoryItem
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /inventory [post]
func (h *InventoryHandler) CreateItem(c *gin.Context) {
	var req types.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.CreateItem(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetItem godoc
// @Summary Get an item by ID
// @Description Retrieve a single inventory item
// @Security BearerAuth
// @Tags Inventory
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} types.InventoryItem
// @Failure 404 {object} types.ErrorResponse
// @Router /inventory/{id} [get]
func (h *InventoryHandler) GetItem(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.GetItem(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetItems godoc
// @Summary Get inventory items
// @Description Retrieve inventory items with optional filters and pagination
// @Security BearerAuth
// @Tags Inventory
// @Produce json
//
// @Param type query string false "Item type"
// @Param status query string false "Item status"
// @Param category query string false "Item category"
// @Param page query int false "Page number (zero-based)" default(0)
// @Param limit query int false "Number of items per page" default(10)
//
// @Success 200 {object} types.InventoryListResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 401 {object} types.ErrorResponse
// @Failure 404 {object} types.ErrorResponse
// @Router /inventory/items [get]
func (h *InventoryHandler) GetItems(c *gin.Context) {
	itemType := c.Query("type")
	status := c.Query("status")
	category := c.Query("category")

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

	// build filter
	var f types.InventoryFilter
	if itemType != "" {
		f.Type = &itemType
	}
	if status != "" {
		f.Status = &status
	}
	if category != "" {
		f.Category = &category
	}

	f.Page = &page
	f.Limit = &limit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.ListItems(ctx, &f)
	if err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateItem godoc
// @Summary Update an inventory item
// @Description Modify fields of an existing inventory item
// @Security BearerAuth
// @Tags Inventory
// @Accept json
// @Produce json
// @Param input body types.UpdateItemRequest true "Updated item info"
// @Success 200 {object} types.InventoryItem
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /inventory/ [put]
func (h *InventoryHandler) UpdateItem(c *gin.Context) {
	var req types.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.UpdateItem(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteItem godoc
// @Summary Delete an item
// @Description Remove inventory item by ID
// @Security BearerAuth
// @Tags Inventory
// @Produce json
// @Param id path string true "Item ID"
// @Success 200 {object} []types.InventoryItem
// @Failure 500 {object} types.ErrorResponse
// @Router /inventory/{id} [delete]
func (h *InventoryHandler) DeleteItem(c *gin.Context) {
	id := c.Param("id")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := h.Service.DeleteItem(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)
}
