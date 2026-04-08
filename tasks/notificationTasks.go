package tasks

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type NotificationTasks struct{}

func (t *NotificationTasks) UpsertEmployeeFCMToken(
	ctx context.Context,
	tx pgx.Tx,
	employeeID string,
	installationID string,
	fcmToken string,
	platform string,
) error {
	res, err := tx.Exec(ctx, `
		INSERT INTO account.fcm_tokens
			(account_id, role, installation_id, fcm_token, platform, is_active, last_seen_at, created_at, updated_at)
		SELECT
			e.account_id,
			'employee',
			$2,
			$3,
			$4,
			TRUE,
			NOW(),
			NOW(),
			NOW()
		FROM account.employees e
		WHERE e.id = $1
		ON CONFLICT (account_id, installation_id)
		DO UPDATE
		SET
			fcm_token = EXCLUDED.fcm_token,
			role = EXCLUDED.role,
			platform = EXCLUDED.platform,
			is_active = TRUE,
			last_seen_at = NOW(),
			updated_at = NOW();
	`, employeeID, installationID, fcmToken, platform)
	if err != nil {
		return fmt.Errorf("failed to upsert employee fcm token: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("employee not found")
	}

	return nil
}

func (t *NotificationTasks) DeactivateToken(
	ctx context.Context,
	tx pgx.Tx,
	fcmToken string,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE account.fcm_tokens
		SET is_active = FALSE, updated_at = NOW()
		WHERE fcm_token = $1;
	`, fcmToken)
	if err != nil {
		return fmt.Errorf("failed to deactivate fcm token: %w", err)
	}

	return nil
}

func (t *NotificationTasks) UpsertAdminFCMToken(
	ctx context.Context,
	tx pgx.Tx,
	adminID string,
	installationID string,
	fcmToken string,
	platform string,
) error {
	res, err := tx.Exec(ctx, `
		INSERT INTO account.fcm_tokens
			(account_id, role, installation_id, fcm_token, platform, is_active, last_seen_at, created_at, updated_at)
		SELECT
			a.account_id,
			'admin',
			$2,
			$3,
			$4,
			TRUE,
			NOW(),
			NOW(),
			NOW()
		FROM account.admins a
		WHERE a.id = $1
		ON CONFLICT (account_id, installation_id)
		DO UPDATE
		SET
			fcm_token = EXCLUDED.fcm_token,
			role = EXCLUDED.role,
			platform = EXCLUDED.platform,
			is_active = TRUE,
			last_seen_at = NOW(),
			updated_at = NOW();
	`, adminID, installationID, fcmToken, platform)
	if err != nil {
		return fmt.Errorf("failed to upsert admin fcm token: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("admin not found")
	}

	return nil
}

func (t *NotificationTasks) UpsertCustomerFCMToken(
	ctx context.Context,
	tx pgx.Tx,
	customerID string,
	installationID string,
	fcmToken string,
	platform string,
) error {
	res, err := tx.Exec(ctx, `
		INSERT INTO account.fcm_tokens
			(account_id, role, installation_id, fcm_token, platform, is_active, last_seen_at, created_at, updated_at)
		SELECT
			c.account_id,
			'customer',
			$2,
			$3,
			$4,
			TRUE,
			NOW(),
			NOW(),
			NOW()
		FROM account.customers c
		WHERE c.id = $1
		ON CONFLICT (account_id, installation_id)
		DO UPDATE
		SET
			fcm_token = EXCLUDED.fcm_token,
			role = EXCLUDED.role,
			platform = EXCLUDED.platform,
			is_active = TRUE,
			last_seen_at = NOW(),
			updated_at = NOW();
	`, customerID, installationID, fcmToken, platform)
	if err != nil {
		return fmt.Errorf("failed to upsert customer fcm token: %w", err)
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

func (t *NotificationTasks) GetActiveEmployeeTokens(
	ctx context.Context,
	tx pgx.Tx,
	employeeID string,
) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT t.fcm_token
		FROM account.fcm_tokens t
		JOIN account.employees e
		  ON e.account_id = t.account_id
		WHERE e.id = $1
		  AND t.role = 'employee'
		  AND t.is_active = TRUE;
	`, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active employee tokens: %w", err)
	}
	defer rows.Close()

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if scanErr := rows.Scan(&token); scanErr != nil {
			return nil, fmt.Errorf("failed to scan employee token: %w", scanErr)
		}
		tokens = append(tokens, token)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed iterating employee token rows: %w", rows.Err())
	}

	return tokens, nil
}

