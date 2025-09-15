package models

import (
	"time"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusPreparing OrderStatus = "preparing"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID          string      `json:"id" db:"id"`
	CustomerID  string      `json:"customer_id" db:"customer_id"`
	Items       []OrderItem `json:"items" db:"items"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`
	Status      OrderStatus `json:"status" db:"status"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

type OrderItem struct {
	ID         string  `json:"id" db:"id"`
	OrderID    string  `json:"order_id" db:"order_id"`
	MenuItemID string  `json:"menu_item_id" db:"menu_item_id"`
	Name       string  `json:"name" db:"name"`
	Price      float64 `json:"price" db:"price"`
	Quantity   int     `json:"quantity" db:"quantity"`
	TotalPrice float64 `json:"total_price" db:"total_price"`
}

type MenuItem struct {
	ID          string  `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Description string  `json:"description" db:"description"`
	Price       float64 `json:"price" db:"price"`
	Category    string  `json:"category" db:"category"`
	Available   bool    `json:"available" db:"available"`
}

type CreateOrderRequest struct {
	CustomerID string            `json:"customer_id" binding:"required"`
	Items      []CreateOrderItem `json:"items" binding:"required"`
}

type CreateOrderItem struct {
	MenuItemID string `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}
