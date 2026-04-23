package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"handworks-api/types"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *PaymentService) withTx(
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

func (s *PaymentService) FetchOrderAndPrices(ctx context.Context, orderId string) (*types.Order, *types.CleaningPrices, error) {
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var order *types.Order
	var prices *types.CleaningPrices

	if err := s.withTx(dbCtx, func(tx pgx.Tx) error {
		var err error
		order, err = s.Tasks.FetchOrderByID(dbCtx, tx, orderId)
		if err != nil {
			return err
		}
		prices, err = s.Tasks.VerifyQuoteAndFetchPrices(dbCtx, tx, order.QuoteID)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch order and prices: %v", err)
		return nil, nil, err
	}

	// Use the order's total amount as the authoritative price for booking calculations
	prices.MainServicePrice = order.TotalAmount

	return order, prices, nil
}

func (s *PaymentService) GetQuotePrices(ctx context.Context, quoteId string) (*types.CleaningPrices, error) {
	dbCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var prices *types.CleaningPrices
	if err := s.withTx(dbCtx, func(tx pgx.Tx) error {
		cleaningPrices, err := s.Tasks.VerifyQuoteAndFetchPrices(dbCtx, tx, quoteId)
		if err != nil {
			return err
		}
		prices = cleaningPrices
		return nil
	}); err != nil {
		s.Logger.Error("Failed to Get Quote Prices: %v", err)
		return nil, err
	}
	return prices, nil
}

func (s *PaymentService) MakePublicQuotation(ctx context.Context, req types.QuoteRequest) (*types.QuoteResponse, error) {
	s.Logger.Info("Generating Quote Preview")
	quotePrev, err := s.Tasks.CalculateQuotePreview(ctx, &req)
	if err != nil {
		s.Logger.Error("Failed to genearte Quote Preview: %v", err)
		return nil, fmt.Errorf("failed to genearte Quote Preview: %v", err)
	}
	addonsBreakdown := s.Tasks.MapAddonstoAddonBreakdown(&quotePrev.Addons)
	return &types.QuoteResponse{
		QuoteId:           quotePrev.ID,
		MainServiceName:   quotePrev.MainService,
		MainServiceTotal:  quotePrev.Subtotal,
		MainServiceHours:  quotePrev.MainServiceHours,
		TotalServiceHours: quotePrev.TotalServiceHours,
		TotalPrice:        quotePrev.TotalPrice,
		AddonTotal:        quotePrev.AddonTotal,
		Addons:            addonsBreakdown,
	}, nil

}

func (s *PaymentService) MakeQuotation(ctx context.Context, req types.QuoteRequest) (*types.QuoteResponse, error) {
	var quoteResponse types.QuoteResponse
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		quote, err := s.Tasks.CreateQuote(ctx, tx, &req)
		if err != nil {
			return fmt.Errorf("failed to create Quote: %v", err)
		}
		quoteResponse.QuoteId = quote.ID
		quoteResponse.MainServiceName = quote.MainService
		quoteResponse.MainServiceTotal = quote.TotalPrice
		quoteResponse.AddonTotal = quote.AddonTotal
		quoteResponse.TotalPrice = quote.TotalPrice
		quoteResponse.Addons = s.Tasks.MapAddonstoAddonBreakdown(&quote.Addons)
		return nil
	}); err != nil {
		s.Logger.Error("Failed to create Quote: %v", err)
		return nil, err
	}
	return &quoteResponse, nil
}

func (s *PaymentService) GetAllQuotesFromCustomer(
	ctx context.Context,
	customerId, startDate, endDate string,
	page, limit int,
) (*types.FetchAllQuotesResponse, error) {

	var result *types.FetchAllQuotesResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllQuotesByCustomer(
			ctx, tx, customerId, startDate, endDate, page, limit, s.Logger,
		)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch Quotes: %v", err)
		return nil, err
	}

	return result, nil
}

