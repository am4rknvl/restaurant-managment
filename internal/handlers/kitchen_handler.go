package handlers

import (
	"net/http"
	"restaurant-system/internal/models"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"github.com/gin-gonic/gin"
)

type KitchenHandler struct {
	kitchenService *services.KitchenService
	hub            *websocket.Hub
}

func NewKitchenHandler(kitchenService *services.KitchenService, hub *websocket.Hub) *KitchenHandler {
	return &KitchenHandler{
		kitchenService: kitchenService,
		hub:            hub,
	}
}

func (h *KitchenHandler) GetPendingOrders(c *gin.Context) {
	orders, err := h.kitchenService.GetPendingOrders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *KitchenHandler) UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.kitchenService.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated order
	order, err := h.kitchenService.GetOrderDetails(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get updated order"})
		return
	}

	// Notify all connected clients about status update
	h.hub.Broadcast(gin.H{
		"type": "order_status_updated",
		"data": order,
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"order":   order,
	})
}

func (h *KitchenHandler) GetOrderDetails(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	order, err := h.kitchenService.GetOrderDetails(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}
