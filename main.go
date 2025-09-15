package main

import (
	"log"
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

	// Serve static files for frontend
	router.Static("/static", "./web/static")
	router.LoadHTMLGlob("web/templates/*")

	// Serve frontend pages
	router.GET("/", func(c *gin.Context) {
		c.HTML(200, "customer.html", nil)
	})

	router.GET("/kitchen", func(c *gin.Context) {
		c.HTML(200, "kitchen.html", nil)
	})

	log.Println("Starting restaurant system server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