func (s *PaymentService) GetAllQuotes(
	ctx context.Context,
	startDate, endDate string,
	page, limit int,
) (*types.FetchAllQuotesResponse, error) {

	var result *types.FetchAllQuotesResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllQuotes(ctx, tx, startDate, endDate, page, limit, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch Quotes: %v", err)
		return nil, err
	}

	return result, nil
}

func (s *PaymentService) GetQuoteByIDForCustomer(ctx context.Context, quoteId, customerId string) (*types.QuoteResponse, error) {
	var quote *types.QuoteResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		quote, err = s.Tasks.FetchQuoteByIDbyCustomer(ctx, tx, quoteId, customerId)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch quote by ID for customer %s: %v", customerId, err)
		return nil, err
	}

	if quote == nil {
		return nil, fmt.Errorf("quote not found for this customer")
	}

	return quote, nil
}

func (s *PaymentService) CreateOrder(ctx context.Context, req types.CreateOrderRequest) (*types.CreateOrderResponse, error) {
	var orderId string
	var order *types.Order
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		orderId, err = s.Tasks.CreateOrder(ctx, tx, req)
		if err != nil {
			return fmt.Errorf("failed to create order for quote %s: %v", req.QuoteID, err)
		}
		order, err = s.Tasks.FetchOrderByID(ctx, tx, orderId)
		if err != nil {
			return fmt.Errorf("failed to create order for quote %s: %v", req.QuoteID, err)
		}
		return err
	}); err != nil {
		s.Logger.Error("Failed to create order for quote %s: %v", req.QuoteID, err)
		return nil, err
	}

	return &types.CreateOrderResponse{Order: *order}, nil
}

