package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"time"

	"github.com/bytedance/gopkg/util/logger"
	"github.com/clerk/clerk-sdk-go/v2/user"
	"github.com/jackc/pgx/v5"
)

type AccountTasks struct{}

func (t *AccountTasks) CreateAccount(c context.Context, tx pgx.Tx, FirstName, LastName, Email, Provider, ClerkId, Role string) (*types.Account, error) {
	var acc types.Account
	err := tx.QueryRow(c,
		`INSERT INTO account.accounts (first_name, last_name, email, provider, clerk_id, role)
		VALUES ($1,$2, $3, $4, $5, $6)
		RETURNING first_name, last_name, email, provider, clerk_id, role, id, created_at, updated_at`,
		FirstName, LastName, Email, Provider, ClerkId, Role,
	).Scan(
		&acc.FirstName,
		&acc.LastName,
		&acc.Email,
		&acc.Provider,
		&acc.ClerkID,
		&acc.Role,
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create account: %w", err)
	}
	return &acc, nil
}

func (t *AccountTasks) CreateCustomer(c context.Context, tx pgx.Tx, id string) (string, error) {
	var customerId string
	if err := tx.QueryRow(c,
		`INSERT INTO account.customers (account_id)
		VALUES ($1)
		RETURNING id`, id).Scan(&customerId); err != nil {
		return "", fmt.Errorf("could not insert into customer table: %w", err)
	}
	return customerId, nil
}

func (t *AccountTasks) CreateEmployee(c context.Context, tx pgx.Tx, id, position string, hireDate time.Time) (*types.Employee, error) {
	var emp types.Employee
	if err := tx.QueryRow(c,
		`INSERT INTO account.employees (account_id, position, status, performance_score, hire_date, num_ratings)
	VALUES ($1, $2, $3, $4, $5, $6) 
	RETURNING id, position, status, performance_score, hire_date, num_ratings`, id, position, "INACTIVE", 5.0, hireDate, 0).Scan(
		&emp.ID, &emp.Position, &emp.Status, &emp.PerformanceScore,
		&emp.HireDate, &emp.NumRatings); err != nil {
		return nil, fmt.Errorf("could not insert into employee table: %w", err)
	}
	return &emp, nil
}
func (t *AccountTasks) CreateAdmin(c context.Context, tx pgx.Tx, id string) (string, error) {
	var adminId string
	if err := tx.QueryRow(c,
		`INSERT INTO account.admins (account_id)
		VALUES ($1)
		RETURNING id`, id).Scan(&adminId); err != nil {
		return "", fmt.Errorf("could not insert into admin table: %w", err)
	}
	return adminId, nil
}

