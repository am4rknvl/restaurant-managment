package handlers

import (
	"net/http"
	"restaurant-system/internal/models"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *services.OrderService
	hub          *websocket.Hub
}

func NewOrderHandler(orderService *services.OrderService, hub *websocket.Hub) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		hub:          hub,
	}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Notify kitchen dashboard via WebSocket
	h.hub.BroadcastToKitchen(gin.H{
		"type": "new_order",
		"data": order,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})
}

func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order ID is required"})
		return
	}

	order, err := h.orderService.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *OrderHandler) GetOrders(c *gin.Context) {
	customerID := c.Query("customer_id")

	orders, err := h.orderService.GetOrders(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
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

	err := h.orderService.UpdateOrderStatus(orderID, req.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get updated order
	order, err := h.orderService.GetOrder(orderID)
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

func (h *OrderHandler) GetMenuItems(c *gin.Context) {
	items, err := h.orderService.GetMenuItems()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"menu_items": items})
}
