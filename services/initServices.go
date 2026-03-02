package services

import (
	"handworks-api/tasks"
	"handworks-api/utils"

	"github.com/jackc/pgx/v5/pgxpool"
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
	DB     *pgxpool.Pool
	Logger *utils.Logger
	Tasks  *tasks.PaymentTasks
}

func NewPaymentService(db *pgxpool.Pool, logger *utils.Logger) *PaymentService {
	return &PaymentService{DB: db, Logger: logger, Tasks: &tasks.PaymentTasks{}}
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