func (t *AccountTasks) FetchAccountData(c context.Context, tx pgx.Tx, ID string) (*types.Account, error) {
	var acc types.Account
	if err := tx.QueryRow(c,
		`SELECT first_name, last_name, email, provider, clerk_id, role, id, created_at, updated_at
		 FROM account.accounts WHERE id = $1`,
		ID,
	).Scan(
		&acc.FirstName,
		&acc.LastName,
		&acc.Email,
		&acc.Provider,
		&acc.ClerkID,
		&acc.Role,
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("could not query accounts table: %w", err)
	}
	return &acc, nil
}
func (t *AccountTasks) FetchCustomerData(c context.Context, tx pgx.Tx, ID string) (*types.Customer, error) {
	var customer types.Customer

	if err := tx.QueryRow(c,
		`SELECT id, account_id FROM account.customers WHERE id = $1`,
		ID,
	).Scan(&customer.ID, &customer.Account.ID); err != nil {
		return nil, fmt.Errorf("could not query customer table: %w", err)
	}
	return &customer, nil
}
func (t *AccountTasks) FetchAllCustomers(c context.Context, tx pgx.Tx, page, limit int) (*types.GetAllCustomersResponse, error) {
	var rawJSON []byte
	err := tx.QueryRow(c,
		`SELECT account.get_all_customers($1, $2)`,
		page, limit,
	).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_all_customers: %w", err)
	}

	var response types.GetAllCustomersResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal quotes JSON: %v", err)
		return nil, fmt.Errorf("unmarshal quotes: %w", err)
	}

	return &response, nil
}
func (t *AccountTasks) FetchAllEmployees(c context.Context, tx pgx.Tx, page, limit int) (*types.GetAllEmployeesResponse, error) {
	var rawJSON []byte
	err := tx.QueryRow(c,
		`SELECT account.get_all_employees($1, $2)`,
		page, limit,
	).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_all_employees: %w", err)
	}

	var response types.GetAllEmployeesResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal quotes JSON: %v", err)
		return nil, fmt.Errorf("unmarshal quotes: %w", err)
	}

	return &response, nil
}
func (t *AccountTasks) FetchEmployeeData(c context.Context, tx pgx.Tx, ID string) (*types.Employee, error) {
	var emp types.Employee

	if err := tx.QueryRow(c,
		`SELECT id, account_id, position, status, performance_score, hire_date, num_ratings
		 FROM account.employees WHERE id = $1`,
		ID,
	).Scan(
		&emp.ID,
		&emp.Account.ID,
		&emp.Position,
		&emp.Status,
		&emp.PerformanceScore,
		&emp.HireDate,
		&emp.NumRatings,
	); err != nil {
		return nil, fmt.Errorf("could not query employees table: %w", err)
	}
	return &emp, nil
}
func (t *AccountTasks) UpdateCustomer(c context.Context, tx pgx.Tx, id, firstName, lastName, email string) (*types.Customer, error) {
	acc, err := t.UpdateAccount(c, tx, id, firstName, lastName, email)
	if err != nil {
		return nil, fmt.Errorf("could not update account: %w", err)
	}
	customer, err := t.FetchCustomerData(c, tx, id)
	if err != nil {
		return nil, fmt.Errorf("could not fetch customer data: %w", err)
	}

	customer.Account = *acc
	return customer, nil
}
func (t *AccountTasks) UpdateEmployee(c context.Context, tx pgx.Tx, id, empId, firstName, lastName, email string) (*types.Employee, error) {
	acc, err := t.UpdateAccount(c, tx, id, firstName, lastName, email)
	if err != nil {
		return nil, fmt.Errorf("could not update account: %w", err)
	}
	employee, err := t.FetchEmployeeData(c, tx, empId)
	if err != nil {
		return nil, fmt.Errorf("could not fetch employee data: %w", err)
	}

	employee.Account = *acc
	return employee, nil
}
func (t *AccountTasks) UpdateAccount(c context.Context, tx pgx.Tx, id, firstName, lastName, email string) (*types.Account, error) {
	var acc types.Account
	err := tx.QueryRow(c,
		`UPDATE account.accounts
		 SET first_name = $1, last_name = $2, email = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, first_name, last_name, email, role, provider, created_at, updated_at, clerk_id`,
		firstName, lastName, email, id,
	).Scan(
		&acc.ID,
		&acc.FirstName,
		&acc.LastName,
		&acc.Email,
		&acc.Role,
		&acc.Provider,
		&acc.CreatedAt,
		&acc.UpdatedAt,
		&acc.ClerkID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not update account: %w", err)
	}
	return &acc, nil
}
func (t *AccountTasks) DeleteCustomerData(c context.Context, tx pgx.Tx, customerId, accId string) (*types.Customer, error) {
	var cust types.Customer
	var acc types.Account

	if err := tx.QueryRow(c, `
		DELETE FROM account.customers
		WHERE id = $1
		RETURNING id, account_id
	`, customerId).Scan(
		&cust.ID,
		&cust.Account.ID,
	); err != nil {
		return nil, fmt.Errorf("could not delete customer with id %s: %w", customerId, err)
	}

	if err := tx.QueryRow(c, `
		DELETE FROM account.accounts
		WHERE id = $1
		RETURNING first_name, last_name, email, provider, clerk_id, role, id, created_at, updated_at
	`, accId).Scan(
		&acc.FirstName,
		&acc.LastName,
		&acc.Email,
		&acc.Provider,
		&acc.ClerkID,
		&acc.Role,
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("could not delete account with id %s: %w", accId, err)
	}
	cust.Account = acc
	return &cust, nil
}
func (t *AccountTasks) DeleteEmployeeData(c context.Context, tx pgx.Tx, empId, accId string) (*types.Employee, error) {
	var emp types.Employee
	var acc types.Account
	// Delete employee, return row
	if err := tx.QueryRow(c,
		`DELETE FROM account.employees
		 WHERE id = $1
		 RETURNING id, account_id, position, status, performance_score, hire_date, num_ratings`,
		empId,
	).Scan(
		&emp.ID,
		&emp.Account.ID,
		&emp.Position,
		&emp.Status,
		&emp.PerformanceScore,
		&emp.HireDate,
		&emp.NumRatings,
	); err != nil {
		return nil, fmt.Errorf("could not delete employee: %w", err)
	}
	if err := tx.QueryRow(c,
		`DELETE FROM account.accounts
		 WHERE id = $1
		 RETURNING first_name, last_name, email, provider, clerk_id, role, id, created_at, updated_at`,
		accId,
	).Scan(
		&acc.FirstName,
		&acc.LastName,
		&acc.Email,
		&acc.Provider,
		&acc.ClerkID,
		&acc.Role,
		&acc.ID,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("could not delete account: %w", err)
	}
	emp.Account = acc
	return &emp, nil
}

