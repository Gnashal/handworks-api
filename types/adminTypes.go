package types

import "time"

type AdminDashboardRequest struct {
	AdminID    string `json:"adminId"     binding:"required"`
	DateFilter string `json:"dateFilter"`
}

type DashboardData struct {
	Sales          int32 `json:"sales"`
	Bookings       int32 `json:"bookings"`
	ActiveSessions int32 `json:"activeSessions"`
	Clients        int32 `json:"clients"`
}

type RevenueSummary struct {
	Revenue float64 `json:"revenue"`
	Paid    float64 `json:"paid"`
	Unpaid  float64 `json:"unpaid"`
}

type InventoryAlert struct {
	ID    int32  `json:"id"`
	Name  string `json:"name"`
	Stock int32  `json:"stock"`
}

type TopService struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Bookings int32  `json:"bookings"`
}

type RecentActivity struct {
	ID    int32  `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	Time  string `json:"time"`
}

type AdminDashboardResponse struct {
	// Legacy fields kept during transition.
	Sales          int32       `json:"sales"`
	Bookings       int32       `json:"bookings"`
	ActiveSessions int32       `json:"activeSessions"`
	Clients        int32       `json:"clients"`
	GrowthIndex    GrowthIndex `json:"growthIndex"`

	TodayBookings    int32            `json:"todayBookings"`
	PendingActions   int32            `json:"pendingActions"`
	Revenue          float64          `json:"revenue"`
	Paid             float64          `json:"paid"`
	Unpaid           float64          `json:"unpaid"`
	ActiveClients    int32            `json:"activeClients"`
	NewClients       int32            `json:"newClients"`
	ReturningClients int32            `json:"returningClients"`
	InactiveClients  int32            `json:"inactiveClients"`
	EmployeesActive  int32            `json:"employeesActive"`
	EmployeesTotal   int32            `json:"employeesTotal"`
	LowStockItems    []InventoryAlert `json:"lowStockItems"`
	UnreadMessages   int32            `json:"unreadMessages"`
	RecentActivities []RecentActivity `json:"recentActivities"`
	TopServices      []TopService     `json:"topServices"`
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

type CalendarBooking struct {
	ID       string `json:"id" binding:"required"`
	Service  string `json:"service" binding:"required"`
	Schedule *struct {
		Date string `json:"date" binding:"required"`
		Time string `json:"time" binding:"required"`
	} `json:"schedule,omitempty"`
	Customer *struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	} `json:"customer,omitempty"`
}

type CalendarBookingResponse struct {
	Bookings []CalendarBooking `json:"bookings"`
}

type BookingTrendPoint struct {
	Label string `json:"label"`
	Value int32  `json:"value"`
}

type BookingTrendsResponse struct {
	WeeklyData  []BookingTrendPoint `json:"weeklyData"`
	MonthlyData []BookingTrendPoint `json:"monthlyData"`
}

type AssignEmployeeAction string

const (
	AssignEmployeeActionAdd    AssignEmployeeAction = "ADD"
	AssignEmployeeActionRemove AssignEmployeeAction = "REMOVE"
)

type AssignEmployeeToBookingRequest struct {
	BookingID  string               `json:"bookingId" binding:"required"`
	EmployeeID string               `json:"employeeId" binding:"required"`
	Action     AssignEmployeeAction `json:"action" binding:"required,oneof=ADD REMOVE"`
}

type AssignEmployeeToBookingResponse struct {
	BookingID  string `json:"bookingId"`
	EmployeeID string `json:"employeeId"`
	Action     string `json:"action"`
	Message    string `json:"message"`
}

type AvailableCleaner struct {
	EmployeeID string `json:"employeeId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	PFPUrl     string `json:"pfpUrl"`
}

type AvailableCleanersRequest struct {
	BookingID string `form:"bookingId" binding:"required"`
}

type AvailableCleanersResponse struct {
	BookingID    string             `json:"bookingId"`
	StartSched   time.Time          `json:"startSched"`
	EndSched     time.Time          `json:"endSched"`
	Cleaners     []AvailableCleaner `json:"cleaners"`
	CleanerCount int                `json:"cleanerCount"`
}
