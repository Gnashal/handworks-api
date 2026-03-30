package endpoints

import (
	"handworks-api/handlers"
	"handworks-api/realtime"

	"github.com/gin-gonic/gin"
)

func AccountEndpoint(r *gin.RouterGroup, h *handlers.AccountHandler) {
	customer := r.Group("/customer")
	{
		customer.GET("/", h.GetCustomer)
		customer.GET("/customers", h.GetCustomers)
		customer.POST("/signup", h.SignUpCustomer)
		customer.PUT("/:id", h.UpdateCustomer)
		customer.DELETE("/:id/:accId", h.DeleteCustomer)

	}

	employee := r.Group("/employee")
	{
		timesheet := employee.Group("/timesheet")
		{
			timesheet.POST("/timein", h.EmployeeTimeIn)
			timesheet.POST("/timeout", h.EmployeeTimeOut)
			timesheet.GET("/today", h.TimesheetToday)
		}
		employee.POST("/signup", h.SignUpEmployee)
		employee.GET("/", h.GetEmployee)
		employee.GET("/employees", h.GetEmployees)
		employee.PUT("/", h.UpdateEmployee)
		employee.PUT("/:id/performance", h.UpdateEmployeePerformanceScore)
		employee.PUT("/:id/status", h.UpdateEmployeeStatus)
		employee.DELETE("/:id/:empId", h.DeleteEmployee)
	}
	admin := r.Group("/admin")
	{
		admin.POST("/signup", h.SignUpAdmin)
	}
}
func InventoryEndpoint(r *gin.RouterGroup, h *handlers.InventoryHandler) {
	r.POST("/create", h.CreateItem)
	r.GET("/", h.GetItem)
	r.GET("/items", h.GetItems)
	r.PUT("/", h.UpdateItem)
	r.DELETE("/:id", h.DeleteItem)
}
func BookingEndpoint(r *gin.RouterGroup, h *handlers.BookingHandler) {
	r.GET("/", h.GetBookingById)
	r.GET("/bookings", h.GetBookings)
	r.GET("/slots", h.GetBookedSlots)
	r.POST("/", h.CreateBooking)
	r.PUT("/:id", h.UpdateBooking)
	r.DELETE("/:id", h.DeleteBooking)
	customers := r.Group("/customer")
	{
		customers.GET("/", h.GetCustomerBookings)
	}
	employees := r.Group("/employee")
	{
		employees.GET("/", h.GetEmployeeAssignedBookings)
	}
	session := r.Group("/session")
	{
		session.POST("/start", h.StartSession)
		session.POST("/end", h.EndSession)
	}
}
func PaymentEndpoint(r *gin.RouterGroup, h *handlers.PaymentHandler) {
	quote := r.Group("/quote")
	{
		quote.GET("/", h.GetQuoteByIDForCustomer)
		quote.GET("/quotes", h.GetAllQuotes)
		quote.POST("/", h.MakeQuotation)
		quote.POST("/preview", h.MakePublicQuotation)
	}
	customer := r.Group("/customer")
	{
		customer.GET("/quotes", h.GetAllQuotesFromCustomer)
	}
	order := r.Group("/order")
	{
		order.POST("/", h.CreateOrder)
		order.GET("/", h.GetOrder)
		order.GET("/orders", h.GetOrders)
		order.GET("/customer", h.GetOrderByCustomer)
		// order.PATCH("/:id", h.UpdateOrderPaymentStatus)
	}
	payments := r.Group("/payments")
	{
		payments.GET("/order", h.GetPaymentsByOrderID)
		payments.GET("/customer", h.GetPaymentsByCustomerID)
		payments.GET("/existing-downpayment", h.HasExistingDownpayment)
		intents := payments.Group("/intent")
		{
			intents.POST("/downpayment/:id", h.CreateDownpaymentIntent)
			intents.POST("/fullpayment/:id", h.CreateFullPaymentIntent)
			intents.POST("/qrph-static", h.CreateStaticQRPHCode)
		}
	}
	webhooks := r.Group("/webhooks")
	{
		webhooks.POST("/paymongo", h.HandlePaymongoWebhook)
	}
}

func AdminEndpoint(r *gin.RouterGroup, h *handlers.AdminHandler) {
	r.GET("/dashboard", h.GetAdminDashboard)
	employees := r.Group("/employee")
	{
		employees.POST("/onboard", h.OnboardEmployee)
		// employees.POST("/assign-to-booking", h.AssignEmployeeToBooking)
	}
	bookings := r.Group("/booking")
	{
		bookings.POST("/approve/:id", h.AcceptBooking)
	}
	inventory := r.Group("/inventory")
	{
		inventory.POST("/assign-resources", h.AssignResourcesToBooking)
		inventory.POST("/assign-equipment", h.AssignEquipmentToBooking)
	}
}

func RealtimeEndpoint(r *gin.RouterGroup, hubs *realtime.RealtimeHubs) {
	r.GET("/ws/admin", realtime.AdminWS(hubs.AdminHub))
	r.GET("/ws/employee", realtime.EmployeeWS(hubs.EmployeeHub))
	r.GET("/ws/chat", realtime.ChatWS(hubs.ChatHub))
}
