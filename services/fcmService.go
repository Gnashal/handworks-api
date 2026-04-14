package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"firebase.google.com/go/v4/messaging"
)

func (s *FCMService) SendToAdmin(ctx context.Context, event string, payload any) error {
	return s.sendToTopic(ctx, s.adminTopic, event, payload)
}

func (s *FCMService) SendToEmployee(ctx context.Context, employeeID string, event string, payload any) error {
	topic := s.topicForEmployee(employeeID)
	if topic == "" {
		return fmt.Errorf("employee topic cannot be empty")
	}

	return s.sendToTopic(ctx, topic, event, payload)
}

func (s *FCMService) SendToCustomer(ctx context.Context, customerID string, event string, payload any) error {
	topic := s.topicForCustomer(customerID)
	if topic == "" {
		return fmt.Errorf("customer topic cannot be empty")
	}

	return s.sendToTopic(ctx, topic, event, payload)
}

func (s *FCMService) SubscribeTokenToAdminTopic(ctx context.Context, token string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("fcm service is not initialized")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token is required")
	}

	if _, err := s.client.SubscribeToTopic(ctx, []string{token}, s.adminTopic); err != nil {
		return "", fmt.Errorf("failed to subscribe token to admin topic: %w", err)
	}

	return s.adminTopic, nil
}

func (s *FCMService) SubscribeTokenToEmployeeTopic(ctx context.Context, token string, employeeID string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("fcm service is not initialized")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token is required")
	}

	topic := s.topicForEmployee(employeeID)
	if topic == "" {
		return "", fmt.Errorf("employee topic cannot be empty")
	}

	if _, err := s.client.SubscribeToTopic(ctx, []string{token}, topic); err != nil {
		return "", fmt.Errorf("failed to subscribe token to employee topic: %w", err)
	}

	return topic, nil
}

func (s *FCMService) SubscribeTokenToCustomerTopic(ctx context.Context, token string, customerID string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("fcm service is not initialized")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return "", fmt.Errorf("token is required")
	}

	topic := s.topicForCustomer(customerID)
	if topic == "" {
		return "", fmt.Errorf("customer topic cannot be empty")
	}

	if _, err := s.client.SubscribeToTopic(ctx, []string{token}, topic); err != nil {
		return "", fmt.Errorf("failed to subscribe token to customer topic: %w", err)
	}

	return topic, nil
}

func (s *FCMService) UnsubscribeTokenFromAdminTopic(ctx context.Context, token string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("fcm service is not initialized")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is required")
	}

	if _, err := s.client.UnsubscribeFromTopic(ctx, []string{token}, s.adminTopic); err != nil {
		return fmt.Errorf("failed to unsubscribe token from admin topic: %w", err)
	}

	return nil
}

func (s *FCMService) UnsubscribeTokenFromEmployeeTopic(ctx context.Context, token string, employeeID string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("fcm service is not initialized")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is required")
	}

	topic := s.topicForEmployee(employeeID)
	if topic == "" {
		return fmt.Errorf("employee topic cannot be empty")
	}

	if _, err := s.client.UnsubscribeFromTopic(ctx, []string{token}, topic); err != nil {
		return fmt.Errorf("failed to unsubscribe token from employee topic: %w", err)
	}

	return nil
}

func (s *FCMService) UnsubscribeTokenFromCustomerTopic(ctx context.Context, token string, customerID string) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("fcm service is not initialized")
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("token is required")
	}

	topic := s.topicForCustomer(customerID)
	if topic == "" {
		return fmt.Errorf("customer topic cannot be empty")
	}

	if _, err := s.client.UnsubscribeFromTopic(ctx, []string{token}, topic); err != nil {
		return fmt.Errorf("failed to unsubscribe token from customer topic: %w", err)
	}

	return nil
}

func (s *FCMService) topicForEmployee(employeeID string) string {
	return strings.TrimSpace(s.employeeTopicPrefix + employeeID)
}

func (s *FCMService) topicForCustomer(customerID string) string {
	return strings.TrimSpace(s.customerTopicPrefix + customerID)
}

func (s *FCMService) sendToTopic(ctx context.Context, topic string, event string, payload any) error {
	if s == nil || s.client == nil {
		return fmt.Errorf("fcm service is not initialized")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal fcm payload: %w", err)
	}

	msg := &messaging.Message{
		Topic: topic,
		Data: map[string]string{
			"event": event,
			"data":  string(payloadBytes),
		},
	}

	messageID, err := s.client.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("fcm send failed: %w", err)
	}

	s.logger.Debug("FCM message sent (topic=%s, event=%s, id=%s)", topic, event, messageID)
	return nil
}

func (s *FCMService) SendToTokens(
	ctx context.Context,
	tokens []string,
	event string,
	payload any,
) ([]string, error) {
	if s == nil || s.client == nil {
		return nil, fmt.Errorf("fcm service is not initialized")
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fcm payload: %w", err)
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Data: map[string]string{
			"event": event,
			"data":  string(payloadBytes),
		},
	}

	res, err := s.client.SendEachForMulticast(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("fcm multicast send failed: %w", err)
	}

	invalid := make([]string, 0)
	for i, r := range res.Responses {
		if r.Success {
			continue
		}
		if messaging.IsUnregistered(r.Error) || messaging.IsInvalidArgument(r.Error) {
			invalid = append(invalid, tokens[i])
		}
	}

	return invalid, nil
}
