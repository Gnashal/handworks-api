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

	var price float32
	var hours int32

	switch {
	case sqm > 0 && sqm <= 30:
		price = 2000.00
		hours = 2
	case sqm > 30 && sqm <= 50:
		price = 2500.00
		hours = 4
	case sqm > 50 && sqm <= 100:
		price = 5000.00
		hours = 8
	default:
		price = 0
		hours = 0
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

		// Add hours based on car type from price list
		var carHours int32
		switch spec.CarType {
		case "SEDAN_5_SEATER":
			carHours = 2
		case "MPV_7_SEATER":
			carHours = 2 // 2 hours 30 min = 2.5 hours, but we'll use 2 and add minutes separately
		case "SUV_7_8_SEATER":
			carHours = 2 // 2 hours 30 min
		case "PICKUP_5_SEATER":
			carHours = 2
		case "FAMILY_VAN_10_SEATER":
			carHours = 4
		case "SPORTS_CAR_1_2_SEATER":
			carHours = 1 // 1 hour 30 min
		default:
			carHours = 1
		}
		totalHours += carHours * int32(spec.Quantity)
	}

	if details.ChildSeats > 0 {
		total += float32(details.ChildSeats) * 250.00
		// Child car seat takes 30 minutes = 0.5 hours
		childSeatHours := float32(details.ChildSeats) * 0.5
		totalHours += int32(childSeatHours)
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

		// Update couch hours based on price list
		var couchHours float32
		switch spec.CouchType {
		case "SEATER_3_LTYPE_LARGE":
			couchHours = 3 // 2-4 hours, using average 3
		case "SEATER_4_LTYPE_SMALL":
			couchHours = 2
		case "SEATER_4_LTYPE_LARGE":
			couchHours = 3 // 2-4 hours
		case "SEATER_5_LTYPE":
			couchHours = 2
		case "SEATER_6_LTYPE":
			couchHours = 4
		case "OTTOMAN":
			couchHours = 0.5
		case "LAZY_BOY":
			couchHours = 1
		case "CHAIR":
			couchHours = 0.5
		default:
			couchHours = 1
		}
		totalHours += int32(couchHours * float32(spec.Quantity))
	}

	if details.BedPillows > 0 {
		total += float32(details.BedPillows) * 100.00
		// Bed pillow takes 30 minutes = 0.5 hours
		pillowHours := float32(details.BedPillows) * 0.5
		totalHours += int32(pillowHours)
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

		// Update mattress hours based on price list
		var bedHours float32
		switch spec.BedType {
		case "KING":
			bedHours = 2.5 // 2 hours 30 min
		case "KING_HEADBAND":
			bedHours = 2.5
		case "QUEEN":
			bedHours = 2
		case "QUEEN_HEADBAND":
			bedHours = 2
		case "SINGLE":
			bedHours = 1.5 // 1 hour 30 min
		default:
			bedHours = 1
		}
		totalHours += int32(bedHours * float32(spec.Quantity))
	}

	return total, totalHours
}

func CalculatePostConstructionCleaning(details *types.PostConstructionDetails) (float32, int32) {
	if details == nil {
		return 0.0, 0
	}

	price := float32(details.SQM) * 50.00

	var hours int32
	sqm := details.SQM

	if sqm <= 30 {
		hours = 2
	} else if sqm <= 50 {
		hours = 4
	} else if sqm <= 100 {
		hours = 8
	} else if sqm <= 200 {
		hours = 12
	} else {
		// For very large areas, approximately 8 hours per 100 SQM
		hours = int32(sqm / 13)
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

	// Check if total service hours exceed the limit (11 hours)
	if totalServiceHours > 11 {
		return nil, fmt.Errorf("total service hours (%d) exceed maximum allowed limit of 11 hours. Please reduce the scope of work", totalServiceHours)
	}

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

	// Check if total service hours exceed the limit (11 hours)
	if totalServiceHours > 11 {
		return nil, fmt.Errorf("total service hours (%d) exceed maximum allowed limit of 11 hours. Please reduce the scope of work", totalServiceHours)
	}

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
