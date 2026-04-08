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

type SavedAddress struct {
	ID        string    `json:"id"`
	AccountID string    `json:"accountId"`
	Address   Address   `json:"address"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateAddressRequest struct {
	AccountID string  `json:"accountId" binding:"required"`
	Address   Address `json:"address" binding:"required"`
}

type UpdateAddressRequest struct {
	ID        string  `json:"id" binding:"required"`
	AccountID string  `json:"accountId" binding:"required"`
	Address   Address `json:"address" binding:"required"`
}

type DeleteAddressRequest struct {
	ID        string `json:"id" binding:"required"`
	AccountID string `json:"accountId" binding:"required"`
}

type CreateAddressResponse struct {
	Address SavedAddress `json:"address"`
}

type GetAddressResponse struct {
	Address SavedAddress `json:"address"`
}

type GetAddressesResponse struct {
	Addresses []SavedAddress `json:"addresses"`
}

type UpdateAddressResponse struct {
	Address SavedAddress `json:"address"`
}

type DeleteAddressResponse struct {
	Ok      bool         `json:"ok"`
	Message string       `json:"message"`
	Address SavedAddress `json:"address"`
}

type AddPhoneNumberRequest struct {
	AccountID   string `json:"accountId" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type DeletePhoneNumberRequest struct {
	AccountID   string `json:"accountId" binding:"required"`
	PhoneNumber string `json:"phoneNumber" binding:"required"`
}

type GetPhoneNumbersResponse struct {
	PhoneNumbers []string `json:"phoneNumbers"`
}

type AddPhoneNumberResponse struct {
	PhoneNumbers []string `json:"phoneNumbers"`
}

type DeletePhoneNumberResponse struct {
	PhoneNumbers []string `json:"phoneNumbers"`
}

type EmployeeTimesheet struct {
	TimesheetId string     `json:"id" db:"id"`
	EmployeeId  string     `json:"employee_id," db:"employee_id"`
	WorkDate    time.Time  `json:"work_date" db:"work_date"`
	TimeIn      *time.Time `json:"time_in" db:"time_in"`
	TimeOut     *time.Time `json:"time_out" db:"time_out"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type TimesheetToday struct {
	Timesheet *EmployeeTimesheet `json:"timesheet,omitempty"`
}

type TimeInRequest struct {
	EmployeeId string    `json:"employee_id"`
	TimeIn     time.Time `json:"time_in"`
}

type TimeOutRequest struct {
	EmployeeId string    `json:"employee_id"`
	TimeOut    time.Time `json:"time_out"`
}
