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

func CalculateGeneralCleaning(details *types.GeneralCleaningDetails) float32 {
	if details == nil {
		return 0.0
	}
	sqm := details.SQM
	homeType := details.HomeType

	switch {
	case homeType == "CONDO_ROOM" || (sqm > 0 && sqm <= 30):
		return 2000.00
	case homeType == "HOUSE" || (sqm > 30 && sqm <= 50):
		return 2500.00
	case sqm > 50 && sqm <= 100:
		return 5000.00
	default:
		return float32(sqm * 50)
	}
}

func CalculateCarCleaning(details *types.CarCleaningDetails) float32 {
	if details == nil {
		return 0.0
	}

	var total float32
	for _, spec := range details.CleaningSpecs {
		price := types.CarPrices[spec.CarType]
		total += price * float32(spec.Quantity)
	}

	if details.ChildSeats > 0 {
		total += float32(details.ChildSeats) * 250.00
	}

	return total
}

func CalculateCouchCleaning(details *types.CouchCleaningDetails) float32 {
	if details == nil {
		return 0.0
	}

	var total float32
	for _, spec := range details.CleaningSpecs {
		price := types.CouchPrices[spec.CouchType]
		total += price * float32(spec.Quantity)
	}

	if details.BedPillows > 0 {
		total += float32(details.BedPillows) * 100.00
	}

	return total
}

func CalculateMattressCleaning(details *types.MattressCleaningDetails) float32 {
	if details == nil {
		return 0.0
	}

	var total float32
	for _, spec := range details.CleaningSpecs {
		price := types.MattressPrices[spec.BedType]
		total += price * float32(spec.Quantity)
	}
	return total
}
func CalculatePostConstructionCleaning(details *types.PostConstructionDetails) float32 {
	if details == nil {
		return 0.0
	}
	return float32(details.SQM * 50.00)
}

func (t *PaymentTasks) CalculatePriceByServiceType(service *types.ServicesRequest) float32 {
	if service == nil {
		return 0
	}

	var calculatedPrice float32 = 0.00

	switch service.ServiceType {
	case types.GeneralCleaning:
		calculatedPrice = CalculateGeneralCleaning(service.Details.General)

	case types.CouchCleaning:
		calculatedPrice = CalculateCouchCleaning(service.Details.Couch)

	case types.MattressCleaning:
		calculatedPrice = CalculateMattressCleaning(service.Details.Mattress)

	case types.CarCleaning:
		calculatedPrice = CalculateCarCleaning(service.Details.Car)

	case types.PostCleaning:
		calculatedPrice = CalculatePostConstructionCleaning(service.Details.Post)

	default:
		// no default action
	}

	return calculatedPrice
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

	subtotal := t.CalculatePriceByServiceType(mainService)
	var addonTotal float32 = 0

	for _, addon := range in.Addons {
		addonService := &types.ServicesRequest{
			ServiceType: addon.ServiceDetail.ServiceType,
			Details:     addon.ServiceDetail.Details,
		}
		addonPrice := t.CalculatePriceByServiceType(addonService)

		serviceDetail, err := json.Marshal(addon.ServiceDetail)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addon service: %v", err)
		}

		addonTotal += addonPrice
		dbAddon := &types.QuoteAddon{
			ServiceType:   string(addon.ServiceDetail.ServiceType),
			ServiceDetail: serviceDetail,
			AddonPrice:    addonPrice,
			CreatedAt:     time.Now(),
		}
		dbAddons = append(dbAddons, dbAddon)
	}

	dbQuote = types.Quote{
		ID:                "",
		CustomerID:        in.CustomerID,
		MainService:       string(in.Service.ServiceType),
		MainServiceDetail: mainServiceDetail,
		Subtotal:          subtotal,
		AddonTotal:        addonTotal,
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
				Price:         float64(addon.AddonPrice),
			}
			breakdowns = append(breakdowns, breakdown)
		}
		return breakdowns
	} else {
		return []types.AddOnBreakdown{}
	}
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

	// Calc subtotal for main service
	subtotal := p.CalculatePriceByServiceType(mainService)
	var addonTotal float32 = 0

	// Calculate each addon price
	for _, addon := range in.Addons {
		// i genuinely dunno why addon.ServiceDetail wont work lmao
		addonService := &types.ServicesRequest{
			ServiceType: addon.ServiceDetail.ServiceType,
			Details:     addon.ServiceDetail.Details,
		}
		addonPrice := p.CalculatePriceByServiceType(addonService)

		// serialize the full addon service detail
		serviceDetail, err := json.Marshal(addon.ServiceDetail)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addon service: %v", err)
		}

		addonTotal += addonPrice

		dbAddon := &types.QuoteAddon{
			ServiceType:   string(addon.ServiceDetail.ServiceType),
			ServiceDetail: serviceDetail,
			AddonPrice:    addonPrice,
			CreatedAt:     time.Now(),
		}
		dbAddons = append(dbAddons, dbAddon)
	}

	totalPrice := subtotal + addonTotal

	// Insert into quote table
	err := tx.QueryRow(c, `
		INSERT INTO payment.quotes (
			customer_id,
			main_service_type,
			main_service_detail,
			subtotal,
			addon_total,
			total_price,
			is_valid
		)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)
		RETURNING id, customer_id, main_service_type, main_service_detail,
		          subtotal, addon_total, total_price, is_valid, created_at, updated_at
	`,
		in.CustomerID,
		in.Service.ServiceType,
		mainServiceDetail,
		subtotal,
		addonTotal,
		totalPrice,
	).Scan(
		&dbQuote.ID,
		&dbQuote.CustomerID,
		&dbQuote.MainService,
		&dbQuote.MainServiceDetail,
		&dbQuote.Subtotal,
		&dbQuote.AddonTotal,
		&dbQuote.TotalPrice,
		&dbQuote.IsValid,
		&dbQuote.CreatedAt,
		&dbQuote.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert quote: %v", err)
	}
	// insert addons
	for _, addon := range dbAddons {
		err := tx.QueryRow(c, `
			INSERT INTO payment.quote_addons (quote_id, service_type, service_detail, addon_price)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at
		`,
			dbQuote.ID,
			addon.ServiceType,
			addon.ServiceDetail,
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
		return &types.CleaningPrices{}, fmt.Errorf("quote is no longer valied")
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
func (t *PaymentTasks) FetchAllQuotes(
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

func (t *PaymentTasks) FetchQuoteByIDbyCustomer(ctx context.Context, tx pgx.Tx, quoteId, customerId string) (*types.QuoteResponse, error) {
	if quoteId == "" {
		return nil, fmt.Errorf("quoteId is required")
	}
	if customerId == "" {
		return nil, fmt.Errorf("customerId is required")
	}

	var quoteResponse types.QuoteResponse
	var responseJSON []byte

	// Call the stored procedure
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
