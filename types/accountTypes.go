package types

import "time"

type Account struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	Role      string    `json:"role"`
	ClerkID   string    `json:"clerk_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Customer struct {
	ID      string  `json:"id"`
	Account Account `json:"account"`
}

type Employee struct {
	ID               string    `json:"id"`
	Account          Account   `json:"account"`
	Position         string    `json:"position"`
	Status           string    `json:"status"` // ACTIVE / ONDUTY / INACTIVE
	PerformanceScore float32   `json:"performance_score"`
	HireDate         time.Time `json:"hire_date"`
	NumRatings       int32     `json:"num_ratings"`
}
type Admin struct {
	ID      string  `json:"id"`
	Account Account `json:"account"`
}
type SignUpAdminRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Email     string `json:"email"      binding:"required,email"`
	Provider  string `json:"provider"   binding:"required"`
	ClerkID   string `json:"clerk_id"   binding:"required"`
	Role      string `json:"role"       binding:"required"`
}

type SignUpCustomerRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Email     string `json:"email"      binding:"required,email"`
	Provider  string `json:"provider"   binding:"required"`
	ClerkID   string `json:"clerk_id"   binding:"required"`
	Role      string `json:"role"       binding:"required"`
}

type SignUpEmployeeRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name"  binding:"required"`
	Email     string `json:"email"      binding:"required,email"`
	Provider  string `json:"provider"   binding:"required"`
	ClerkID   string `json:"clerk_id"   binding:"required"`
	Role      string `json:"role"       binding:"required"`
	Position  string `json:"position"   binding:"required"`
	HireDate  string `json:"hire_date"  binding:"required"`
}

type UpdateCustomerRequest struct {
	ID         string `json:"id"           binding:"required"`
	FirstName  string `json:"first_name"   binding:"omitempty"`
	LastName   string `json:"last_name"    binding:"omitempty"`
	Email      string `json:"email"        binding:"omitempty,email"`
	CustomerID string `json:"customer_id"`
}

type UpdateEmployeeRequest struct {
	ID         string `json:"id"           binding:"required"`
	FirstName  string `json:"first_name"   binding:"omitempty"`
	LastName   string `json:"last_name"    binding:"omitempty"`
	Email      string `json:"email"        binding:"omitempty,email"`
	EmployeeID string `json:"employee_id"`
}

type UpdatePerformanceScoreRequest struct {
	ID                  string  `form:"id"    binding:"required"`
	NewPerformanceScore float32 `form:"score" binding:"required"`
}

type UpdateEmployeeStatusRequest struct {
	ID     string `form:"id"     binding:"required"`
	Status string `form:"status" binding:"required"`
}

type DeleteEmployeeRequest struct {
	ID    string `json:"id"     binding:"required"`
	EmpID string `json:"empId"  binding:"required"`
}

type DeleteCustomerRequest struct {
	ID     string `json:"id"      binding:"required"`
	CustID string `json:"custId"  binding:"required"`
}

type SignUpCustomerResponse struct {
	Customer Customer `json:"customer"`
}

type SignUpEmployeeResponse struct {
	Employee Employee `json:"employee"`
}
type SignUpAdminResponse struct {
	Admin Admin `json:"admin"`
}

// READ
type GetCustomerResponse struct {
	Customer Customer `json:"customer"`
}

type GetEmployeeResponse struct {
	Employee Employee `json:"employee"`
}

type GetAllCustomersResponse struct {
	Customers []Customer `json:"customers"`
}
type GetAllEmployeesResponse struct {
	Employees []Employee `json:"employees"`
}

// UPDATE
type UpdateCustomerResponse struct {
	Customer Customer `json:"customer"`
}

type UpdateEmployeeResponse struct {
	Employee Employee `json:"employee"`
}

type UpdatePerformanceScoreResponse struct {
	Ok bool `json:"ok"`
}

type UpdateEmployeeStatusResponse struct {
	Ok bool `json:"ok"`
}

// DELETE
type DeleteEmployeeResponse struct {
	Ok       bool     `json:"ok"`
	Message  string   `json:"message"`
	Employee Employee `json:"employee"`
}

type DeleteCustomerResponse struct {
	Ok       bool     `json:"ok"`
	Message  string   `json:"message"`
	Customer Customer `json:"customer"`
}