func (t *AccountTasks) CreateAddress(c context.Context, tx pgx.Tx, accountID string, address types.Address) (*types.SavedAddress, error) {
	var saved types.SavedAddress
	err := tx.QueryRow(c,
		`INSERT INTO account.addresses (account_id, address_human, address_lat, address_lng)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, account_id, address_human, address_lat, address_lng, created_at, updated_at`,
		accountID,
		address.AddressHuman,
		address.AddressLat,
		address.AddressLng,
	).Scan(
		&saved.ID,
		&saved.AccountID,
		&saved.Address.AddressHuman,
		&saved.Address.AddressLat,
		&saved.Address.AddressLng,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create address: %w", err)
	}
	return &saved, nil
}

func (t *AccountTasks) FetchAddressByID(c context.Context, tx pgx.Tx, id, accountID string) (*types.SavedAddress, error) {
	var saved types.SavedAddress
	err := tx.QueryRow(c,
		`SELECT id, account_id, address_human, address_lat, address_lng, created_at, updated_at
		 FROM account.addresses
		 WHERE id = $1 AND account_id = $2`,
		id,
		accountID,
	).Scan(
		&saved.ID,
		&saved.AccountID,
		&saved.Address.AddressHuman,
		&saved.Address.AddressLat,
		&saved.Address.AddressLng,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not query address with id %s: %w", id, err)
	}
	return &saved, nil
}

func (t *AccountTasks) FetchAddressesByAccountID(c context.Context, tx pgx.Tx, accountID string) ([]types.SavedAddress, error) {
	rows, err := tx.Query(c,
		`SELECT id, account_id, address_human, address_lat, address_lng, created_at, updated_at
		 FROM account.addresses
		 WHERE account_id = $1
		 ORDER BY created_at DESC`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not query addresses by account id %s: %w", accountID, err)
	}
	defer rows.Close()

	addresses := make([]types.SavedAddress, 0)
	for rows.Next() {
		var saved types.SavedAddress
		if err := rows.Scan(
			&saved.ID,
			&saved.AccountID,
			&saved.Address.AddressHuman,
			&saved.Address.AddressLat,
			&saved.Address.AddressLng,
			&saved.CreatedAt,
			&saved.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("could not scan addresses: %w", err)
		}
		addresses = append(addresses, saved)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate addresses: %w", err)
	}

	return addresses, nil
}

func (t *AccountTasks) UpdateAddress(c context.Context, tx pgx.Tx, id, accountID string, address types.Address) (*types.SavedAddress, error) {
	var saved types.SavedAddress
	err := tx.QueryRow(c,
		`UPDATE account.addresses
		 SET address_human = $1,
		     address_lat = $2,
		     address_lng = $3,
		     updated_at = NOW()
		 WHERE id = $4 AND account_id = $5
		 RETURNING id, account_id, address_human, address_lat, address_lng, created_at, updated_at`,
		address.AddressHuman,
		address.AddressLat,
		address.AddressLng,
		id,
		accountID,
	).Scan(
		&saved.ID,
		&saved.AccountID,
		&saved.Address.AddressHuman,
		&saved.Address.AddressLat,
		&saved.Address.AddressLng,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not update address with id %s: %w", id, err)
	}
	return &saved, nil
}

