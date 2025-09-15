package handlers

import (
	"net/http"
	"restaurant-system/internal/models"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
	hub            *websocket.Hub
}

func NewPaymentHandler(paymentService *services.PaymentService, hub *websocket.Hub) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		hub:            hub,
	}
}

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	var req models.ProcessPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.paymentService.ProcessPayment(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Notify all connected clients about payment status
	h.hub.Broadcast(gin.H{
		"type": "payment_processed",
		"data": response,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment initiated",
		"payment": response,
	})
}

func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentID := c.Param("id")
	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment ID is required"})
		return
	}

	payment, err := h.paymentService.GetPaymentStatus(paymentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}
