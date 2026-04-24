package services

import (
	"context"
	"fmt"
	"handworks-api/config"
	"handworks-api/tasks"
	"handworks-api/utils"
	"strings"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/api/option"
)

// --- Account Service ---
type AccountService struct {
	DB     *pgxpool.Pool
	Logger *utils.Logger
	Tasks  *tasks.AccountTasks
}

func NewAccountService(db *pgxpool.Pool, logger *utils.Logger) *AccountService {
	return &AccountService{DB: db, Logger: logger, Tasks: &tasks.AccountTasks{}}
}

// --- Inventory Service ---
type InventoryService struct {
	DB     *pgxpool.Pool
	Logger *utils.Logger
	Tasks  *tasks.InventoryTasks
}

func NewInventoryService(db *pgxpool.Pool, logger *utils.Logger) *InventoryService {
	return &InventoryService{DB: db, Logger: logger, Tasks: &tasks.InventoryTasks{}}
}

// --- Booking Service ---

type BookingService struct {
	DB          *pgxpool.Pool
	Logger      *utils.Logger
	Tasks       *tasks.BookingTasks
	PaymentPort tasks.PaymentPort
}

func NewBookingService(db *pgxpool.Pool, logger *utils.Logger, paymentPort tasks.PaymentPort) *BookingService {
	return &BookingService{DB: db, Logger: logger, Tasks: &tasks.BookingTasks{}, PaymentPort: paymentPort}
}

// --- Payment Service ---
type PaymentService struct {
	DB             *pgxpool.Pool
	Logger         *utils.Logger
	Tasks          *tasks.PaymentTasks
	PaymongoClient *config.PaymongoClient
}

func NewPaymentService(db *pgxpool.Pool, logger *utils.Logger, paymongoClient *config.PaymongoClient) *PaymentService {
	return &PaymentService{DB: db, Logger: logger, Tasks: &tasks.PaymentTasks{}, PaymongoClient: paymongoClient}
}

// Admin Service
type AdminService struct {
	DB          *pgxpool.Pool
	Logger      *utils.Logger
	Tasks       *tasks.AdminTasks
	AccountPort tasks.AccountPort
}

func NewAdminService(db *pgxpool.Pool, logger *utils.Logger, accountService tasks.AccountPort) *AdminService {
	return &AdminService{DB: db, Logger: logger, Tasks: &tasks.AdminTasks{}, AccountPort: accountService}
}

// FCM Service

type FCMService struct {
	logger              *utils.Logger
	client              fcmMessagingClient
	adminTopic          string
	employeeTopicPrefix string
	customerTopicPrefix string
}

type fcmMessagingClient interface {
	Send(ctx context.Context, msg *messaging.Message) (string, error)
	SendEachForMulticast(ctx context.Context, message *messaging.MulticastMessage) (*messaging.BatchResponse, error)
	SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error)
	UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error)
}

func NewFCMService(
	ctx context.Context,
	logger *utils.Logger,
	credentialsFile string,
	projectID string,
	adminTopic string,
	employeeTopicPrefix string,
	customerTopicPrefix string,
) (*FCMService, error) {
	if strings.TrimSpace(credentialsFile) == "" {
		return nil, fmt.Errorf("firebase credentials file is required")
	}

	if strings.TrimSpace(projectID) == "" {
		return nil, fmt.Errorf("firebase project id is required")
	}

	if strings.TrimSpace(adminTopic) == "" {
		return nil, fmt.Errorf("admin topic is required")
	}

	if strings.TrimSpace(employeeTopicPrefix) == "" {
		return nil, fmt.Errorf("employee topic prefix is required")
	}

	if strings.TrimSpace(customerTopicPrefix) == "" {
		return nil, fmt.Errorf("customer topic prefix is required")
	}

	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID},
		option.WithAuthCredentialsFile(option.ServiceAccount, credentialsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase app: %w", err)
	}

	client, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize firebase messaging client: %w", err)
	}

	return &FCMService{
		logger:              logger,
		client:              client,
		adminTopic:          adminTopic,
		employeeTopicPrefix: employeeTopicPrefix,
		customerTopicPrefix: customerTopicPrefix,
	}, nil
}

// Notification Service
type NotificationService struct {
	DB     notificationTxBeginner
	Logger *utils.Logger
	FCM    notificationFCMPort
	Tasks  notificationTasker
}

type notificationTxBeginner interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

type notificationFCMPort interface {
	SubscribeTokenToAdminTopic(ctx context.Context, token string) (string, error)
	SubscribeTokenToEmployeeTopic(ctx context.Context, token string, employeeID string) (string, error)
	SubscribeTokenToCustomerTopic(ctx context.Context, token string, customerID string) (string, error)
	UnsubscribeTokenFromAdminTopic(ctx context.Context, token string) error
	UnsubscribeTokenFromEmployeeTopic(ctx context.Context, token string, employeeID string) error
	UnsubscribeTokenFromCustomerTopic(ctx context.Context, token string, customerID string) error
	SendToTokens(ctx context.Context, tokens []string, event string, payload any) ([]string, error)
}

type notificationTasker interface {
	UpsertEmployeeFCMToken(ctx context.Context, tx pgx.Tx, employeeID string, installationID string, fcmToken string, platform string) error
	DeactivateToken(ctx context.Context, tx pgx.Tx, fcmToken string) error
	UpsertAdminFCMToken(ctx context.Context, tx pgx.Tx, adminID string, installationID string, fcmToken string, platform string) error
	UpsertCustomerFCMToken(ctx context.Context, tx pgx.Tx, customerID string, installationID string, fcmToken string, platform string) error
	GetActiveEmployeeTokens(ctx context.Context, tx pgx.Tx, employeeID string) ([]string, error)
	GetActiveAdminTokens(ctx context.Context, tx pgx.Tx) ([]string, error)
	GetActiveCustomerTokens(ctx context.Context, tx pgx.Tx, customerID string) ([]string, error)
	DeactivateEmployeeToken(ctx context.Context, tx pgx.Tx, employeeID string, fcmToken string) error
	DeactivateAdminToken(ctx context.Context, tx pgx.Tx, adminID string, fcmToken string) error
	DeactivateCustomerToken(ctx context.Context, tx pgx.Tx, customerID string, fcmToken string) error
}

func NewNotificationService(db *pgxpool.Pool, logger *utils.Logger, fcm *FCMService) *NotificationService {
	return &NotificationService{DB: db, Logger: logger, FCM: fcm, Tasks: &tasks.NotificationTasks{}}
}
