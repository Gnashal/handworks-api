package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"time"

	"github.com/jackc/pgx/v5"
)

type PaymentTasks struct{}

func CalculateGeneralCleaning(details *types.GeneralCleaningDetails) (float32, int32) {
	if details == nil {
		return 0.0, 0
	}
	sqm := details.SQM
	homeType := details.HomeType

	var price float32
	var hours int32

	switch {
	case homeType == "CONDO_ROOM" || (sqm > 0 && sqm <= 30):
		price = 2000.00
		hours = 2
	case homeType == "HOUSE" || (sqm > 30 && sqm <= 50):
		price = 2500.00
		hours = 4
	case sqm > 50 && sqm <= 100:
		price = 5000.00
		hours = 8
	default:
		price = float32(sqm * 50)
		calculatedHours := int32(sqm * 1)
		if calculatedHours < 8 {
			hours = 8
		} else {
			hours = calculatedHours
		}
	}

	return price, hours
}

func CalculateCarCleaning(details *types.CarCleaningDetails) (float32, int32) {
	if details == nil {
		return 0.0, 0
	}

	var total float32
	var totalHours int32

	for _, spec := range details.CleaningSpecs {
		price := types.CarPrices[spec.CarType]
		total += price * float32(spec.Quantity)

		// Add hours based on car type
		var carHours int32
		switch spec.CarType {
		case "VAN":
			carHours = 2
		default:
			carHours = 1
		}
		totalHours += carHours * int32(spec.Quantity)
	}

	if details.ChildSeats > 0 {
		total += float32(details.ChildSeats) * 250.00
		totalHours += int32(details.ChildSeats)
	}

	return total, totalHours
}

func CalculateCouchCleaning(details *types.CouchCleaningDetails) (float32, int32) {
	if details == nil {
		return 0.0, 0
	}

	var total float32
	var totalHours int32

	for _, spec := range details.CleaningSpecs {
		price := types.CouchPrices[spec.CouchType]
		total += price * float32(spec.Quantity)

		var couchHours int32
		switch spec.CouchType {
		case "SEATER_4_LTYPE_LARGE", "SEATER_5_LTYPE", "SEATER_6_LTYPE":
			couchHours = 2
		default:
			couchHours = 1
		}
		totalHours += couchHours * int32(spec.Quantity)
	}

	if details.BedPillows > 0 {
		total += float32(details.BedPillows) * 100.00
		pillowHours := float64(details.BedPillows) * 0.25
		totalHours += int32(pillowHours)
		if totalHours == 0 && details.BedPillows > 0 {
			totalHours = 1
		}
	}

	return total, totalHours
}

func CalculateMattressCleaning(details *types.MattressCleaningDetails) (float32, int32) {
	if details == nil {
		return 0.0, 0
	}

	var total float32
	var totalHours int32

	for _, spec := range details.CleaningSpecs {
		price := types.MattressPrices[spec.BedType]
		total += price * float32(spec.Quantity)

		var bedHours int32
		if spec.BedType == "KING_HEADBAND" || spec.BedType == "QUEEN_HEADBAND" {
			bedHours = 2
		} else {
			bedHours = 1
		}
		totalHours += bedHours * int32(spec.Quantity)
	}

	return total, totalHours
}

func CalculatePostConstructionCleaning(details *types.PostConstructionDetails) (float32, int32) {
	if details == nil {
		return 0.0, 0
	}

	price := float32(details.SQM * 50.00)

	var hours int32

	if details.SQM <= 50 && details.SQM > 0 {
		hours = 2
	} else if details.SQM > 50 && details.SQM <= 100 {
		hours = 4
	} else if details.SQM > 100 && details.SQM <= 200 {
		hours = 8
	}

	return price, hours
}