func (t *AccountTasks) DeleteAddress(c context.Context, tx pgx.Tx, id, accountID string) (*types.SavedAddress, error) {
	var saved types.SavedAddress
	err := tx.QueryRow(c,
		`DELETE FROM account.addresses
		 WHERE id = $1 AND account_id = $2
		 RETURNING id, account_id, address_human, address_lat, address_lng, created_at, updated_at`,
		id,
		accountID,
	).Scan(
		&saved.ID,
		&saved.AccountID,
		&saved.Address.AddressHuman,
		&saved.Address.AddressLat,
		&saved.Address.AddressLng,
		&saved.CreatedAt,
		&saved.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("could not delete address with id %s: %w", id, err)
	}
	return &saved, nil
}

func (t *AccountTasks) FetchPhoneNumbers(c context.Context, tx pgx.Tx, accountID string) ([]string, error) {
	phoneNumbers := make([]string, 0)
	err := tx.QueryRow(c,
		`SELECT COALESCE(phone_numbers, '{}'::text[])
		 FROM account.accounts
		 WHERE id = $1`,
		accountID,
	).Scan(&phoneNumbers)
	if err != nil {
		return nil, fmt.Errorf("could not fetch phone numbers for account id %s: %w", accountID, err)
	}
	return phoneNumbers, nil
}

func (t *AccountTasks) AddPhoneNumber(c context.Context, tx pgx.Tx, accountID, phoneNumber string) ([]string, error) {
	phoneNumbers := make([]string, 0)
	err := tx.QueryRow(c,
		`UPDATE account.accounts
		 SET phone_numbers = CASE
		 	WHEN $2 = ANY(phone_numbers) THEN phone_numbers
		 	ELSE array_append(phone_numbers, $2)
		 END,
		 updated_at = NOW()
		 WHERE id = $1
		 RETURNING COALESCE(phone_numbers, '{}'::text[])`,
		accountID,
		phoneNumber,
	).Scan(&phoneNumbers)
	if err != nil {
		return nil, fmt.Errorf("could not add phone number for account id %s: %w", accountID, err)
	}
	return phoneNumbers, nil
}

func (t *AccountTasks) DeletePhoneNumber(c context.Context, tx pgx.Tx, accountID, phoneNumber string) ([]string, error) {
	phoneNumbers := make([]string, 0)
	err := tx.QueryRow(c,
		`UPDATE account.accounts
		 SET phone_numbers = array_remove(phone_numbers, $2),
		 updated_at = NOW()
		 WHERE id = $1
		 RETURNING COALESCE(phone_numbers, '{}'::text[])`,
		accountID,
		phoneNumber,
	).Scan(&phoneNumbers)
	if err != nil {
		return nil, fmt.Errorf("could not delete phone number for account id %s: %w", accountID, err)
	}
	return phoneNumbers, nil
}

func (t *AccountTasks) AddPerformanceScore(c context.Context, tx pgx.Tx, newScore float32, empId string) error {

	args := pgx.NamedArgs{
		"newScore": newScore,
		"id":       empId,
	}
	cmdTag, err := tx.Exec(c,
		`UPDATE account.employees
	 SET performance_score = ((performance_score * num_ratings) + @newScore) / (num_ratings + 1),
	     num_ratings = num_ratings + 1, updated_at = NOW()
	 WHERE id = @id::uuid`,
		args,
	)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no employee found with id %s", empId)
	}
	return nil
}
func (a *AccountTasks) UpdateStatus(c context.Context, tx pgx.Tx, status, empId string) error {
	args := pgx.NamedArgs{
		"newStatus": status,
		"id":        empId,
	}
	_, err := tx.Exec(c,
		`UPDATE account.employees
	SET status = @newStatus,
	    updated_at = NOW()
	WHERE id = @id`, args)
	if err != nil {
		return err
	}
	return nil
}
func (a *AccountTasks) UpdateCustomerMetadata(c context.Context, tx pgx.Tx, customerId, accId, clerkId string) error {
	metadata := map[string]string{"custId": customerId, "accId": accId}
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	raw := json.RawMessage(jsonData)

	_, err = user.UpdateMetadata(c, clerkId, &user.UpdateMetadataParams{
		PublicMetadata: &raw,
	})
	if err != nil {
		return fmt.Errorf("failed to update clerk metadata: %w", err)
	}
	return nil
}
func (a *AccountTasks) UpdateEmployeeMetadata(c context.Context, tx pgx.Tx, employeeId, accId, clerkId string) error {
	metadata := map[string]string{"empId": employeeId, "accId": accId}
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	raw := json.RawMessage(jsonData)

	_, err = user.UpdateMetadata(c, clerkId, &user.UpdateMetadataParams{
		PublicMetadata: &raw,
	})
	if err != nil {
		return fmt.Errorf("failed to update clerk metadata: %w", err)
	}
	return nil
}
func (a *AccountTasks) UpdateAdminMetadata(c context.Context, tx pgx.Tx, adminId, accId, clerkId string) error {
	metadata := map[string]string{"adminId": adminId, "accId": accId}
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	raw := json.RawMessage(jsonData)

	_, err = user.UpdateMetadata(c, clerkId, &user.UpdateMetadataParams{
		PublicMetadata: &raw,
	})
	if err != nil {
		return fmt.Errorf("failed to update clerk metadata: %w", err)
	}
	return nil
}