func (t *NotificationTasks) GetActiveAdminTokens(
	ctx context.Context,
	tx pgx.Tx,
) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT fcm_token
		FROM account.fcm_tokens
		WHERE role = 'admin'
		  AND is_active = TRUE;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active admin tokens: %w", err)
	}
	defer rows.Close()

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if scanErr := rows.Scan(&token); scanErr != nil {
			return nil, fmt.Errorf("failed to scan admin token: %w", scanErr)
		}
		tokens = append(tokens, token)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed iterating admin token rows: %w", rows.Err())
	}

	return tokens, nil
}

func (t *NotificationTasks) GetActiveCustomerTokens(
	ctx context.Context,
	tx pgx.Tx,
	customerID string,
) ([]string, error) {
	rows, err := tx.Query(ctx, `
		SELECT t.fcm_token
		FROM account.fcm_tokens t
		JOIN account.customers c
		  ON c.account_id = t.account_id
		WHERE c.id = $1
		  AND t.role = 'customer'
		  AND t.is_active = TRUE;
	`, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch active customer tokens: %w", err)
	}
	defer rows.Close()

	tokens := make([]string, 0)
	for rows.Next() {
		var token string
		if scanErr := rows.Scan(&token); scanErr != nil {
			return nil, fmt.Errorf("failed to scan customer token: %w", scanErr)
		}
		tokens = append(tokens, token)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed iterating customer token rows: %w", rows.Err())
	}

	return tokens, nil
}

func (t *NotificationTasks) DeactivateEmployeeToken(
	ctx context.Context,
	tx pgx.Tx,
	employeeID string,
	fcmToken string,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE account.fcm_tokens t
		SET is_active = FALSE, updated_at = NOW()
		FROM account.employees e
		WHERE e.id = $1
		  AND t.account_id = e.account_id
		  AND t.role = 'employee'
		  AND t.fcm_token = $2;
	`, employeeID, fcmToken)
	if err != nil {
		return fmt.Errorf("failed to deactivate employee token: %w", err)
	}

	return nil
}

func (t *NotificationTasks) DeactivateAdminToken(
	ctx context.Context,
	tx pgx.Tx,
	adminID string,
	fcmToken string,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE account.fcm_tokens t
		SET is_active = FALSE, updated_at = NOW()
		FROM account.admins a
		WHERE a.id = $1
		  AND t.account_id = a.account_id
		  AND t.role = 'admin'
		  AND t.fcm_token = $2;
	`, adminID, fcmToken)
	if err != nil {
		return fmt.Errorf("failed to deactivate admin token: %w", err)
	}

	return nil
}

func (t *NotificationTasks) DeactivateCustomerToken(
	ctx context.Context,
	tx pgx.Tx,
	customerID string,
	fcmToken string,
) error {
	_, err := tx.Exec(ctx, `
		UPDATE account.fcm_tokens t
		SET is_active = FALSE, updated_at = NOW()
		FROM account.customers c
		WHERE c.id = $1
		  AND t.account_id = c.account_id
		  AND t.role = 'customer'
		  AND t.fcm_token = $2;
	`, customerID, fcmToken)
	if err != nil {
		return fmt.Errorf("failed to deactivate customer token: %w", err)
	}

	return nil
}