func (t *PaymentTasks) CalculatePriceByServiceType(service *types.ServicesRequest) (float32, int32) {
	if service == nil {
		return 0, 0
	}

	var calculatedPrice float32 = 0.00
	var calculatedHours int32 = 0

	switch service.ServiceType {
	case types.GeneralCleaning:
		calculatedPrice, calculatedHours = CalculateGeneralCleaning(service.Details.General)

	case types.CouchCleaning:
		calculatedPrice, calculatedHours = CalculateCouchCleaning(service.Details.Couch)

	case types.MattressCleaning:
		calculatedPrice, calculatedHours = CalculateMattressCleaning(service.Details.Mattress)

	case types.CarCleaning:
		calculatedPrice, calculatedHours = CalculateCarCleaning(service.Details.Car)

	case types.PostCleaning:
		calculatedPrice, calculatedHours = CalculatePostConstructionCleaning(service.Details.Post)

	default:
		// no default action
	}

	return calculatedPrice, calculatedHours
}

func (t *PaymentTasks) CalculateQuotePreview(c context.Context, in *types.QuoteRequest) (*types.Quote, error) {
	var dbQuote types.Quote
	var dbAddons []*types.QuoteAddon

	mainService := &types.ServicesRequest{
		ServiceType: in.Service.ServiceType,
		Details:     in.Service.Details,
	}

	mainServiceDetail, err := json.Marshal(in.Service)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal main service: %v", err)
	}

	subtotal, mainHours := t.CalculatePriceByServiceType(mainService)
	var addonTotal float32 = 0
	var addonTotalHours int32 = 0

	for _, addon := range in.Addons {
		addonService := &types.ServicesRequest{
			ServiceType: addon.ServiceDetail.ServiceType,
			Details:     addon.ServiceDetail.Details,
		}
		addonPrice, addonHours := t.CalculatePriceByServiceType(addonService)

		serviceDetail, err := json.Marshal(addon.ServiceDetail)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addon service: %v", err)
		}

		addonTotal += addonPrice
		addonTotalHours += addonHours

		dbAddon := &types.QuoteAddon{
			ServiceType:   string(addon.ServiceDetail.ServiceType),
			ServiceDetail: serviceDetail,
			ServiceHours:  addonHours,
			AddonPrice:    addonPrice,
			CreatedAt:     time.Now(),
		}
		dbAddons = append(dbAddons, dbAddon)
	}

	totalServiceHours := mainHours + addonTotalHours

	dbQuote = types.Quote{
		ID:                "",
		CustomerID:        in.CustomerID,
		MainService:       string(in.Service.ServiceType),
		MainServiceDetail: mainServiceDetail,
		MainServiceHours:  mainHours,
		Subtotal:          subtotal,
		AddonTotal:        addonTotal,
		TotalServiceHours: totalServiceHours,
		TotalPrice:        subtotal + addonTotal,
		IsValid:           false,
		CreatedAt:         time.Now(),
		Addons:            dbAddons,
	}

	return &dbQuote, nil
}

func (t *PaymentTasks) MapAddonstoAddonBreakdown(addons *[]*types.QuoteAddon) []types.AddOnBreakdown {
	var breakdowns []types.AddOnBreakdown
	if addons != nil && len(*addons) > 0 {
		for _, addon := range *addons {
			breakdown := types.AddOnBreakdown{
				AddonID:       addon.ID,
				ServiceType:   addon.ServiceType,
				ServiceDetail: addon.ServiceDetail,
				ServiceHours:  addon.ServiceHours,
				Price:         float64(addon.AddonPrice),
			}
			breakdowns = append(breakdowns, breakdown)
		}
		return breakdowns
	}
	return []types.AddOnBreakdown{}
}