func (s *PaymentService) GetOrder(ctx context.Context, orderId string) (*types.Order, error) {
	var order *types.Order

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		order, err = s.Tasks.FetchOrderByID(ctx, tx, orderId)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch order by ID %s: %v", orderId, err)
		return nil, err
	}

	if order == nil {
		return nil, fmt.Errorf("order not found")
	}

	return order, nil
}
func (s *PaymentService) GetOrders(ctx context.Context, page, limit int, startDate, endDate string) (*types.GetOrdersResponse, error) {
	var ordersResponse *types.GetOrdersResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		ordersResponse, err = s.Tasks.FetchOrders(ctx, tx, page, limit, startDate, endDate, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch orders: %v", err)
		return nil, err
	}

	return ordersResponse, nil
}
func (s *PaymentService) GetOrdersByCustomer(ctx context.Context, page, limit int, startDate, endDate, customerId string) (*types.GetOrdersResponse, error) {
	var ordersResponse *types.GetOrdersResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		ordersResponse, err = s.Tasks.FetchOrdersByCustomer(ctx, tx, page, limit, startDate, endDate, customerId, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch orders for customer %s: %v", customerId, err)
		return nil, err
	}

	return ordersResponse, nil
}
func (s *PaymentService) GetPaymentsByOrderID(ctx context.Context, page, limit int, startDate, endDate, orderId string) (*types.GetPaymentsResponse, error) {
	var paymentsResponse *types.GetPaymentsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		paymentsResponse, err = s.Tasks.FetchPaymentsByOrderID(ctx, tx, page, limit, startDate, endDate, orderId)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch payments for order %s: %v", orderId, err)
		return nil, err
	}

	return paymentsResponse, nil
}
func (s *PaymentService) GetPaymentsByCustomerID(ctx context.Context, page, limit int, startDate, endDate, customerId string) (*types.GetPaymentsResponse, error) {
	var paymentsResponse *types.GetPaymentsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		paymentsResponse, err = s.Tasks.FetchPaymentsByCustomer(ctx, tx, page, limit, startDate, endDate, customerId)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch payments for customer %s: %v", customerId, err)
		return nil, err
	}

	return paymentsResponse, nil
}
func (s *PaymentService) CreateDownpaymentIntent(ctx context.Context, orderID string) (*types.PaymentIntentResponse, error) {
	var intent *types.PaymentIntentResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		order, err := s.Tasks.FetchOrderByID(ctx, tx, orderID)
		if err != nil {
			return err
		}

		if order.PaymentStatus != "pending_downpayment" {
			return errors.New("order not eligible for downpayment")
		}

		// Convert PHP to centavos
		amountInCents := int64(math.Round(float64(order.DownpaymentRequired) * 100))

		body := map[string]any{
			"data": map[string]any{
				"attributes": map[string]any{
					"amount":                 amountInCents,
					"currency":               "PHP",
					"capture_type":           "automatic",
					"payment_method_allowed": []string{"card", "gcash", "qrph"},
					"description":            "Handworks Cleaning Downpayment",
				},
			},
		}

		intent, err = s.PaymongoClient.CreatePaymentIntent(ctx, body)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(intent)
		if err != nil {
			return fmt.Errorf("failed to marshal payment intent response: %v", err)
		}

		var failedReason *string
		if intent.Data.Attributes.LastPaymentError != nil {
			msg := intent.Data.Attributes.LastPaymentError.Message
			failedReason = &msg
		}

		payment := &types.StorePayment{
			OrderID:         orderID,
			ClientKey:       intent.Data.Attributes.ClientKey,
			Type:            "DOWNPAYMENT",
			PaymentIntentID: &intent.Data.ID,
			Currency:        intent.Data.Attributes.Currency,
			Provider:        order.PaymentMethod,
			FailedReason:    failedReason,
			RawResponse:     raw,
			Amount:          order.DownpaymentRequired,
			Status:          intent.Data.Attributes.Status,
		}
		err = s.Tasks.StorePayment(ctx, tx, payment)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.Logger.Error("Failed to create downpayment intent for order %s: %v", orderID, err)
		return nil, err
	}

	return intent, nil
}
func (s *PaymentService) CreateFullPaymentIntent(ctx context.Context, orderID string) (*types.PaymentIntentResponse, error) {
	var intent *types.PaymentIntentResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		order, err := s.Tasks.FetchOrderByID(ctx, tx, orderID)
		if err != nil {
			return err
		}

		if order.PaymentStatus != "pending_fullpayment" {
			return errors.New("order not eligible for full payment")
		}

		// Convert PHP to centavos
		amountInCents := int64(math.Round(float64(order.RemainingBalance) * 100))

		body := map[string]any{
			"data": map[string]any{
				"attributes": map[string]any{
					"amount":                 amountInCents,
					"currency":               "PHP",
					"capture_type":           "automatic",
					"payment_method_allowed": []string{"card", "gcash", "qrph"},
					"description":            "Handworks Cleaning Full Payment",
				},
			},
		}

		intent, err = s.PaymongoClient.CreatePaymentIntent(ctx, body)
		if err != nil {
			return err
		}
		raw, err := json.Marshal(intent)
		if err != nil {
			return fmt.Errorf("failed to marshal payment intent response: %v", err)
		}

		var failedReason *string
		if intent.Data.Attributes.LastPaymentError != nil {
			msg := intent.Data.Attributes.LastPaymentError.Message
			failedReason = &msg
		}

		payment := &types.StorePayment{
			OrderID:         orderID,
			ClientKey:       intent.Data.Attributes.ClientKey,
			Type:            "FULLPAYMENT",
			PaymentIntentID: &intent.Data.ID,
			Currency:        intent.Data.Attributes.Currency,
			Provider:        order.PaymentMethod,
			FailedReason:    failedReason,
			RawResponse:     raw,
			Amount:          order.RemainingBalance,
			Status:          intent.Data.Attributes.Status,
		}
		err = s.Tasks.StorePayment(ctx, tx, payment)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.Logger.Error("Failed to create full payment intent for order %s: %v", orderID, err)
		return nil, err
	}

	return intent, nil
}

