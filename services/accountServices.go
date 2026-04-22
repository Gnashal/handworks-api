package services

import (
	"context"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *AccountService) withTx(
	ctx context.Context,
	fn func(pgx.Tx) error,
) (err error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				s.Logger.Error("rollback failed: %v", rbErr)
			}
		} else {
			err = tx.Commit(ctx)
		}
	}()
	return fn(tx)
}

// Customer methods
func (s *AccountService) SignUpCustomer(ctx context.Context, req types.SignUpCustomerRequest) (*types.SignUpCustomerResponse, error) {
	var customer types.Customer
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		acc, err := s.Tasks.CreateAccount(ctx, tx, req.FirstName, req.LastName, req.Email, req.Provider, req.ClerkID, req.Role)
		if err != nil {
			return err
		}
		customer.Account = *acc
		id, err := s.Tasks.CreateCustomer(ctx, tx, acc.ID)
		if err != nil {
			return err
		}
		customer.ID = id
		err = s.Tasks.UpdateCustomerMetadata(ctx, tx, customer.ID, customer.Account.ID, req.ClerkID)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	resp := &types.SignUpCustomerResponse{
		Customer: customer,
	}
	return resp, nil
}

func (s *AccountService) GetCustomer(ctx context.Context, id string) (*types.GetCustomerResponse, error) {
	var customer types.Customer
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		cust, err := s.Tasks.FetchCustomerData(ctx, tx, id)
		if err != nil {
			return err
		}
		acc, err := s.Tasks.FetchAccountData(ctx, tx, cust.Account.ID)
		if err != nil {
			return err
		}
		customer = *cust
		customer.Account = *acc
		return nil
	}); err != nil {
		return nil, err
	}
	resp := &types.GetCustomerResponse{
		Customer: customer,
	}
	return resp, nil
}

func (s *AccountService) GetCustomers(ctx context.Context, page, limit int) (*types.GetAllCustomersResponse, error) {
	var res *types.GetAllCustomersResponse
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		customers, err := s.Tasks.FetchAllCustomers(ctx, tx, page, limit)
		if err != nil {
			return err
		}
		res = customers
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}
func (s *AccountService) GetEmployees(ctx context.Context, page, limit int) (*types.GetAllEmployeesResponse, error) {
	var res *types.GetAllEmployeesResponse
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		employees, err := s.Tasks.FetchAllEmployees(ctx, tx, page, limit)
		if err != nil {
			return err
		}
		res = employees
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}
func (s *AccountService) UpdateCustomer(ctx context.Context, req types.UpdateCustomerRequest) (*types.UpdateCustomerResponse, error) {
	var customer types.Customer

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		cust, err := s.Tasks.UpdateCustomer(ctx, tx, req.ID, req.FirstName, req.LastName, req.Email)
		if err != nil {
			return err
		}
		customer = *cust

		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not update customer: %w", err)
	}
	return &types.UpdateCustomerResponse{
		Customer: customer,
	}, nil
}

func (s *AccountService) DeleteCustomer(ctx context.Context, id, accId string) (*types.DeleteCustomerResponse, error) {
	var customer types.Customer

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		cust, err := s.Tasks.DeleteCustomerData(ctx, tx, id, accId)
		if err != nil {
			return err
		}
		customer = *cust

		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not update customer: %w", err)
	}
	return &types.DeleteCustomerResponse{
		Ok:       true,
		Message:  "Success",
		Customer: customer,
	}, nil
}

// Employee methods
func (s *AccountService) SignUpEmployee(ctx context.Context, req types.SignUpEmployeeRequest) (*types.SignUpEmployeeResponse, error) {
	var employee types.Employee

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		acc, err := s.Tasks.CreateAccount(ctx, tx, req.FirstName, req.LastName, req.Email, req.Provider, req.ClerkID, req.Role)
		if err != nil {
			return err
		}
		parsedDate, err := time.Parse(time.RFC3339, req.HireDate)
		if err != nil {
			return fmt.Errorf("invalid hire date format: %w", err)
		}

		emp, err := s.Tasks.CreateEmployee(ctx, tx, acc.ID, req.Position, parsedDate)
		if err != nil {
			return err
		}
		employee = *emp
		err = s.Tasks.UpdateEmployeeMetadata(ctx, tx, employee.ID, acc.ID, req.ClerkID)
		if err != nil {
			return err
		}
		employee.Account = *acc
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to sign up employee: %w", err)
	}

	resp := &types.SignUpEmployeeResponse{
		Employee: employee,
	}
	return resp, nil
}

