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
		customer.PUT("/:id", h.UpdateCustomer)
		// Route should be like this in your router:
		customer.DELETE("/:id/:accId", h.DeleteCustomer)

	}

	employee := r.Group("/employee")
	{
		employee.POST("/signup", h.SignUpEmployee)
		employee.GET("/:id", h.GetEmployee)
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
	r.GET("/bookings", h.GetBookings)
	r.GET("/", h.GetCustomerBookings)
	r.PUT("/:id", h.UpdateBooking)
	r.DELETE("/:id", h.DeleteBooking)
}
func PaymentEndpoint(r *gin.RouterGroup, h *handlers.PaymentHandler) {
	r.POST("/quote", h.MakeQuotation)
	r.POST("/quote/preview", h.MakePublicQuotation)
	r.GET("/quotes", h.GetAllQuotesFromCustomer)
	r.GET("/quote/:id", h.GetQuoteByIDForCustomer)

}

func RealtimeEndpoint(r *gin.RouterGroup, hub *realtime.AdminHub) {
	// admin websocket endpoint
	r.GET("/ws/admin", realtime.AdminWS(hub))
}
