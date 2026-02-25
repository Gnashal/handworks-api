package endpoints

import (
	"handworks-api/handlers"
	"handworks-api/realtime"

	"github.com/gin-gonic/gin"
)

func AccountEndpoint(r *gin.RouterGroup, h *handlers.AccountHandler) {
	customer := r.Group("/customer")
	{
		customer.POST("/signup", h.SignUpCustomer)
		customer.GET("/:id", h.GetCustomer)
		customer.GET("/", h.GetCustomers)
		customer.PUT("/:id", h.UpdateCustomer)
		customer.DELETE("/:id/:accId", h.DeleteCustomer)

	}

	employee := r.Group("/employee")
	{
		employee.POST("/signup", h.SignUpEmployee)
		employee.GET("/:id", h.GetEmployee)
		employee.GET("/", h.GetEmployees)
		employee.PUT("/:id", h.UpdateEmployee)
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
	r.GET("/:id", h.GetItem)
	r.GET("/items", h.GetItems)
	r.PUT("/", h.UpdateItem)
	r.DELETE("/:id", h.DeleteItem)
}
func BookingEndpoint(r *gin.RouterGroup, h *handlers.BookingHandler) {
	r.POST("/", h.CreateBooking)
	r.GET("/", h.GetBookingById)
	r.GET("/bookings", h.GetBookings)
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
}
func PaymentEndpoint(r *gin.RouterGroup, h *handlers.PaymentHandler) {
	r.POST("/quote", h.MakeQuotation)
	r.POST("/quote/preview", h.MakePublicQuotation)
	r.GET("/quotes", h.GetAllQuotes)
	customer := r.Group("/customer")
	{
		customer.GET("/quotes", h.GetAllQuotesFromCustomer)
	}
}

func AdminEndpoint(r *gin.RouterGroup, h *handlers.AdminHandler) {
	r.GET("/dashboard", h.GetAdminDashboard)
	employees := r.Group("/employee")
	{
		employees.POST("/onboard", h.OnboardEmployee)
	}
}

func RealtimeEndpoint(r *gin.RouterGroup,hubs * realtime.RealtimeHubs) {
	r.GET("/ws/admin", realtime.AdminWS(hubs.AdminHub))
	r.GET("/ws/employee", realtime.EmployeeWS(hubs.EmployeeHub))
	r.GET("/ws/chat", realtime.ChatWS(hubs.ChatHub))
}
