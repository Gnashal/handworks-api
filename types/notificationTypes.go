package types

type SubscribeNotificationRequest struct {
	Token          string `json:"token" binding:"required"`
	Role           string `json:"role" binding:"required"`
	EmployeeID     string `json:"employeeId"`
	AdminID        string `json:"adminId"`
	CustomerID     string `json:"customerId"`
	InstallationID string `json:"installationId"`
	Platform       string `json:"platform"`
}

type SubscribeNotificationResponse struct {
	Ok     bool     `json:"ok"`
	Topics []string `json:"topics"`
}

type UnsubscribeNotificationRequest struct {
	Token      string `json:"token" binding:"required"`
	Role       string `json:"role" binding:"required"`
	EmployeeID string `json:"employeeId"`
	AdminID    string `json:"adminId"`
	CustomerID string `json:"customerId"`
}

type UnsubscribeNotificationResponse struct {
	Ok bool `json:"ok"`
}
