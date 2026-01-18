package services

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"time"

	"github.com/clerk/clerk-sdk-go/v2/user"
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
			return  err
		}
		customer.ID = id
		return nil
	}); err != nil {
		return nil, err
	}
	// metadata to store in clerk
	metadata := map[string]string{"custId": customer.ID}
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	raw := json.RawMessage(jsonData)

	_, err = user.UpdateMetadata(ctx, req.ClerkID, &user.UpdateMetadataParams{
		PublicMetadata: &raw,
	})
	if err != nil {
		s.Logger.Error("Failed to update Clerk Metadata: %v", err)
		return nil, fmt.Errorf("failed to update clerk metadata: %w", err)
	}
	s.Logger.Info("Updated Clerk Metadata")
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

func (s *AccountService) DeleteCustomer(ctx context.Context, id, accId string) (*types.DeleteCustomerResponse,error) {
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
		Ok: true,
		Message: "Success",
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
		parsedDate, err := time.Parse("2006-01-02", req.HireDate)
		if err != nil {
			return fmt.Errorf("invalid hire date format: %w", err)
		}

		emp, err := s.Tasks.CreateEmployee(ctx, tx, acc.ID, req.Position, parsedDate)
		if err != nil {
			return err
		}

		employee = *emp
		employee.Account = *acc
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to sign up employee: %w", err)
	}
	// metadata to store in clerk
	metadata := map[string]string{"empId": employee.ID}
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	raw := json.RawMessage(jsonData)

	_, err = user.UpdateMetadata(ctx, req.ClerkID, &user.UpdateMetadataParams{
		PublicMetadata: &raw,
	})
	if err != nil {
		s.Logger.Error("Failed to update Clerk Metadata: %v", err)
		return nil, fmt.Errorf("failed to update clerk metadata: %w", err)
	}
	s.Logger.Info("Updated Clerk Metadata")
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
		return nil, fmt.Errorf("could not fetch employee: %w", err)
	}

	return &types.GetEmployeeResponse{
		Employee: employee,
	}, nil
}

func (s *AccountService) UpdateEmployee(ctx context.Context, req types.UpdateEmployeeRequest) (*types.UpdateEmployeeResponse, error) {
	var employee types.Employee

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		emp, err := s.Tasks.UpdateEmployee(ctx, tx, req.ID, req.FirstName, req.LastName, req.Email)
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
func (s* AccountService) SignUpAdmin(ctx context.Context, req types.SignUpAdminRequest) (*types.SignUpAdminResponse, error) {
	var admin types.Admin
	
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		acc, err := s.Tasks.CreateAccount(ctx, tx, req.FirstName, req.LastName, req.Email, req.Provider, req.ClerkID, req.Role)
		if err != nil {
			return err
		}
		admin.Account = *acc
		id, err := s.Tasks.CreateAdmin(ctx, tx, acc.ID)
		if err != nil {
			return  err
		}
		admin.ID = id
		return nil
	}); err != nil {
		return nil, fmt.Errorf("could not create admin: %w", err)
	}
	// metadata to store in clerk
	metadata := map[string]string{"adminId": admin.ID}
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	raw := json.RawMessage(jsonData)

	_, err = user.UpdateMetadata(ctx, req.ClerkID, &user.UpdateMetadataParams{
		PublicMetadata: &raw,
	})
	if err != nil {
		s.Logger.Error("Failed to update Clerk Metadata: %v", err)
		return nil, fmt.Errorf("failed to update clerk metadata: %w", err)
	}
	s.Logger.Info("Updated Clerk Metadata")
	return &types.SignUpAdminResponse{
		Admin: admin,
	}, nil
}
