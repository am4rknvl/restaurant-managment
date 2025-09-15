package services

import (
	"fmt"
	"restaurant-system/internal/database"
	"restaurant-system/internal/models"
	"time"

	"github.com/google/uuid"
)

type OrderService struct {
	db *database.DB
}

func NewOrderService(db *database.DB) *OrderService {
	return &OrderService{db: db}
}

func (s *OrderService) CreateOrder(req *models.CreateOrderRequest) (*models.Order, error) {
	// Generate order ID
	orderID := uuid.New().String()

	// Calculate total amount and prepare order items
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		// Get menu item details
		var menuItem models.MenuItem
		err := s.db.Conn().QueryRow(
			"SELECT id, name, price FROM menu_items WHERE id = $1 AND available = TRUE",
			item.MenuItemID,
		).Scan(&menuItem.ID, &menuItem.Name, &menuItem.Price)

		if err != nil {
			return nil, fmt.Errorf("menu item not found or unavailable: %s", item.MenuItemID)
		}

		itemTotal := menuItem.Price * float64(item.Quantity)
		totalAmount += itemTotal

		orderItems = append(orderItems, models.OrderItem{
			ID:         uuid.New().String(),
			OrderID:    orderID,
			MenuItemID: item.MenuItemID,
			Name:       menuItem.Name,
			Price:      menuItem.Price,
			Quantity:   item.Quantity,
			TotalPrice: itemTotal,
		})
	}

	// Create order
	order := &models.Order{
		ID:          orderID,
		CustomerID:  req.CustomerID,
		Items:       orderItems,
		TotalAmount: totalAmount,
		Status:      models.OrderStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Store order in database
	_, err := s.db.Conn().Exec(
		"INSERT INTO orders (id, customer_id, total_amount, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		order.ID, order.CustomerID, order.TotalAmount, order.Status, order.CreatedAt, order.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Store order items
	for _, item := range orderItems {
		_, err := s.db.Conn().Exec(
			"INSERT INTO order_items (id, order_id, menu_item_id, name, price, quantity, total_price) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			item.ID, item.OrderID, item.MenuItemID, item.Name, item.Price, item.Quantity, item.TotalPrice,
		)
		if err != nil {
			return nil, err
		}
	}

	return order, nil
}

func (s *OrderService) GetOrder(orderID string) (*models.Order, error) {
	var order models.Order
	err := s.db.Conn().QueryRow(
		"SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE id = $1",
		orderID,
	).Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, err
	}

	// Get order items
	order.Items, err = s.db.GetOrderItems(orderID)
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (s *OrderService) GetOrders(customerID string) ([]*models.Order, error) {
	var query string
	var args []interface{}

	if customerID != "" {
		query = "SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders WHERE customer_id = $1 ORDER BY created_at DESC"
		args = []interface{}{customerID}
	} else {
		query = "SELECT id, customer_id, total_amount, status, created_at, updated_at FROM orders ORDER BY created_at DESC"
	}

	rows, err := s.db.Conn().Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.CustomerID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}

		// Get order items
		order.Items, err = s.db.GetOrderItems(order.ID)
		if err != nil {
			return nil, err
		}

		orders = append(orders, &order)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(orderID string, status models.OrderStatus) error {
	_, err := s.db.Conn().Exec(
		"UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3",
		status, time.Now(), orderID,
	)
	return err
}

func (s *OrderService) GetMenuItems() ([]*models.MenuItem, error) {
	rows, err := s.db.Conn().Query("SELECT id, name, description, price, category, available FROM menu_items WHERE available = TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*models.MenuItem
	for rows.Next() {
		var item models.MenuItem
		err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.Price, &item.Category, &item.Available)
		if err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, nil
}
