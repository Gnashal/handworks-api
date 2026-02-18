package main

import (
	"context"
	"handworks-api/config"
	"handworks-api/endpoints"
	"handworks-api/handlers"
	"handworks-api/middleware"
	"handworks-api/realtime"
	listeners "handworks-api/realtime/listeners"
	"handworks-api/services"
	"handworks-api/utils"
	"os"

	_ "handworks-api/docs"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Handworks API
// @version 1.0
// @description This is the official API documentation for the Handworks Api.
// @host localhost:8080
// @BasePath /api/

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter "Bearer <your_token>"
func main() {
	_ = godotenv.Load()
	c := context.Background()
	logger, err := utils.NewLogger()
	if err != nil {
		panic(err)
	}
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Fatal("Failed to set trusted proxies: %v", err)
	}
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))
	router.Use(cors.New(config.NewCors()))
	conn, err := config.InitDB(logger, c)
	if err != nil {
		logger.Fatal("DB init failed: %v", err)
	}
	logger.Info("Connected to Db")
	defer conn.Close()

	// public paths for Clerk middleware
	publicPaths := []string{"/api/account/customer/signup",
		"/api/account/employee/signup", "/api/account/admin/signup",
		"/api/payment/quote/preview", "/health", "/api/admin/dashboard", "/api/payment/quote"}

	// websocket
	hubs := realtime.NewRealtimeHubs(logger)

	router.Use(middleware.ClerkAuthMiddleware(publicPaths, logger))

	accountService := services.NewAccountService(conn, logger)
	inventoryService := services.NewInventoryService(conn, logger)
	paymentService := services.NewPaymentService(conn, logger)
	bookingService := services.NewBookingService(conn, logger, paymentService)
	adminServie := services.NewAdminService(conn, logger)

	accountHandler := handlers.NewAccountHandler(accountService, logger)
	inventoryHandler := handlers.NewInventoryHandler(inventoryService, logger)
	bookingHandler := handlers.NewBookingHandler(bookingService, logger)
	paymentHandler := handlers.NewPaymentHandler(paymentService, logger)
	adminHandler := handlers.NewAdminHandler(adminServie, logger)

	api := router.Group("/api")
	{
		endpoints.AccountEndpoint(api.Group("/account"), accountHandler)
		endpoints.InventoryEndpoint(api.Group("/inventory"), inventoryHandler)
		endpoints.BookingEndpoint(api.Group("/booking"), bookingHandler)
		endpoints.PaymentEndpoint(api.Group("/payment"), paymentHandler)
		endpoints.AdminEndpoint(api.Group("/admin"), adminHandler)
		endpoints.RealtimeEndpoint(api, hubs)
	}

	// running websocket hubs
	go hubs.EmployeeHub.Run()
	go hubs.AdminHub.Run()
	go hubs.ChatHub.Run()

	// listeners
	listener := listeners.NewListener(
		c,
		logger,
		hubs,
		os.Getenv("DB_CONN_REALTIME"),
		bookingService,
		inventoryService,
	)
	// listener events
	go func() {
		if err := listener.Start(); err != nil {
			logger.Fatal("Listener failed: %v", err)
		}
	}()

	port := "8080"
	logger.Info("Starting server on port %s", port)
	logger.Info("Swagger on localhost:8080/swagger/index.html")
	if err := router.Run(":" + port); err != nil {
		logger.Fatal("Server failed: %v", err)
	}
}
