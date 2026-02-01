package types
type AdminDashboardRequest struct {
	AdminID    string `json:"adminId"     binding:"required"`
	DateFilter string `json:"dateFilter"  binding:"required"`
}

type ClientsGrowthIndex struct {
	New int32 `json:"new"`
	Returning int32 `json:"returning"`
	Inactive int32 `json:"inactive"`
	GrowthIndex float64 `json:"growthIndex"`
}

type AdminDashboardResponse struct {
	Sales float64 `json:"sales"`
	SalesGrowthIndex float64 `json:"salesGrowthIndex"`
	Bookings int32 `json:"bookings"`
	BookingsGrowthIndex float64 `json:"bookingsGrowthIndex"`
	ActiveSessions int32 `json:"activeSessions"`
	ActiveSessionsGrowthIndex float64 `json:"activeSessionsGrowthIndex"`
	Clients int32 `json:"clients"`
	ClientGrowthIndex ClientsGrowthIndex `json:"clientsGrothIndex"`
}