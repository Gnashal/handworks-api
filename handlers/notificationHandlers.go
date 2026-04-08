package handlers

import (
	"context"
	"handworks-api/types"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SubscribeToken godoc
// @Summary Subscribe a device token to notification topics
// @Description Subscribes an FCM token to admin or employee topic based on role
// @Tags Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.SubscribeNotificationRequest true "Subscription request"
// @Success 200 {object} types.SubscribeNotificationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /notifications/subscribe [post]
func (h *NotificationHandler) SubscribeToken(c *gin.Context) {
	var req types.SubscribeNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.SubscribeToken(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}

// UnsubscribeToken godoc
// @Summary Unsubscribe a device token from notification topics
// @Description Deactivates token persistence and removes topic subscription for admin or employee role
// @Tags Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param input body types.UnsubscribeNotificationRequest true "Unsubscription request"
// @Success 200 {object} types.UnsubscribeNotificationResponse
// @Failure 400 {object} types.ErrorResponse
// @Failure 500 {object} types.ErrorResponse
// @Router /notifications/unsubscribe [post]
func (h *NotificationHandler) UnsubscribeToken(c *gin.Context) {
	var req types.UnsubscribeNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := h.Service.UnsubscribeToken(ctx, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, res)
}