func (s *AccountService) GetEmployee(ctx context.Context, id string) (*types.GetEmployeeResponse, error) {
	var employee types.Employee

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		emp, err := s.Tasks.FetchEmployeeData(ctx, tx, id)
		if err != nil {
			return err
		}
		acc, err := s.Tasks.FetchAccountData(ctx, tx, emp.Account.ID)
		if err != nil {
			return err
		}
		employee = *emp
		employee.Account = *acc
		return nil
	}); err != nil {
		s.Logger.Error("failed to get employee: %v", err)
		return nil, fmt.Errorf("could not fetch employee: %w", err)
	}

	return &types.GetEmployeeResponse{
		Employee: employee,
	}, nil
}

func (s *AccountService) UpdateEmployee(ctx context.Context, req types.UpdateEmployeeRequest) (*types.UpdateEmployeeResponse, error) {
	var employee types.Employee

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		emp, err := s.Tasks.UpdateEmployee(ctx, tx, req.ID, req.EmployeeID, req.FirstName, req.LastName, req.Email)
		if err != nil {
			return err
		}
		employee = *emp
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not update employee: %w", err)
	}

	return &types.UpdateEmployeeResponse{
		Employee: employee,
	}, nil
}

func (s *AccountService) UpdateEmployeePerformanceScore(ctx context.Context, req types.UpdatePerformanceScoreRequest) (*types.UpdatePerformanceScoreResponse, error) {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		return s.Tasks.AddPerformanceScore(ctx, tx, req.NewPerformanceScore, req.ID)
	}); err != nil {
		return nil, fmt.Errorf("could not update employee performance score: %w", err)
	}

	return &types.UpdatePerformanceScoreResponse{
		Ok: true,
	}, nil
}

func (s *AccountService) UpdateEmployeeStatus(ctx context.Context, req types.UpdateEmployeeStatusRequest) (*types.UpdateEmployeeStatusResponse, error) {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		return s.Tasks.UpdateStatus(ctx, tx, req.Status, req.ID)
	}); err != nil {
		return nil, fmt.Errorf("could not update employee status: %w", err)
	}

	return &types.UpdateEmployeeStatusResponse{
		Ok: true,
	}, nil
}

func (s *AccountService) DeleteEmployee(ctx context.Context, id, accId string) (*types.DeleteEmployeeResponse, error) {
	var employee types.Employee

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		emp, err := s.Tasks.DeleteEmployeeData(ctx, tx, id, accId)
		if err != nil {
			return err
		}

		employee = *emp
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not delete employee: %w", err)
	}

	return &types.DeleteEmployeeResponse{
		Ok:       true,
		Message:  "Success",
		Employee: employee,
	}, nil
}
func (s *AccountService) SignUpAdmin(ctx context.Context, req types.SignUpAdminRequest) (*types.SignUpAdminResponse, error) {
	var admin types.Admin

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		acc, err := s.Tasks.CreateAccount(ctx, tx, req.FirstName, req.LastName, req.Email, req.Provider, req.ClerkID, req.Role)
		if err != nil {
			return err
		}
		admin.Account = *acc
		id, err := s.Tasks.CreateAdmin(ctx, tx, acc.ID)
		if err != nil {
			return err
		}
		admin.ID = id
		err = s.Tasks.UpdateAdminMetadata(ctx, tx, admin.ID, admin.Account.ID, req.ClerkID)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not create admin: %w", err)
	}
	return &types.SignUpAdminResponse{
		Admin: admin,
	}, nil
}

func (s *AccountService) EmployeeTimeIn(ctx context.Context, req types.TimeInRequest) (*types.EmployeeTimesheet, error) {
	var timesheet *types.EmployeeTimesheet

	if err := s.withTx(ctx, func(tx pgx.Tx) error {

		status := utils.DetermineAttendanceStatus(req.TimeIn)

		ts, err := s.Tasks.EmployeeTimeIn(ctx, tx, status, req)
		if err != nil {
			return err
		}

		if err := s.Tasks.UpdateStatus(ctx, tx, "ACTIVE", req.EmployeeId); err != nil {
			return err
		}

		timesheet = ts
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not time in employee: %w", err)
	}

	return timesheet, nil
}

func (s *AccountService) EmployeeTimeOut(ctx context.Context, req types.TimeOutRequest) (*types.EmployeeTimesheet, error) {
	var timesheet *types.EmployeeTimesheet

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		ts, err := s.Tasks.EmployeeTimeOut(ctx, tx, req)
		if err != nil {
			return err
		}

		if err := s.Tasks.UpdateStatus(ctx, tx, "INACTIVE", req.EmployeeId); err != nil {
			return err
		}

		timesheet = ts
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not time out employee: %w", err)
	}

	return timesheet, nil
}

