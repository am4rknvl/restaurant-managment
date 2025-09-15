package handlers

import (
	"net/http"
	"restaurant-system/internal/services"

	"github.com/gin-gonic/gin"
)

// TelebirrNotifyHandler handles Telebirr payment gateway callbacks
func TelebirrNotifyHandler(c *gin.Context) {
	var payload map[string]string
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// PaymentService is stored in context by DI in main, or we can resolve via a package var.
	svc, ok := c.MustGet("paymentService").(*services.PaymentService)
	if !ok || svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}

	if err := svc.HandleTelebirrCallback(payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