func (a *AccountTasks) EmployeeTimeIn(c context.Context, tx pgx.Tx, status string, req types.TimeInRequest) (*types.EmployeeTimesheet, error) {
	timesheet := &types.EmployeeTimesheet{}

	query := `
	INSERT INTO account.employee_timesheet 
	(employee_id, work_date, time_in, status, created_at, updated_at)
	VALUES ($1, CURRENT_DATE, $2, $3, NOW(), NOW())
	RETURNING id, employee_id, work_date, time_in, time_out, status, created_at, updated_at
	`

	err := tx.QueryRow(c, query, req.EmployeeId, req.TimeIn, status).Scan(
		&timesheet.TimesheetId,
		&timesheet.EmployeeId,
		&timesheet.WorkDate,
		&timesheet.TimeIn,
		&timesheet.TimeOut,
		&timesheet.Status,
		&timesheet.CreatedAt,
		&timesheet.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("could not insert employee_timesheet: %w", err)
	}

	return timesheet, nil
}

func (a *AccountTasks) TimesheetToday(c context.Context, tx pgx.Tx, empId, currentDate string) (*types.EmployeeTimesheet, error) {
	timesheet := &types.EmployeeTimesheet{}

	query := `
	SELECT id, employee_id, work_date, time_in, time_out, status, created_at, updated_at
	FROM account.employee_timesheet
	WHERE employee_id = $1
	AND work_date = $2
	`

	err := tx.QueryRow(c, query, empId, currentDate).Scan(
		&timesheet.TimesheetId,
		&timesheet.EmployeeId,
		&timesheet.WorkDate,
		&timesheet.TimeIn,
		&timesheet.TimeOut,
		&timesheet.Status,
		&timesheet.CreatedAt,
		&timesheet.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No timesheet for today
		}
		return nil, fmt.Errorf("could not query employee_timesheet: %w", err)
	}

	return timesheet, nil
}

func (a *AccountTasks) EmployeeTimeOut(c context.Context, tx pgx.Tx, req types.TimeOutRequest) (*types.EmployeeTimesheet, error) {

	timesheet := &types.EmployeeTimesheet{}

	args := pgx.NamedArgs{
		"employee_id": req.EmployeeId,
		"time_out":    req.TimeOut,
	}

	query := `
	UPDATE account.employee_timesheet
	SET time_out = @time_out,
	    updated_at = NOW()
	WHERE employee_id = @employee_id
	AND work_date = CURRENT_DATE
	AND time_in IS NOT NULL
	AND time_out IS NULL
	RETURNING id, employee_id, work_date, time_in, time_out, status, created_at, updated_at
	`

	err := tx.QueryRow(c, query, args).Scan(
		&timesheet.TimesheetId,
		&timesheet.EmployeeId,
		&timesheet.WorkDate,
		&timesheet.TimeIn,
		&timesheet.TimeOut,
		&timesheet.Status,
		&timesheet.CreatedAt,
		&timesheet.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("could not update employee_timesheet: %w", err)
	}

	return timesheet, nil
}