func (s *AccountService) TimesheetToday(ctx context.Context, empId string) (*types.TimesheetToday, error) {
	currentDate := time.Now().Format("2006-01-02")
	var timesheet *types.EmployeeTimesheet

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		ts, err := s.Tasks.TimesheetToday(ctx, tx, empId, currentDate)
		if err != nil {
			return err
		}
		timesheet = ts
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not retrieve today's timesheet: %w", err)
	}

	return &types.TimesheetToday{
		Timesheet: timesheet,
	}, nil
}

func (s *AccountService) CreateAddress(ctx context.Context, req types.CreateAddressRequest) (*types.CreateAddressResponse, error) {
	var saved types.SavedAddress
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		address, err := s.Tasks.CreateAddress(ctx, tx, req.AccountID, req.Address)
		if err != nil {
			return err
		}
		saved = *address
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not create address: %w", err)
	}

	return &types.CreateAddressResponse{Address: saved}, nil
}

func (s *AccountService) GetAddress(ctx context.Context, id, accountID string) (*types.GetAddressResponse, error) {
	var saved types.SavedAddress
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		address, err := s.Tasks.FetchAddressByID(ctx, tx, id, accountID)
		if err != nil {
			return err
		}
		saved = *address
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not get address: %w", err)
	}

	return &types.GetAddressResponse{Address: saved}, nil
}

func (s *AccountService) GetAddresses(ctx context.Context, accountID string) (*types.GetAddressesResponse, error) {
	addresses := make([]types.SavedAddress, 0)
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		result, err := s.Tasks.FetchAddressesByAccountID(ctx, tx, accountID)
		if err != nil {
			return err
		}
		addresses = result
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not get addresses: %w", err)
	}

	return &types.GetAddressesResponse{Addresses: addresses}, nil
}

func (s *AccountService) UpdateAddress(ctx context.Context, req types.UpdateAddressRequest) (*types.UpdateAddressResponse, error) {
	var saved types.SavedAddress
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		address, err := s.Tasks.UpdateAddress(ctx, tx, req.ID, req.AccountID, req.Address)
		if err != nil {
			return err
		}
		saved = *address
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not update address: %w", err)
	}

	return &types.UpdateAddressResponse{Address: saved}, nil
}

func (s *AccountService) DeleteAddress(ctx context.Context, req types.DeleteAddressRequest) (*types.DeleteAddressResponse, error) {
	var saved types.SavedAddress
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		address, err := s.Tasks.DeleteAddress(ctx, tx, req.ID, req.AccountID)
		if err != nil {
			return err
		}
		saved = *address
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not delete address: %w", err)
	}

	return &types.DeleteAddressResponse{
		Ok:      true,
		Message: "Success",
		Address: saved,
	}, nil
}

func (s *AccountService) GetPhoneNumbers(ctx context.Context, accountID string) (*types.GetPhoneNumbersResponse, error) {
	phoneNumbers := make([]string, 0)
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		result, err := s.Tasks.FetchPhoneNumbers(ctx, tx, accountID)
		if err != nil {
			return err
		}
		phoneNumbers = result
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not get phone numbers: %w", err)
	}

	return &types.GetPhoneNumbersResponse{PhoneNumbers: phoneNumbers}, nil
}

func (s *AccountService) AddPhoneNumber(ctx context.Context, req types.AddPhoneNumberRequest) (*types.AddPhoneNumberResponse, error) {
	phoneNumber := strings.TrimSpace(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, fmt.Errorf("phone number cannot be empty")
	}

	phoneNumbers := make([]string, 0)
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		result, err := s.Tasks.AddPhoneNumber(ctx, tx, req.AccountID, phoneNumber)
		if err != nil {
			return err
		}
		phoneNumbers = result
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not add phone number: %w", err)
	}

	return &types.AddPhoneNumberResponse{PhoneNumbers: phoneNumbers}, nil
}

func (s *AccountService) DeletePhoneNumber(ctx context.Context, req types.DeletePhoneNumberRequest) (*types.DeletePhoneNumberResponse, error) {
	phoneNumber := strings.TrimSpace(req.PhoneNumber)
	if phoneNumber == "" {
		return nil, fmt.Errorf("phone number cannot be empty")
	}

	phoneNumbers := make([]string, 0)
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		result, err := s.Tasks.DeletePhoneNumber(ctx, tx, req.AccountID, phoneNumber)
		if err != nil {
			return err
		}
		phoneNumbers = result
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not delete phone number: %w", err)
	}

	return &types.DeletePhoneNumberResponse{PhoneNumbers: phoneNumbers}, nil
}