func (p *PaymentTasks) CreateQuote(c context.Context, tx pgx.Tx, in *types.QuoteRequest) (*types.Quote, error) {
	var dbQuote types.Quote
	var dbAddons []*types.QuoteAddon
	var mainServiceDetail []byte

	mainService := &types.ServicesRequest{
		ServiceType: in.Service.ServiceType,
		Details:     in.Service.Details,
	}

	mainServiceDetail, marshalErr := json.Marshal(in.Service)
	if marshalErr != nil {
		return nil, fmt.Errorf("failed to marshal main service: %v", marshalErr)
	}

	subtotal, mainHours := p.CalculatePriceByServiceType(mainService)
	var addonTotal float32 = 0
	var addonTotalHours int32 = 0

	for _, addon := range in.Addons {
		addonService := &types.ServicesRequest{
			ServiceType: addon.ServiceDetail.ServiceType,
			Details:     addon.ServiceDetail.Details,
		}
		addonPrice, addonHours := p.CalculatePriceByServiceType(addonService)

		serviceDetail, err := json.Marshal(addon.ServiceDetail)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addon service: %v", err)
		}

		addonTotal += addonPrice
		addonTotalHours += addonHours

		dbAddon := &types.QuoteAddon{
			ServiceType:   string(addon.ServiceDetail.ServiceType),
			ServiceDetail: serviceDetail,
			ServiceHours:  addonHours,
			AddonPrice:    addonPrice,
			CreatedAt:     time.Now(),
		}
		dbAddons = append(dbAddons, dbAddon)
	}

	totalPrice := subtotal + addonTotal
	totalServiceHours := mainHours + addonTotalHours

	err := tx.QueryRow(c, `
		INSERT INTO payment.quotes (
			customer_id,
			main_service_type,
			main_service_detail,
			main_service_hours,
			subtotal,
			addon_total,
			total_service_hours,
			total_price,
			is_valid
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, TRUE)
		RETURNING id, customer_id, main_service_type, main_service_detail,
		          main_service_hours, subtotal, addon_total, total_service_hours,
		          total_price, is_valid, created_at, updated_at
	`,
		in.CustomerID,
		in.Service.ServiceType,
		mainServiceDetail,
		mainHours,
		subtotal,
		addonTotal,
		totalServiceHours,
		totalPrice,
	).Scan(
		&dbQuote.ID,
		&dbQuote.CustomerID,
		&dbQuote.MainService,
		&dbQuote.MainServiceDetail,
		&dbQuote.MainServiceHours,
		&dbQuote.Subtotal,
		&dbQuote.AddonTotal,
		&dbQuote.TotalServiceHours,
		&dbQuote.TotalPrice,
		&dbQuote.IsValid,
		&dbQuote.CreatedAt,
		&dbQuote.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert quote: %v", err)
	}

	for _, addon := range dbAddons {
		err := tx.QueryRow(c, `
			INSERT INTO payment.quote_addons (quote_id, service_type, service_detail, service_hours, addon_price)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at
		`,
			dbQuote.ID,
			addon.ServiceType,
			addon.ServiceDetail,
			addon.ServiceHours,
			addon.AddonPrice,
		).Scan(&addon.ID, &addon.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to insert addon: %v", err)
		}

		addon.QuoteID = dbQuote.ID
	}

	dbQuote.Addons = dbAddons
	return &dbQuote, nil
}

func (t *PaymentTasks) VerifyQuoteAndFetchPrices(ctx context.Context, tx pgx.Tx, quoteId string) (*types.CleaningPrices, error) {
	var prices types.CleaningPrices

	var dbQuote types.Quote
	err := tx.QueryRow(ctx, `
		SELECT total_price, is_valid
		FROM payment.quotes
		WHERE id = $1
	`, quoteId).Scan(
		&dbQuote.TotalPrice,
		&dbQuote.IsValid,
	)
	if err != nil {
		return &prices, fmt.Errorf("fetch main quote: %w", err)
	}
	if !dbQuote.IsValid {
		return &types.CleaningPrices{}, fmt.Errorf("quote is no longer valid")
	}

	rows, err := tx.Query(ctx, `
		SELECT service_type, addon_price
		FROM payment.quote_addons
		WHERE quote_id = $1
	`, quoteId)
	if err != nil {
		return &prices, fmt.Errorf("fetch addons: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var addon types.QuoteAddon
		if err := rows.Scan(
			&addon.ServiceType,
			&addon.AddonPrice,
		); err != nil {
			return &prices, fmt.Errorf("scan addon: %w", err)
		}
		dbQuote.Addons = append(dbQuote.Addons, &addon)
	}

	for _, a := range dbQuote.Addons {
		prices.AddonPrices = append(prices.AddonPrices, types.AddonCleaningPrice{
			AddonName:  a.ServiceType,
			AddonPrice: a.AddonPrice,
		})
	}
	prices.MainServicePrice = dbQuote.TotalPrice
	return &prices, nil
}

func (t *PaymentTasks) FetchAllQuotesByCustomer(
	ctx context.Context,
	tx pgx.Tx,
	customerId, startDate, endDate string,
	page, limit int,
	logger *utils.Logger,
) (*types.FetchAllQuotesResponse, error) {

	var rawJSON []byte
	err := tx.QueryRow(ctx,
		`SELECT payment.get_quotes_by_customer($1, $2, $3, $4, $5)`,
		customerId, startDate, endDate, page, limit,
	).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_quotes_by_customer: %w", err)
	}

	var response types.FetchAllQuotesResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal quotes JSON: %v", err)
		return nil, fmt.Errorf("unmarshal quotes: %w", err)
	}

	return &response, nil
}

