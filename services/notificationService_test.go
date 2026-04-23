package services

import (
	"context"
	"errors"
	"testing"

	"handworks-api/types"
	"handworks-api/utils"

	"github.com/jackc/pgx/v5"
	pgxmock "github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

type mockNotificationFCM struct {
	subscribeAdminFn      func(context.Context, string) (string, error)
	subscribeEmployeeFn   func(context.Context, string, string) (string, error)
	subscribeCustomerFn   func(context.Context, string, string) (string, error)
	unsubscribeAdminFn    func(context.Context, string) error
	unsubscribeEmployeeFn func(context.Context, string, string) error
	unsubscribeCustomerFn func(context.Context, string, string) error
	sendToTokensFn        func(context.Context, []string, string, any) ([]string, error)
}

func (m *mockNotificationFCM) SubscribeTokenToAdminTopic(ctx context.Context, token string) (string, error) {
	if m.subscribeAdminFn == nil {
		return "", nil
	}
	return m.subscribeAdminFn(ctx, token)
}

func (m *mockNotificationFCM) SubscribeTokenToEmployeeTopic(ctx context.Context, token string, employeeID string) (string, error) {
	if m.subscribeEmployeeFn == nil {
		return "", nil
	}
	return m.subscribeEmployeeFn(ctx, token, employeeID)
}

func (m *mockNotificationFCM) SubscribeTokenToCustomerTopic(ctx context.Context, token string, customerID string) (string, error) {
	if m.subscribeCustomerFn == nil {
		return "", nil
	}
	return m.subscribeCustomerFn(ctx, token, customerID)
}

func (m *mockNotificationFCM) UnsubscribeTokenFromAdminTopic(ctx context.Context, token string) error {
	if m.unsubscribeAdminFn == nil {
		return nil
	}
	return m.unsubscribeAdminFn(ctx, token)
}

func (m *mockNotificationFCM) UnsubscribeTokenFromEmployeeTopic(ctx context.Context, token string, employeeID string) error {
	if m.unsubscribeEmployeeFn == nil {
		return nil
	}
	return m.unsubscribeEmployeeFn(ctx, token, employeeID)
}

func (m *mockNotificationFCM) UnsubscribeTokenFromCustomerTopic(ctx context.Context, token string, customerID string) error {
	if m.unsubscribeCustomerFn == nil {
		return nil
	}
	return m.unsubscribeCustomerFn(ctx, token, customerID)
}

func (m *mockNotificationFCM) SendToTokens(ctx context.Context, tokens []string, event string, payload any) ([]string, error) {
	if m.sendToTokensFn == nil {
		return nil, nil
	}
	return m.sendToTokensFn(ctx, tokens, event, payload)
}

type mockNotificationTasks struct {
	upsertEmployeeFn     func(context.Context, pgx.Tx, string, string, string, string) error
	deactivateTokenFn    func(context.Context, pgx.Tx, string) error
	upsertAdminFn        func(context.Context, pgx.Tx, string, string, string, string) error
	upsertCustomerFn     func(context.Context, pgx.Tx, string, string, string, string) error
	getActiveEmployeeFn  func(context.Context, pgx.Tx, string) ([]string, error)
	getActiveAdminFn     func(context.Context, pgx.Tx) ([]string, error)
	getActiveCustomerFn  func(context.Context, pgx.Tx, string) ([]string, error)
	deactivateEmployeeFn func(context.Context, pgx.Tx, string, string) error
	deactivateAdminFn    func(context.Context, pgx.Tx, string, string) error
	deactivateCustomerFn func(context.Context, pgx.Tx, string, string) error
}

func (m *mockNotificationTasks) UpsertEmployeeFCMToken(ctx context.Context, tx pgx.Tx, employeeID string, installationID string, fcmToken string, platform string) error {
	if m.upsertEmployeeFn == nil {
		return nil
	}
	return m.upsertEmployeeFn(ctx, tx, employeeID, installationID, fcmToken, platform)
}

func (m *mockNotificationTasks) DeactivateToken(ctx context.Context, tx pgx.Tx, fcmToken string) error {
	if m.deactivateTokenFn == nil {
		return nil
	}
	return m.deactivateTokenFn(ctx, tx, fcmToken)
}

func (m *mockNotificationTasks) UpsertAdminFCMToken(ctx context.Context, tx pgx.Tx, adminID string, installationID string, fcmToken string, platform string) error {
	if m.upsertAdminFn == nil {
		return nil
	}
	return m.upsertAdminFn(ctx, tx, adminID, installationID, fcmToken, platform)
}

func (m *mockNotificationTasks) UpsertCustomerFCMToken(ctx context.Context, tx pgx.Tx, customerID string, installationID string, fcmToken string, platform string) error {
	if m.upsertCustomerFn == nil {
		return nil
	}
	return m.upsertCustomerFn(ctx, tx, customerID, installationID, fcmToken, platform)
}

func (m *mockNotificationTasks) GetActiveEmployeeTokens(ctx context.Context, tx pgx.Tx, employeeID string) ([]string, error) {
	if m.getActiveEmployeeFn == nil {
		return nil, nil
	}
	return m.getActiveEmployeeFn(ctx, tx, employeeID)
}

func (m *mockNotificationTasks) GetActiveAdminTokens(ctx context.Context, tx pgx.Tx) ([]string, error) {
	if m.getActiveAdminFn == nil {
		return nil, nil
	}
	return m.getActiveAdminFn(ctx, tx)
}

func (m *mockNotificationTasks) GetActiveCustomerTokens(ctx context.Context, tx pgx.Tx, customerID string) ([]string, error) {
	if m.getActiveCustomerFn == nil {
		return nil, nil
	}
	return m.getActiveCustomerFn(ctx, tx, customerID)
}

func (m *mockNotificationTasks) DeactivateEmployeeToken(ctx context.Context, tx pgx.Tx, employeeID string, fcmToken string) error {
	if m.deactivateEmployeeFn == nil {
		return nil
	}
	return m.deactivateEmployeeFn(ctx, tx, employeeID, fcmToken)
}

func (m *mockNotificationTasks) DeactivateAdminToken(ctx context.Context, tx pgx.Tx, adminID string, fcmToken string) error {
	if m.deactivateAdminFn == nil {
		return nil
	}
	return m.deactivateAdminFn(ctx, tx, adminID, fcmToken)
}

func (m *mockNotificationTasks) DeactivateCustomerToken(ctx context.Context, tx pgx.Tx, customerID string, fcmToken string) error {
	if m.deactivateCustomerFn == nil {
		return nil
	}
	return m.deactivateCustomerFn(ctx, tx, customerID, fcmToken)
}

func newTestNotificationService(t *testing.T, db notificationTxBeginner, fcm notificationFCMPort, tasker notificationTasker) *NotificationService {
	t.Helper()

	logger, err := utils.NewLogger()
	require.NoError(t, err)

	return &NotificationService{
		DB:     db,
		Logger: logger,
		FCM:    fcm,
		Tasks:  tasker,
	}
}

func TestNotificationService_SubscribeToken_AdminSuccess(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()

	pool.ExpectBegin()
	pool.ExpectCommit()

	fcm := &mockNotificationFCM{
		subscribeAdminFn: func(_ context.Context, token string) (string, error) {
			require.Equal(t, "token-1", token)
			return "admins", nil
		},
	}

	tasks := &mockNotificationTasks{
		upsertAdminFn: func(_ context.Context, _ pgx.Tx, adminID string, installationID string, fcmToken string, platform string) error {
			require.Equal(t, "admin-1", adminID)
			require.Equal(t, "inst-1", installationID)
			require.Equal(t, "token-1", fcmToken)
			require.Equal(t, "unknown", platform)
			return nil
		},
	}

	svc := newTestNotificationService(t, pool, fcm, tasks)

	resp, err := svc.SubscribeToken(context.Background(), &types.SubscribeNotificationRequest{
		Token:          "token-1",
		Role:           "admin",
		AdminID:        "admin-1",
		InstallationID: "inst-1",
	})

	require.NoError(t, err)
	require.True(t, resp.Ok)
	require.Equal(t, []string{"admins"}, resp.Topics)
	require.NoError(t, pool.ExpectationsWereMet())
}

func TestNotificationService_UnsubscribeToken_EmployeeContinuesWhenFCMFails(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()

	pool.ExpectBegin()
	pool.ExpectCommit()

	deactivateCalled := false

	fcm := &mockNotificationFCM{
		unsubscribeEmployeeFn: func(_ context.Context, _ string, _ string) error {
			return errors.New("fcm temporary failure")
		},
	}

	tasks := &mockNotificationTasks{
		deactivateEmployeeFn: func(_ context.Context, _ pgx.Tx, employeeID string, token string) error {
			deactivateCalled = true
			require.Equal(t, "emp-1", employeeID)
			require.Equal(t, "token-1", token)
			return nil
		},
	}

	svc := newTestNotificationService(t, pool, fcm, tasks)

	resp, err := svc.UnsubscribeToken(context.Background(), &types.UnsubscribeNotificationRequest{
		Token:      "token-1",
		Role:       "employee",
		EmployeeID: "emp-1",
	})

	require.NoError(t, err)
	require.True(t, resp.Ok)
	require.True(t, deactivateCalled)
	require.NoError(t, pool.ExpectationsWereMet())
}

func TestNotificationService_SubscribeToken_UnsupportedRole(t *testing.T) {
	pool, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer pool.Close()

	svc := newTestNotificationService(t, pool, &mockNotificationFCM{}, &mockNotificationTasks{})

	resp, err := svc.SubscribeToken(context.Background(), &types.SubscribeNotificationRequest{
		Token:          "token-1",
		Role:           "manager",
		InstallationID: "inst-1",
	})

	require.Error(t, err)
	require.Nil(t, resp)
	require.Contains(t, err.Error(), "unsupported role")
	// No tx is expected for unsupported roles.
	require.NoError(t, pool.ExpectationsWereMet())
}
