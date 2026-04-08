package services

import (
	"context"
	"fmt"
	"handworks-api/types"
	"strings"

	"github.com/jackc/pgx/v5"
)

func (s *NotificationService) withTx(
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

func (s *NotificationService) SubscribeToken(ctx context.Context, req *types.SubscribeNotificationRequest) (*types.SubscribeNotificationResponse, error) {
	if s == nil || s.FCM == nil {
		return nil, fmt.Errorf("notification service is unavailable")
	}
	if s.DB == nil || s.Tasks == nil {
		return nil, fmt.Errorf("notification persistence is unavailable")
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	req.Platform = strings.ToLower(strings.TrimSpace(req.Platform))
	if req.Platform == "" {
		req.Platform = "unknown"
	}
	topics := make([]string, 0, 1)

	switch role {
	case "admin":
		req.AdminID = strings.TrimSpace(req.AdminID)
		req.InstallationID = strings.TrimSpace(req.InstallationID)
		if req.AdminID == "" {
			return nil, fmt.Errorf("adminId is required for admin subscriptions")
		}
		if req.InstallationID == "" {
			return nil, fmt.Errorf("installationId is required for admin subscriptions")
		}

		topic, err := s.FCM.SubscribeTokenToAdminTopic(ctx, req.Token)
		if err != nil {
			return nil, err
		}

		err = s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.UpsertAdminFCMToken(ctx, tx, req.AdminID, req.InstallationID, req.Token, req.Platform)
		})
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)

	case "employee":
		req.EmployeeID = strings.TrimSpace(req.EmployeeID)
		req.InstallationID = strings.TrimSpace(req.InstallationID)
		if req.EmployeeID == "" {
			return nil, fmt.Errorf("employeeId is required for employee subscriptions")
		}
		if req.InstallationID == "" {
			return nil, fmt.Errorf("installationId is required for employee subscriptions")
		}

		topic, err := s.FCM.SubscribeTokenToEmployeeTopic(ctx, req.Token, req.EmployeeID)
		if err != nil {
			return nil, err
		}

		err = s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.UpsertEmployeeFCMToken(ctx, tx, req.EmployeeID, req.InstallationID, req.Token, req.Platform)
		})
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)

	case "customer":
		req.CustomerID = strings.TrimSpace(req.CustomerID)
		req.InstallationID = strings.TrimSpace(req.InstallationID)
		if req.CustomerID == "" {
			return nil, fmt.Errorf("customerId is required for customer subscriptions")
		}
		if req.InstallationID == "" {
			return nil, fmt.Errorf("installationId is required for customer subscriptions")
		}

		topic, err := s.FCM.SubscribeTokenToCustomerTopic(ctx, req.Token, req.CustomerID)
		if err != nil {
			return nil, err
		}

		err = s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.UpsertCustomerFCMToken(ctx, tx, req.CustomerID, req.InstallationID, req.Token, req.Platform)
		})
		if err != nil {
			return nil, err
		}

		topics = append(topics, topic)

	default:
		return nil, fmt.Errorf("unsupported role: %s", req.Role)
	}

	return &types.SubscribeNotificationResponse{
		Ok:     true,
		Topics: topics,
	}, nil
}