func (t *PaymentTasks) FetchAllQuotes(
	ctx context.Context,
	tx pgx.Tx,
	startDate, endDate string,
	page, limit int,
	logger *utils.Logger,
) (*types.FetchAllQuotesResponse, error) {

	var rawJSON []byte
	err := tx.QueryRow(ctx,
		`SELECT payment.get_quotes($1, $2, $3, $4)`,
		startDate, endDate, page, limit,
	).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_quotes: %w", err)
	}

	var response types.FetchAllQuotesResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal quotes JSON: %v", err)
		return nil, fmt.Errorf("unmarshal quotes: %w", err)
	}

	return &response, nil
}

func (t *PaymentTasks) FetchQuoteByIDbyCustomer(ctx context.Context, tx pgx.Tx, quoteId, customerId string) (*types.QuoteResponse, error) {
	var quoteResponse types.QuoteResponse
	var responseJSON []byte

	err := tx.QueryRow(ctx, `
		SELECT payment.get_customer_quote($1::uuid, $2::uuid)
	`, quoteId, customerId).Scan(&responseJSON)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("quote not found for this customer")
		}
		return nil, fmt.Errorf("failed to fetch quote: %w", err)
	}

	if err := json.Unmarshal(responseJSON, &quoteResponse); err != nil {
		return nil, fmt.Errorf("failed to parse quote response: %w", err)
	}

	var validAddons []types.AddOnBreakdown
	for _, addon := range quoteResponse.Addons {
		if addon.AddonID != "" && addon.Price > 0 {
			validAddons = append(validAddons, addon)
		}
	}
	quoteResponse.Addons = validAddons

	var filteredAddonTotal float32 = 0
	for _, addon := range validAddons {
		filteredAddonTotal += float32(addon.Price)
	}
	quoteResponse.AddonTotal = filteredAddonTotal

	return &quoteResponse, nil
}

func (t *PaymentTasks) CreateOrder(
	ctx context.Context,
	tx pgx.Tx,
	req types.CreateOrderRequest,
) (string, error) {
	// Calculate downpayment
	downpayment := req.TotalAmount * 0.20
	remaining := req.TotalAmount - downpayment

	const query = `
		INSERT INTO payment.orders (
			order_number,
			customer_id,
			quote_id,
			currency,
			subtotal,
			addon_total,
			total_amount,
			downpayment_required,
			remaining_balance,
			payment_status,
			created_at,
			updated_at
		)
		VALUES (
			$1, $2, $3, $4,
			$5, $6, $7,
			$8, $9,
			$10,
			NOW(), NOW()
		)
		RETURNING id;
	`

	var orderID string

	err := tx.QueryRow(ctx, query,
		utils.GenerateOrderNumber(req.QuoteID, time.Now()),
		req.CustomerID,
		req.QuoteID,
		"PHP",
		req.Subtotal,
		req.AddonTotal,
		req.TotalAmount,
		downpayment,
		remaining,
		"pending_downpayment",
	).Scan(&orderID)

	if err != nil {
		return "", fmt.Errorf("failed to create order: %w", err)
	}

	return orderID, nil
}

