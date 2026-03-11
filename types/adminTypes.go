package types

type AdminDashboardRequest struct {
	AdminID    string `json:"adminId"     binding:"required"`
	DateFilter string `json:"dateFilter"  binding:"required"`
}
type DashboardData struct {
	Sales          int32 `json:"sales"`
	Bookings       int32 `json:"bookings"`
	ActiveSessions int32 `json:"activeSessions"`
	Clients        int32 `json:"clients"`
}
type AdminDashboardResponse struct {
	Sales          int32       `json:"sales"`
	Bookings       int32       `json:"bookings"`
	ActiveSessions int32       `json:"activeSessions"`
	Clients        int32       `json:"clients"`
	GrowthIndex    GrowthIndex `json:"growthIndex"`
}

type GrowthIndex struct {
	SalesGrowthIndex          float64 `json:"salesGrowthIndex"`
	BookingsGrowthIndex       float64 `json:"bookingsGrowthIndex"`
	ActiveSessionsGrowthIndex float64 `json:"activeSessionsGrowthIndex"`
}
type ClientsGrowthIndex struct {
	New         int32   `json:"new"`
	Returning   int32   `json:"returning"`
	Inactive    int32   `json:"inactive"`
	GrowthIndex float64 `json:"growthIndex"`
}

type OnboardEmployeeRequest struct {
	Role           string `json:"role" binding:"required"`
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	OrganizationID string `json:"organizationId" binding:"required"`
	HireDate       string `json:"hireDate" binding:"required"`
	Position       string `json:"position" binding:"required"`
}
type AcceptBookingRequest struct {
	BookingID string `json:"bookingId" binding:"required"`
}

type AcceptBookingResponse struct {
	BookingID string `json:"bookingId"`
	Status    string `json:"status"`
}

// ItemQuantity maps an inventory item ID to how much was used
type ItemQuantity struct {
	ItemID   string  `json:"itemId" binding:"required"`
	Quantity float64 `json:"quantity" binding:"required"`
}

type AssignResourcesToBookingRequest struct {
	BookingID string         `json:"bookingId" binding:"required"`
	Resources []ItemQuantity `json:"resources" binding:"required"`
}

type AssignEquipmentToBookingRequest struct {
	BookingID string         `json:"bookingId" binding:"required"`
	Equipment []ItemQuantity `json:"equipment" binding:"required"`
}

type AssignInventoryResponse struct {
	BookingID string `json:"bookingId"`
	Message   string `json:"message"`
}
