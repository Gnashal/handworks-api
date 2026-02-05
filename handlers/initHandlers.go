package handlers

import (
	"handworks-api/services"
	"handworks-api/utils"
)

// --- Account Handler ---
type AccountHandler struct {
	Service *services.AccountService
	Logger  *utils.Logger
}

func NewAccountHandler(service *services.AccountService, logger *utils.Logger) *AccountHandler {
	return &AccountHandler{
		Service: service,
		Logger:  logger,
	}
}

// --- Inventory Handler ---
type InventoryHandler struct {
	Service *services.InventoryService
	Logger  *utils.Logger
}

func NewInventoryHandler(service *services.InventoryService, logger *utils.Logger) *InventoryHandler {
	return &InventoryHandler{
		Service: service,
		Logger:  logger,
	}
}

// --- Booking Handler ---
type BookingHandler struct {
	Service *services.BookingService
	Logger  *utils.Logger
}

func NewBookingHandler(service *services.BookingService, logger *utils.Logger) *BookingHandler {
	return &BookingHandler{
		Service: service,
		Logger:  logger,
	}
}

// --- Payment Handler ---
type PaymentHandler struct {
	Service *services.PaymentService
	Logger  *utils.Logger
}

func NewPaymentHandler(service *services.PaymentService, logger *utils.Logger) *PaymentHandler {
	return &PaymentHandler{
		Service: service,
		Logger:  logger,
	}
}

type AdminHandler struct {
	Service *services.AdminService
	Logger  *utils.Logger
}

func NewAdminHandler(service *services.AdminService, logger *utils.Logger) *AdminHandler {
	return &AdminHandler{
		Service: service,
		Logger:  logger,
	}
}