func (t *PaymentTasks) FetchOrderByID(ctx context.Context, tx pgx.Tx, orderId string) (*types.Order, error) {
	var order types.Order

	err := tx.QueryRow(ctx, `SELECT * FROM payment.orders WHERE id = $1`, orderId).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.QuoteID,
		&order.Currency,
		&order.Subtotal,
		&order.AddonTotal,
		&order.TotalAmount,
		&order.DownpaymentRequired,
		&order.RemainingBalance,
		&order.PaymentStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	return &order, nil
}
func (t *PaymentTasks) FetchOrders(ctx context.Context, tx pgx.Tx, page, limit int, startDate, endDate string, logger *utils.Logger) (*types.GetOrdersResponse, error) {
	var response types.GetOrdersResponse
	var orders []types.Order
	rows, err := tx.Query(ctx,
		`SELECT payment.get_orders($1, $2, $3, $4)`,
		startDate, endDate, page, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o types.Order
		if err := rows.Scan(
			&o.ID,
			&o.OrderNumber,
			&o.CustomerID,
			&o.QuoteID,
			&o.Currency,
			&o.Subtotal,
			&o.AddonTotal,
			&o.TotalAmount,
			&o.DownpaymentRequired,
			&o.RemainingBalance,
			&o.PaymentStatus,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	response.Orders = orders
	response.TotalOrders = len(orders)
	response.OrdersRequested = limit
	return &response, nil
}
func (t *PaymentTasks) FetchOrdersByCustomer(ctx context.Context, tx pgx.Tx, page, limit int, startDate, endDate, customerId string, logger *utils.Logger) (*types.GetOrdersResponse, error) {
	var response types.GetOrdersResponse
	var orders []types.Order
	rows, err := tx.Query(ctx,
		`SELECT payment.get_orders_by_customer($1, $2, $3, $4, $5)`,
		startDate, endDate, page, limit, customerId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var o types.Order
		if err := rows.Scan(
			&o.ID,
			&o.OrderNumber,
			&o.CustomerID,
			&o.QuoteID,
			&o.Currency,
			&o.Subtotal,
			&o.AddonTotal,
			&o.TotalAmount,
			&o.DownpaymentRequired,
			&o.RemainingBalance,
			&o.PaymentStatus,
			&o.CreatedAt,
			&o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	response.Orders = orders
	response.TotalOrders = len(orders)
	response.OrdersRequested = limit
	return &response, nil
}
func (t *PaymentTasks) FetchPaymentsByOrderID(ctx context.Context, tx pgx.Tx, page, limit int, startDate, endDate, orderId string) (*types.GetPaymentsResponse, error) {
	var response types.GetPaymentsResponse
	var payments []types.Payment
	rows, err := tx.Query(ctx,
		`SELECT payment.get_payments_by_order($1, $2, $3, $4, $5)`,
		startDate, endDate, page, limit, orderId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p types.Payment
		if err := rows.Scan(
			&p.ID,
			&p.OrderID,
			&p.Amount,
			&p.Currency,
			&p.FailedReason,
			&p.PaidAt,
			&p.PaymentID,
			&p.PaymentIntentID,
			&p.Status,
			&p.Type,
			&p.Provider,
			&p.RawResponse,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	response.Payments = payments
	response.TotalPayments = len(payments)
	response.PaymentsRequested = limit
	return &response, nil
}
func (t *PaymentTasks) FetchPaymentsByCustomer(ctx context.Context, tx pgx.Tx, page, limit int, startDate, endDate, customerId string) (*types.GetPaymentsResponse, error) {
	var response types.GetPaymentsResponse
	var payments []types.Payment
	rows, err := tx.Query(ctx,
		`SELECT payment.get_payments_by_customer($1, $2, $3, $4, $5)`,
		startDate, endDate, page, limit, customerId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var p types.Payment
		if err := rows.Scan(
			&p.ID,
			&p.OrderID,
			&p.Amount,
			&p.Currency,
			&p.FailedReason,
			&p.PaidAt,
			&p.PaymentID,
			&p.PaymentIntentID,
			&p.Status,
			&p.Type,
			&p.Provider,
			&p.RawResponse,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	response.Payments = payments
	response.TotalPayments = len(payments)
	response.PaymentsRequested = limit
	return &response, nil
}