func (s *PaymentService) CashFullPayment(ctx context.Context, orderID string) error {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		order, err := s.Tasks.FetchOrderByID(ctx, tx, orderID)
		if err != nil {
			return err
		}

		if order.PaymentStatus != "pending_fullpayment" {
			return errors.New("order not eligible for full payment")
		}

		err = s.Tasks.UpdateOrderPaymentStatus(ctx, tx, "", "", "paid")
		if err != nil {
			return err
		}

		payment := &types.StorePayment{
			OrderID:  orderID,
			Type:     "FULLPAYMENT",
			Currency: "PHP",
			Provider: "CASH",
			Amount:   order.RemainingBalance,
			Status:   "paid",
		}
		err = s.Tasks.StorePayment(ctx, tx, payment)
		return err
	}); err != nil {
		s.Logger.Error("Failed to process cash full payment for order %s: %v", orderID, err)
		return err
	}
	return nil
}

func (s *PaymentService) CreateStaticQRPHCode(ctx context.Context, req types.CreateQRPHCodeRequest) (*types.QRPHCodeResponse, error) {
	kind := strings.TrimSpace(strings.ToLower(req.Kind))
	if kind == "" {
		kind = "instore"
	}

	if kind != "instore" {
		return nil, errors.New("kind must be instore")
	}

	body := map[string]any{
		"data": map[string]any{
			"attributes": map[string]any{
				"mobile_number": req.MobileNumber,
				"kind":          kind,
			},
		},
	}

	if req.Notes != nil {
		body["data"].(map[string]any)["attributes"].(map[string]any)["notes"] = *req.Notes
	}

	res, err := s.PaymongoClient.CreateQRPHCode(ctx, body)
	if err != nil {
		s.Logger.Error("Failed to create QRPH static code: %v", err)
		return nil, err
	}

	return res, nil
}

func (s *PaymentService) HandlePaymentPaid(ctx context.Context, data types.WebhookEventData) error {
	paymentIntentId := *data.Attributes.Data.Attributes.PaymentIntentID
	paymentId := data.Attributes.Data.ID
	status := data.Attributes.Data.Attributes.Status
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		if err := s.Tasks.UpdateOrderPaymentStatus(ctx, tx, paymentIntentId, paymentId, "pending_fullpayment"); err != nil {
			return err
		}
		if err := s.Tasks.UpdatePaymentStatus(ctx, tx, paymentId, paymentIntentId, status); err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.Logger.Error("Failed to update order payment status for payment intent %s: %v", paymentIntentId, err)
		return err
	}
	return nil
}
func (s *PaymentService) HandlePaymentFailed(ctx context.Context, data types.WebhookEventData) error {
	paymentIntentId := *data.Attributes.Data.Attributes.PaymentIntentID
	paymentId := data.Attributes.Data.ID
	failMessage := data.Attributes.Data.Attributes.FailedMessage
	status := data.Attributes.Data.Attributes.Status
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		if err := s.Tasks.UpdateOrderPaymentStatus(ctx, tx, paymentIntentId, paymentId, "failed"); err != nil {
			return err
		}
		if err := s.Tasks.UpdatePaymentStatusFailed(ctx, tx, paymentId, paymentIntentId, *failMessage, status); err != nil {
			return err
		}
		return nil
	}); err != nil {
		s.Logger.Error("Failed to update order payment status for payment intent %s: %v", paymentIntentId, err)
		return err
	}
	return nil
}
func (s *PaymentService) HasExistingDownpayment(ctx context.Context, orderID string) (*types.ExistingDownpaymentResponse, error) {
	var res *types.ExistingDownpaymentResponse
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		res, err = s.Tasks.CheckExistingDownpayment(ctx, tx, orderID)
		return err
	}); err != nil {
		s.Logger.Error("Failed to check existing downpayment for order %s: %v", orderID, err)
		return nil, err
	}
	return res, nil
}
