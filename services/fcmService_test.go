package services

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/require"

	"handworks-api/utils"
)

type mockFCMMessagingClient struct {
	sendFn                 func(context.Context, *messaging.Message) (string, error)
	sendEachForMulticastFn func(context.Context, *messaging.MulticastMessage) (*messaging.BatchResponse, error)
	subscribeFn            func(context.Context, []string, string) (*messaging.TopicManagementResponse, error)
	unsubscribeFn          func(context.Context, []string, string) (*messaging.TopicManagementResponse, error)
}

func (m *mockFCMMessagingClient) Send(ctx context.Context, msg *messaging.Message) (string, error) {
	if m.sendFn == nil {
		return "", nil
	}
	return m.sendFn(ctx, msg)
}

func (m *mockFCMMessagingClient) SendEachForMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
	if m.sendEachForMulticastFn == nil {
		return &messaging.BatchResponse{}, nil
	}
	return m.sendEachForMulticastFn(ctx, message)
}

func (m *mockFCMMessagingClient) SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	if m.subscribeFn == nil {
		return &messaging.TopicManagementResponse{}, nil
	}
	return m.subscribeFn(ctx, tokens, topic)
}

func (m *mockFCMMessagingClient) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	if m.unsubscribeFn == nil {
		return &messaging.TopicManagementResponse{}, nil
	}
	return m.unsubscribeFn(ctx, tokens, topic)
}

func newTestFCMService(t *testing.T, client fcmMessagingClient) *FCMService {
	t.Helper()

	logger, err := utils.NewLogger()
	require.NoError(t, err)

	return &FCMService{
		logger:              logger,
		client:              client,
		adminTopic:          "admins",
		employeeTopicPrefix: "employee_",
		customerTopicPrefix: "customer_",
	}
}

func TestFCMService_SendToEmployee_UsesEmployeeTopic(t *testing.T) {
	var sent *messaging.Message

	client := &mockFCMMessagingClient{
		sendFn: func(_ context.Context, msg *messaging.Message) (string, error) {
			sent = msg
			return "message-id", nil
		},
	}

	svc := newTestFCMService(t, client)

	err := svc.SendToEmployee(context.Background(), "emp-9", "booking.created", map[string]any{"id": "b1"})

	require.NoError(t, err)
	require.NotNil(t, sent)
	require.Equal(t, "employee_emp-9", sent.Topic)
}

func TestFCMService_SendToAdmin_SendsExpectedMessage(t *testing.T) {
	var sent *messaging.Message

	client := &mockFCMMessagingClient{
		sendFn: func(_ context.Context, msg *messaging.Message) (string, error) {
			sent = msg
			return "message-id", nil
		},
	}

	svc := newTestFCMService(t, client)

	err := svc.SendToAdmin(context.Background(), "booking.created", map[string]any{"id": "b1"})

	require.NoError(t, err)
	require.NotNil(t, sent)
	require.Equal(t, "admins", sent.Topic)
	require.Equal(t, "booking.created", sent.Data["event"])
	require.Contains(t, sent.Data["data"], `"id":"b1"`)
}

func TestFCMService_SubscribeTokenToCustomerTopic_ValidatesToken(t *testing.T) {
	svc := newTestFCMService(t, &mockFCMMessagingClient{})

	topic, err := svc.SubscribeTokenToCustomerTopic(context.Background(), "   ", "cust-1")

	require.Error(t, err)
	require.Empty(t, topic)
	require.Contains(t, err.Error(), "token is required")
}

func TestFCMService_SubscribeTokenToEmployeeTopic_UsesEmployeeTopic(t *testing.T) {
	var capturedTopic string

	client := &mockFCMMessagingClient{
		subscribeFn: func(_ context.Context, _ []string, topic string) (*messaging.TopicManagementResponse, error) {
			capturedTopic = topic
			return &messaging.TopicManagementResponse{}, nil
		},
	}

	svc := newTestFCMService(t, client)

	topic, err := svc.SubscribeTokenToEmployeeTopic(context.Background(), "token-1", "emp-7")

	require.NoError(t, err)
	require.Equal(t, "employee_emp-7", topic)
	require.Equal(t, "employee_emp-7", capturedTopic)
}

func TestFCMService_SendToTokens_EmptyInputReturnsNil(t *testing.T) {
	svc := newTestFCMService(t, &mockFCMMessagingClient{})

	invalidTokens, err := svc.SendToTokens(context.Background(), nil, "booking.created", map[string]any{"id": "b1"})

	require.NoError(t, err)
	require.Nil(t, invalidTokens)
}

func TestFCMService_SendToTokens_ReturnsClientError(t *testing.T) {
	client := &mockFCMMessagingClient{
		sendEachForMulticastFn: func(_ context.Context, _ *messaging.MulticastMessage) (*messaging.BatchResponse, error) {
			return nil, errors.New("multicast failed")
		},
	}

	svc := newTestFCMService(t, client)

	invalidTokens, err := svc.SendToTokens(context.Background(), []string{"token-1"}, "booking.created", map[string]any{"id": "b1"})

	require.Error(t, err)
	require.Nil(t, invalidTokens)
	require.Contains(t, err.Error(), "fcm multicast send failed")
}
