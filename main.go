package main

import (
	"log"
	"restaurant-system/internal/config"
	"restaurant-system/internal/database"
	"restaurant-system/internal/handlers"
	"restaurant-system/internal/services"
	"restaurant-system/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default values")
	}
	config.Load()

	// Initialize database
	db, err := database.Initialize()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize services
	orderService := services.NewOrderService(db)
	paymentService := services.NewPaymentService(db)
	accountService := services.NewAccountService(db)
	kitchenService := services.NewKitchenService(db)
	authService := services.NewAuthService(db)

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()

	// Initialize handlers
	orderHandler := handlers.NewOrderHandler(orderService, hub)
	paymentHandler := handlers.NewPaymentHandler(paymentService, hub)
	accountHandler := handlers.NewAccountHandler(accountService)
	kitchenHandler := handlers.NewKitchenHandler(kitchenService, hub)
	authHandler := handlers.NewAuthHandler(authService)

	// Setup router
	router := gin.Default()

	// Share services in context
	router.Use(func(c *gin.Context) {
		c.Set("paymentService", paymentService)
		c.Next()
	})

	// Enable CORS for frontend
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/request-otp", authHandler.RequestOTP)
			auth.POST("/verify-otp", authHandler.VerifyOTP)
		}

		// Order routes
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.GET("", orderHandler.GetOrders)
			orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
		}

		// Payment routes
		payments := api.Group("/payments")
		{
			payments.POST("", paymentHandler.ProcessPayment)
			payments.GET("/:id", paymentHandler.GetPaymentStatus)
			payments.POST("/notify/telebirr", handlers.TelebirrNotifyHandler)
		}

		// Account routes
		accounts := api.Group("/accounts")
		{
			accounts.GET("/:id/balance", accountHandler.GetAccountBalance)
			accounts.POST("", accountHandler.CreateAccount)
		}

		// Kitchen routes
		kitchen := api.Group("/kitchen")
		{
			kitchen.GET("/orders", kitchenHandler.GetPendingOrders)
			kitchen.PUT("/orders/:id/status", kitchenHandler.UpdateOrderStatus)
		}

		// WebSocket route
		api.GET("/ws", func(c *gin.Context) {
			websocket.HandleWebSocket(hub, c.Writer, c.Request)
		})
	}

	// Serve static files (optional; Next.js runs separately)
	router.Static("/static", "./web/static")

	log.Println("Starting restaurant system server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
