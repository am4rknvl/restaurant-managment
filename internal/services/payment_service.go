package services

import (
	"fmt"
	"restaurant-system/internal/database"
	"restaurant-system/internal/models"
	"restaurant-system/internal/payments"
	"time"

	"github.com/google/uuid"
)

type PaymentService struct {
	db *database.DB
}

func NewPaymentService(db *database.DB) *PaymentService {
	return &PaymentService{db: db}
}

func (s *PaymentService) ProcessPayment(req *models.ProcessPaymentRequest) (*models.PaymentResponse, error) {
	// Check if order exists and get its details
	var orderAmount float64
	var orderStatus string
	err := s.db.Conn().QueryRow(
		"SELECT total_amount, status FROM orders WHERE id = $1",
		req.OrderID,
	).Scan(&orderAmount, &orderStatus)
	
	if err != nil {
		return nil, fmt.Errorf("order not found")
	}
	
	if orderStatus == "completed" {
		return nil, fmt.Errorf("order already completed")
	}
	
	// Generate payment ID
	paymentID := uuid.New().String()
	transactionID := uuid.New().String()
	
	// Create payment record
	payment := &models.Payment{
		ID:            paymentID,
		OrderID:       req.OrderID,
		Amount:        orderAmount,
		Method:        req.Method,
		Status:        models.PaymentStatusProcessing,
		TransactionID: transactionID,
		PhoneNumber:   req.PhoneNumber,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	
	// Store payment in database
	_, err = s.db.Conn().Exec(
		"INSERT INTO payments (id, order_id, amount, method, status, transaction_id, phone_number, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		payment.ID, payment.OrderID, payment.Amount, payment.Method, payment.Status, payment.TransactionID, payment.PhoneNumber, payment.CreatedAt, payment.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	// Route by method
	switch req.Method {
	case models.PaymentMethodMobileMoney:
		return s.processTelebirr(payment)
	case models.PaymentMethodCash:
		return s.processCashPayment(payment)
	case models.PaymentMethodCard:
		return s.processCardPayment(payment)
	default:
		return nil, fmt.Errorf("unsupported payment method")
	}
}

func (s *PaymentService) processTelebirr(payment *models.Payment) (*models.PaymentResponse, error) {
	// Initiate Telebirr payment and return checkout URL
	initReq := &payments.InitiateRequest{
		OutTradeNo:  payment.ID,
		Subject:    "Restaurant Order",
		TotalAmount: payment.Amount,
		ReturnUrl:  "http://localhost:3000/app",
		NotifyUrl:  "http://localhost:8080/api/v1/payments/telebirr/notify",
		PhoneNumber: payment.PhoneNumber,
	}
	gwResp, err := payments.InitiatePayment(initReq)
	if err != nil {
		// Mark as failed
		_, _ = s.db.Conn().Exec("UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3", models.PaymentStatusFailed, time.Now(), payment.ID)
		return nil, err
	}
	// Stay in processing until callback confirms
	resp := &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        models.PaymentStatusProcessing,
		TransactionID: gwResp.TradeNo,
		Message:       "Redirect to complete payment",
		CheckoutURL:   gwResp.CheckoutURL,
	}
	return resp, nil
}

func (s *PaymentService) processCashPayment(payment *models.Payment) (*models.PaymentResponse, error) {
	// Cash payments are typically handled manually
	response := &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        models.PaymentStatusCompleted,
		TransactionID: payment.TransactionID,
		Message:       "Cash payment received - order confirmed",
	}
	// Update order on success
	_ = s.updateOrderAfterPayment(payment.OrderID)
	_, _ = s.db.Conn().Exec("UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3", models.PaymentStatusCompleted, time.Now(), payment.ID)
	return response, nil
}

func (s *PaymentService) processCardPayment(payment *models.Payment) (*models.PaymentResponse, error) {
	response := &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        models.PaymentStatusCompleted,
		TransactionID: payment.TransactionID,
		Message:       "Payment processed successfully via card",
	}
	_ = s.updateOrderAfterPayment(payment.OrderID)
	_, _ = s.db.Conn().Exec("UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3", models.PaymentStatusCompleted, time.Now(), payment.ID)
	return response, nil
}

func (s *PaymentService) updateOrderAfterPayment(orderID string) error {
	_, err := s.db.Conn().Exec(
		"UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3",
		models.OrderStatusConfirmed, time.Now(), orderID,
	)
	return err
}

func (s *PaymentService) GetPaymentStatus(paymentID string) (*models.Payment, error) {
	var payment models.Payment
	err := s.db.Conn().QueryRow(
		"SELECT id, order_id, amount, method, status, transaction_id, phone_number, created_at, updated_at FROM payments WHERE id = $1",
		paymentID,
	).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &payment.TransactionID, &payment.PhoneNumber, &payment.CreatedAt, &payment.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