func (s *NotificationService) UnsubscribeToken(ctx context.Context, req *types.UnsubscribeNotificationRequest) (*types.UnsubscribeNotificationResponse, error) {
	if s == nil || s.FCM == nil {
		return nil, fmt.Errorf("notification service is unavailable")
	}
	if s.DB == nil || s.Tasks == nil {
		return nil, fmt.Errorf("notification persistence is unavailable")
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))

	switch role {
	case "admin":
		req.AdminID = strings.TrimSpace(req.AdminID)
		if req.AdminID == "" {
			return nil, fmt.Errorf("adminId is required for admin unsubscription")
		}

		if err := s.FCM.UnsubscribeTokenFromAdminTopic(ctx, req.Token); err != nil {
			s.Logger.Warn("failed to unsubscribe token from admin topic: %v", err)
		}

		if err := s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.DeactivateAdminToken(ctx, tx, req.AdminID, req.Token)
		}); err != nil {
			return nil, err
		}

	case "customer":
		req.CustomerID = strings.TrimSpace(req.CustomerID)
		if req.CustomerID == "" {
			return nil, fmt.Errorf("customerId is required for customer unsubscription")
		}

		if err := s.FCM.UnsubscribeTokenFromCustomerTopic(ctx, req.Token, req.CustomerID); err != nil {
			s.Logger.Warn("failed to unsubscribe token from customer topic: %v", err)
		}

		if err := s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.DeactivateCustomerToken(ctx, tx, req.CustomerID, req.Token)
		}); err != nil {
			return nil, err
		}

	case "employee":
		req.EmployeeID = strings.TrimSpace(req.EmployeeID)
		if req.EmployeeID == "" {
			return nil, fmt.Errorf("employeeId is required for employee unsubscription")
		}

		if err := s.FCM.UnsubscribeTokenFromEmployeeTopic(ctx, req.Token, req.EmployeeID); err != nil {
			s.Logger.Warn("failed to unsubscribe token from employee topic: %v", err)
		}

		if err := s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.DeactivateEmployeeToken(ctx, tx, req.EmployeeID, req.Token)
		}); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unsupported role: %s", req.Role)
	}

	return &types.UnsubscribeNotificationResponse{Ok: true}, nil
}

func (s *NotificationService) SendToEmployee(ctx context.Context, employeeID string, event string, payload any) error {
	if s == nil || s.FCM == nil || s.DB == nil || s.Tasks == nil {
		return fmt.Errorf("notification service is unavailable")
	}

	employeeID = strings.TrimSpace(employeeID)
	if employeeID == "" {
		return fmt.Errorf("employeeID is required")
	}

	var tokens []string
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		var queryErr error
		tokens, queryErr = s.Tasks.GetActiveEmployeeTokens(ctx, tx, employeeID)
		return queryErr
	})
	if err != nil {
		return err
	}

	invalidTokens, err := s.FCM.SendToTokens(ctx, tokens, event, payload)
	if err != nil {
		return err
	}

	for _, invalidToken := range invalidTokens {
		deactivateErr := s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.DeactivateEmployeeToken(ctx, tx, employeeID, invalidToken)
		})
		if deactivateErr != nil {
			s.Logger.Warn("failed to deactivate invalid employee token: %v", deactivateErr)
		}
	}

	return nil
}

func (s *NotificationService) SendToCustomer(ctx context.Context, customerID string, event string, payload any) error {
	if s == nil || s.FCM == nil || s.DB == nil || s.Tasks == nil {
		return fmt.Errorf("notification service is unavailable")
	}

	customerID = strings.TrimSpace(customerID)
	if customerID == "" {
		return fmt.Errorf("customerID is required")
	}

	var tokens []string
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		var queryErr error
		tokens, queryErr = s.Tasks.GetActiveCustomerTokens(ctx, tx, customerID)
		return queryErr
	})
	if err != nil {
		return err
	}

	invalidTokens, err := s.FCM.SendToTokens(ctx, tokens, event, payload)
	if err != nil {
		return err
	}

	for _, invalidToken := range invalidTokens {
		deactivateErr := s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.DeactivateCustomerToken(ctx, tx, customerID, invalidToken)
		})
		if deactivateErr != nil {
			s.Logger.Warn("failed to deactivate invalid customer token: %v", deactivateErr)
		}
	}

	return nil
}

func (s *NotificationService) SendToAdmins(ctx context.Context, event string, payload any) error {
	if s == nil || s.FCM == nil || s.DB == nil || s.Tasks == nil {
		return fmt.Errorf("notification service is unavailable")
	}

	var tokens []string
	err := s.withTx(ctx, func(tx pgx.Tx) error {
		var queryErr error
		tokens, queryErr = s.Tasks.GetActiveAdminTokens(ctx, tx)
		return queryErr
	})
	if err != nil {
		return err
	}

	invalidTokens, err := s.FCM.SendToTokens(ctx, tokens, event, payload)
	if err != nil {
		return err
	}

	for _, invalidToken := range invalidTokens {
		deactivateErr := s.withTx(ctx, func(tx pgx.Tx) error {
			return s.Tasks.DeactivateToken(ctx, tx, invalidToken)
		})
		if deactivateErr != nil {
			s.Logger.Warn("failed to deactivate invalid admin token: %v", deactivateErr)
		}
	}

	return nil
}
