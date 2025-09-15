package services

import (
	"fmt"
	"os"
	"restaurant-system/internal/database"
	"restaurant-system/internal/models"
	"restaurant-system/internal/payments"
	"strings"
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

	if orderStatus == string(models.OrderStatusCompleted) {
		return nil, fmt.Errorf("order already completed")
	}

	// Generate payment ID
	paymentID := uuid.New().String()

	// Create payment record (initial status processing)
	payment := &models.Payment{
		ID:            paymentID,
		OrderID:       req.OrderID,
		Amount:        orderAmount,
		Method:        req.Method,
		Status:        models.PaymentStatusProcessing,
		TransactionID: "",
		PhoneNumber:   req.PhoneNumber,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if _, err = s.db.Conn().Exec(
		"INSERT INTO payments (id, order_id, amount, method, status, transaction_id, phone_number, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
		payment.ID, payment.OrderID, payment.Amount, payment.Method, payment.Status, payment.TransactionID, payment.PhoneNumber, payment.CreatedAt, payment.UpdatedAt,
	); err != nil {
		return nil, err
	}

	// Route by method
	switch req.Method {
	case models.PaymentMethodMobileMoney:
		return s.initiateTelebirr(payment)
	case models.PaymentMethodCash:
		resp := s.processCashPayment(payment)
		// Update payment status
		_, err = s.db.Conn().Exec("UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3", resp.Status, time.Now(), resp.ID)
		if err != nil {
			return nil, err
		}
		// Update order if completed
		if resp.Status == models.PaymentStatusCompleted {
			if err := s.updateOrderAfterPayment(payment.OrderID); err != nil {
				return nil, err
			}
		}
		return resp, nil
	case models.PaymentMethodCard:
		resp := s.processCardPayment(payment)
		_, err = s.db.Conn().Exec("UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3", resp.Status, time.Now(), resp.ID)
		if err != nil {
			return nil, err
		}
		if resp.Status == models.PaymentStatusCompleted {
			if err := s.updateOrderAfterPayment(payment.OrderID); err != nil {
				return nil, err
			}
		}
		return resp, nil
	default:
		return nil, fmt.Errorf("unsupported payment method")
	}
}

func (s *PaymentService) initiateTelebirr(payment *models.Payment) (*models.PaymentResponse, error) {
	notifyURL := os.Getenv("PUBLIC_NOTIFY_URL")
	if notifyURL == "" {
		// fallback to localhost for dev
		notifyURL = "http://localhost:8080/api/v1/payments/notify/telebirr"
	}
	returnURL := os.Getenv("PUBLIC_RETURN_URL")
	if returnURL == "" {
		returnURL = "http://localhost:3000/app"
	}

	gwResp, err := payments.InitiatePayment(&payments.InitiateRequest{
		OutTradeNo:  payment.ID,
		Subject:     "Restaurant Order " + payment.OrderID,
		TotalAmount: payment.Amount,
		ReturnUrl:   returnURL,
		NotifyUrl:   notifyURL,
		PhoneNumber: payment.PhoneNumber,
	})
	if err != nil {
		// Mark payment failed
		_, _ = s.db.Conn().Exec("UPDATE payments SET status = $1, updated_at = $2 WHERE id = $3", models.PaymentStatusFailed, time.Now(), payment.ID)
		return nil, err
	}

	// Store gateway trade number
	_, _ = s.db.Conn().Exec("UPDATE payments SET transaction_id = $1, status = $2, updated_at = $3 WHERE id = $4", gwResp.TradeNo, models.PaymentStatusPending, time.Now(), payment.ID)

	return &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        models.PaymentStatusPending,
		TransactionID: gwResp.TradeNo,
		Message:       "Proceed to Telebirr to complete payment",
		CheckoutURL:   gwResp.CheckoutURL,
	}, nil
}

func (s *PaymentService) processMobileMoneyPayment(payment *models.Payment) *models.PaymentResponse {
	// Deprecated in favor of initiateTelebirr
	return &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        models.PaymentStatusProcessing,
		TransactionID: payment.TransactionID,
		Message:       "Initiated",
	}
}

func (s *PaymentService) processCashPayment(payment *models.Payment) *models.PaymentResponse {
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

	return response
}

func (s *PaymentService) processCardPayment(payment *models.Payment) *models.PaymentResponse {
	// Simulate card payment processing
	response := &models.PaymentResponse{
		ID:            payment.ID,
		OrderID:       payment.OrderID,
		Amount:        payment.Amount,
		Method:        payment.Method,
		Status:        models.PaymentStatusCompleted,
		TransactionID: payment.TransactionID,
		Message:       "Payment processed successfully via card",
	}

	return response
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

// HandleTelebirrCallback updates payment and order based on Telebirr callback payload.
// Expected fields include at least one of: outTradeNo (our payment ID), tradeNo (gateway id),
// and a status/result code (e.g., SUCCESS).
func (s *PaymentService) HandleTelebirrCallback(data map[string]string) error {
	// Extract identifiers
	outTradeNo := data["outTradeNo"]
	if outTradeNo == "" {
		outTradeNo = data["out_trade_no"]
	}
	tradeNo := data["tradeNo"]
	if tradeNo == "" {
		tradeNo = data["trade_no"]
	}

	// Determine success
	status := data["status"]
	if status == "" {
		status = data["result"]
	}
	success := false
	switch strings.ToUpper(status) {
	case "SUCCESS", "SUCCEED", "COMPLETED", "PAID":
		success = true
	}

	// Find payment
	var payment models.Payment
	var err error
	if outTradeNo != "" {
		paymentPtr, e := s.GetPaymentStatus(outTradeNo)
		if e == nil && paymentPtr != nil {
			payment = *paymentPtr
		} else {
			err = e
		}
	}
	if payment.ID == "" && tradeNo != "" {
		// Lookup by transaction_id
		e := s.db.Conn().QueryRow(
			"SELECT id, order_id, amount, method, status, transaction_id, phone_number, created_at, updated_at FROM payments WHERE transaction_id = $1",
			tradeNo,
		).Scan(&payment.ID, &payment.OrderID, &payment.Amount, &payment.Method, &payment.Status, &payment.TransactionID, &payment.PhoneNumber, &payment.CreatedAt, &payment.UpdatedAt)
		if e != nil {
			return e
		}
	} else if payment.ID == "" && err != nil {
		return err
	}

	// Update payment status
	newStatus := models.PaymentStatusFailed
	if success {
		newStatus = models.PaymentStatusCompleted
	}
	if tradeNo != "" && payment.TransactionID == "" {
		payment.TransactionID = tradeNo
	}
	if _, e := s.db.Conn().Exec(
		"UPDATE payments SET status = $1, transaction_id = COALESCE(NULLIF($2,''), transaction_id), updated_at = $3 WHERE id = $4",
		newStatus, payment.TransactionID, time.Now(), payment.ID,
	); e != nil {
		return e
	}

	// Update order on success
	if success {
		if e := s.updateOrderAfterPayment(payment.OrderID); e != nil {
			return e
		}
	}
	return nil
}
