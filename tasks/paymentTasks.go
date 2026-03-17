package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"log"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type PaymentTasks struct{}

// Maximum daily hours limit
const MaxDailyHours = 11

func CalculateGeneralCleaning(details *types.GeneralCleaningDetails) (float32, int32, error) {
	if details == nil {
		return 0.0, 0, fmt.Errorf("general cleaning details cannot be nil")
	}

	// Validate SQM
	if details.SQM <= 0 {
		return 0.0, 0, fmt.Errorf("invalid square meters: %d, must be greater than 0", details.SQM)
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
		// For areas above 100 SQM, return error to encourage splitting
		return 0.0, 0, fmt.Errorf("areas above 100 SQM require %d hours which exceeds our daily limit of %d hours. Please divide your cleaning into multiple bookings (e.g., book different floors/areas on separate days)",
			calculateHoursForLargeArea(sqm), MaxDailyHours)
	}

	// Individual service validation
	if hours > MaxDailyHours {
		return price, hours, fmt.Errorf("this general cleaning requires %d hours which exceeds our daily limit of %d hours. Please divide your cleaning into multiple bookings",
			hours, MaxDailyHours)
	}

	return price, hours, nil
}

// Helper function to calculate hours for large areas
func calculateHoursForLargeArea(sqm int32) int32 {
	if sqm <= 100 {
		return 8
	}
	additionalSQM := sqm - 100
	additionalHours := (additionalSQM + 12) / 13
	hours := 8 + int32(additionalHours)
	if hours > 24 {
		hours = 24
	}
	return hours
}

func CalculateCarCleaning(details *types.CarCleaningDetails) (float32, int32, error) {
	if details == nil {
		return 0.0, 0, fmt.Errorf("car cleaning details cannot be nil")
	}

	// Validate cleaning specs
	if len(details.CleaningSpecs) == 0 {
		return 0.0, 0, fmt.Errorf("at least one car cleaning specification is required")
	}

	var total float32
	var totalHours float32

	for _, spec := range details.CleaningSpecs {
		// Validate quantity
		if spec.Quantity <= 0 {
			return 0.0, 0, fmt.Errorf("invalid quantity %d for car type %s", spec.Quantity, spec.CarType)
		}

		price, ok := types.CarPrices[spec.CarType]
		if !ok {
			return 0.0, 0, fmt.Errorf("unknown car type: %s", spec.CarType)
		}
		total += price * float32(spec.Quantity)

		var carHours float32
		switch spec.CarType {
		case "SEDAN_5_SEATER":
			carHours = 2.0
		case "MPV_7_SEATER":
			carHours = 2.5
		case "SUV_7_8_SEATER":
			carHours = 2.5
		case "PICKUP_5_SEATER":
			carHours = 2.0
		case "FAMILY_VAN_10_SEATER":
			carHours = 4.0
		case "SPORTS_CAR_1_2_SEATER":
			carHours = 1.5
		default:
			return 0.0, 0, fmt.Errorf("unhandled car type: %s", spec.CarType)
		}
		totalHours += carHours * float32(spec.Quantity)
	}

	if details.ChildSeats > 0 {
		if details.ChildSeats > 10 {
			return 0.0, 0, fmt.Errorf("child seats quantity %d exceeds maximum limit of 10", details.ChildSeats)
		}
		total += float32(details.ChildSeats) * 250.00
		totalHours += float32(details.ChildSeats) * 0.5
	}

	finalHours := int32(totalHours*2+0.5) / 2

	// Individual service validation
	if finalHours > MaxDailyHours {
		return total, finalHours, fmt.Errorf("this car cleaning requires %d hours which exceeds our daily limit of %d hours. Please divide your car cleaning into multiple bookings (e.g., clean different vehicles on separate days)",
			finalHours, MaxDailyHours)
	}

	return total, finalHours, nil
}

func CalculateCouchCleaning(details *types.CouchCleaningDetails) (float32, int32, error) {
	if details == nil {
		return 0.0, 0, fmt.Errorf("couch cleaning details cannot be nil")
	}

	if len(details.CleaningSpecs) == 0 {
		return 0.0, 0, fmt.Errorf("at least one couch cleaning specification is required")
	}

	var total float32
	var totalHours float32

	for _, spec := range details.CleaningSpecs {
		if spec.WidthCM <= 0 || spec.DepthCM <= 0 || spec.HeightCM <= 0 {
			return 0.0, 0, fmt.Errorf("invalid dimensions for couch type %s", spec.CouchType)
		}

		if spec.Quantity <= 0 {
			return 0.0, 0, fmt.Errorf("invalid quantity %d for couch type %s", spec.Quantity, spec.CouchType)
		}

		price, ok := types.CouchPrices[spec.CouchType]
		if !ok {
			return 0.0, 0, fmt.Errorf("unknown couch type: %s", spec.CouchType)
		}
		total += price * float32(spec.Quantity)

		var couchHours float32
		switch spec.CouchType {
		case "SEATER_3_LTYPE_LARGE":
			couchHours = 3.0
		case "SEATER_4_LTYPE_SMALL":
			couchHours = 2.0
		case "SEATER_4_LTYPE_LARGE":
			couchHours = 3.0
		case "SEATER_5_LTYPE":
			couchHours = 2.0
		case "SEATER_6_LTYPE":
			couchHours = 4.0
		case "OTTOMAN":
			couchHours = 0.5
		case "LAZY_BOY":
			couchHours = 1.0
		case "CHAIR":
			couchHours = 0.5
		default:
			return 0.0, 0, fmt.Errorf("unhandled couch type: %s", spec.CouchType)
		}
		totalHours += couchHours * float32(spec.Quantity)
	}

	if details.BedPillows > 0 {
		if details.BedPillows > 20 {
			return 0.0, 0, fmt.Errorf("bed pillows quantity %d exceeds maximum limit of 20", details.BedPillows)
		}
		total += float32(details.BedPillows) * 100.00
		totalHours += float32(details.BedPillows) * 0.5
	}

	finalHours := int32(totalHours*2+0.5) / 2

	if finalHours > MaxDailyHours {
		return total, finalHours, fmt.Errorf("this couch cleaning requires %d hours which exceeds our daily limit of %d hours. Please divide your couch cleaning into multiple bookings (e.g., clean different rooms on separate days)",
			finalHours, MaxDailyHours)
	}

	return total, finalHours, nil
}

func CalculateMattressCleaning(details *types.MattressCleaningDetails) (float32, int32, error) {
	if details == nil {
		return 0.0, 0, fmt.Errorf("mattress cleaning details cannot be nil")
	}

	if len(details.CleaningSpecs) == 0 {
		return 0.0, 0, fmt.Errorf("at least one mattress cleaning specification is required")
	}

	var total float32
	var totalHours float32

	for _, spec := range details.CleaningSpecs {
		if spec.WidthCM <= 0 || spec.DepthCM <= 0 || spec.HeightCM <= 0 {
			return 0.0, 0, fmt.Errorf("invalid dimensions for bed type %s", spec.BedType)
		}

		if spec.Quantity <= 0 {
			return 0.0, 0, fmt.Errorf("invalid quantity %d for bed type %s", spec.Quantity, spec.BedType)
		}

		price, ok := types.MattressPrices[spec.BedType]
		if !ok {
			return 0.0, 0, fmt.Errorf("unknown bed type: %s", spec.BedType)
		}
		total += price * float32(spec.Quantity)

		var bedHours float32
		switch spec.BedType {
		case "KING", "KING_HEADBAND":
			bedHours = 2.5
		case "QUEEN", "QUEEN_HEADBAND":
			bedHours = 2.0
		case "SINGLE":
			bedHours = 1.5
		default:
			return 0.0, 0, fmt.Errorf("unhandled bed type: %s", spec.BedType)
		}
		totalHours += bedHours * float32(spec.Quantity)
	}

	finalHours := int32(totalHours*2+0.5) / 2

	if finalHours > MaxDailyHours {
		return total, finalHours, fmt.Errorf("this mattress cleaning requires %d hours which exceeds our daily limit of %d hours. Please divide your mattress cleaning into multiple bookings (e.g., clean different mattresses on separate days)",
			finalHours, MaxDailyHours)
	}

	return total, finalHours, nil
}

func CalculatePostConstructionCleaning(details *types.PostConstructionDetails) (float32, int32, error) {
	if details == nil {
		return 0.0, 0, fmt.Errorf("post construction cleaning details cannot be nil")
	}

	if details.SQM <= 0 {
		return 0.0, 0, fmt.Errorf("invalid square meters: %d, must be greater than 0", details.SQM)
	}

	price := float32(details.SQM) * 50.00
	var hours int32
	sqm := details.SQM

	switch {
	case sqm <= 30:
		hours = 2
	case sqm <= 50:
		hours = 4
	case sqm <= 100:
		hours = 8
	case sqm <= 200:
		hours = 12
	case sqm <= 300:
		hours = 16
	case sqm <= 400:
		hours = 20
	default:
		hours = 24 + int32((sqm-400)/50)
		if hours > 48 {
			hours = 48
		}
	}

	if hours > MaxDailyHours {
		daysNeeded := (hours + int32(MaxDailyHours) - 1) / int32(MaxDailyHours)
		return price, hours, fmt.Errorf("this post-construction cleaning requires %d hours, which exceeds our daily limit of %d hours. Please divide into %d separate day bookings (e.g., book %d hours on day 1 and %d hours on day 2)",
			hours, MaxDailyHours, daysNeeded, min(int32(MaxDailyHours), hours), hours-int32(MaxDailyHours))
	}

	return price, hours, nil
}

// Updated CalculatePriceByServiceType to return errors
func (t *PaymentTasks) CalculatePriceByServiceType(service *types.ServicesRequest) (float32, int32, error) {
	if service == nil {
		return 0, 0, fmt.Errorf("service request cannot be nil")
	}

	var calculatedPrice float32
	var calculatedHours int32
	var err error

	switch service.ServiceType {
	case types.GeneralCleaning:
		calculatedPrice, calculatedHours, err = CalculateGeneralCleaning(service.Details.General)
	case types.CouchCleaning:
		calculatedPrice, calculatedHours, err = CalculateCouchCleaning(service.Details.Couch)
	case types.MattressCleaning:
		calculatedPrice, calculatedHours, err = CalculateMattressCleaning(service.Details.Mattress)
	case types.CarCleaning:
		calculatedPrice, calculatedHours, err = CalculateCarCleaning(service.Details.Car)
	case types.PostCleaning:
		calculatedPrice, calculatedHours, err = CalculatePostConstructionCleaning(service.Details.Post)
	default:
		return 0, 0, fmt.Errorf("unsupported service type: %s", service.ServiceType)
	}

	if err != nil {
		return 0, 0, err
	}

	return calculatedPrice, calculatedHours, nil
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

	// Validate main service first
	subtotal, mainHours, err := t.CalculatePriceByServiceType(mainService)

	if err != nil {
		return nil, fmt.Errorf("main service validation failed: %v", err)
	}

	var addonTotal float32 = 0
	var addonTotalHours int32 = 0
	var validationErrors []string

	for i, addon := range in.Addons {
		addonService := &types.ServicesRequest{
			ServiceType: addon.ServiceDetail.ServiceType,
			Details:     addon.ServiceDetail.Details,
		}

		addonPrice, addonHours, err := t.CalculatePriceByServiceType(addonService)
		if err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Addon %d (%s): %v", i+1, addon.ServiceDetail.ServiceType, err))
			continue
		}

		if mainHours+addonTotalHours+addonHours > MaxDailyHours {
			validationErrors = append(validationErrors,
				fmt.Sprintf("Addon %d (%s) would exceed daily limit. Current total: %d hours, Addon requires: %d hours, Daily limit: %d hours. Please remove some items or create separate bookings.",
					i+1, addon.ServiceDetail.ServiceType, mainHours+addonTotalHours, addonHours, MaxDailyHours))
			continue
		}

		serviceDetail, err := json.Marshal(addon.ServiceDetail)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal addon service: %v", err)
		}
		addonTotalHours += addonHours
		addonTotal += addonPrice

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

	log.Printf("DEBUG CalculateQuotePreview - totalServiceHours: %d", totalServiceHours)

	if len(validationErrors) > 0 {
		var sb strings.Builder
		sb.WriteString("The following issues were found:\n")
		for _, ve := range validationErrors {
			sb.WriteString("• ")
			sb.WriteString(ve)
			sb.WriteString("\n")
		}
		sb.WriteString("\nPlease adjust your selections or create multiple bookings for large requests.")
		return nil, errors.New(sb.String())
	}

	if totalServiceHours > MaxDailyHours {
		return nil, fmt.Errorf("total service hours (%d) exceed daily limit of %d hours. Please divide your cleaning into multiple bookings (e.g., book different services on separate days)",
			totalServiceHours, MaxDailyHours)
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

// Helper function
func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
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

	// Handle error from main service calculation
	subtotal, mainHours, err := p.CalculatePriceByServiceType(mainService)
	if err != nil {
		return nil, fmt.Errorf("main service validation failed: %v", err)
	}

	var addonTotal float32 = 0
	var addonTotalHours int32 = 0
	var validationErrors []string

	// Validate each addon
	for i, addon := range in.Addons {
		addonService := &types.ServicesRequest{
			ServiceType: addon.ServiceDetail.ServiceType,
			Details:     addon.ServiceDetail.Details,
		}

		addonPrice, addonHours, err := p.CalculatePriceByServiceType(addonService)
		if err != nil {
			validationErrors = append(validationErrors,
				fmt.Sprintf("Addon %d (%s): %v", i+1, addon.ServiceDetail.ServiceType, err))
			continue
		}

		// Check if adding this addon would exceed daily limit
		if mainHours+addonTotalHours+addonHours > 11 {
			validationErrors = append(validationErrors,
				fmt.Sprintf("Addon %d (%s) would exceed daily limit of 11 hours. Current total: %d hours, Addon requires: %d hours",
					i+1, addon.ServiceDetail.ServiceType, mainHours+addonTotalHours, addonHours))
			continue
		}

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

	if len(validationErrors) > 0 {
		var sb strings.Builder
		sb.WriteString("The following issues were found:\n")
		for _, ve := range validationErrors {
			sb.WriteString("• ")
			sb.WriteString(ve)
			sb.WriteString("\n")
		}
		sb.WriteString("\nPlease adjust your selections or create multiple bookings for large requests.")
		return nil, errors.New(sb.String())
	}

	totalPrice := subtotal + addonTotal
	totalServiceHours := mainHours + addonTotalHours

	// Final validation
	if totalServiceHours > 11 {
		return nil, fmt.Errorf("total service hours (%d) exceed daily limit of 11 hours. Please divide your cleaning into multiple bookings",
			totalServiceHours)
	}

	err = tx.QueryRow(c, `
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
	var addonTotal float32 = 0.00
	if req.AddonTotal != nil {
		addonTotal = *req.AddonTotal
	}
	const query = `
		INSERT INTO payment.orders (
			order_number,
			full_payment_method,
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
			$10, $11,
			NOW(), NOW()
		)
		RETURNING id;
	`

	var orderID string

	err := tx.QueryRow(ctx, query,
		utils.GenerateOrderNumber(req.QuoteID, time.Now()),
		req.PaymentMethod,
		req.CustomerID,
		req.QuoteID,
		"PHP",
		req.Subtotal,
		addonTotal,
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
		&order.PaymentMethod,
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
			&o.PaymentMethod,
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
			&o.PaymentMethod,
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
		orderId, startDate, endDate, page, limit,
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
		customerId, startDate, endDate, page, limit,
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
func (s *PaymentTasks) StorePayment(ctx context.Context, tx pgx.Tx, payment *types.StorePayment) error {
	const query = `
		INSERT INTO payment.payments (
			order_id,
			client_key,
			amount,
			currency,
			failed_reason,
			payment_id,
			payment_intent_id,
			status,
			type,
			provider,
			raw_response,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
	`
	_, err := tx.Exec(ctx, query,
		payment.OrderID,
		payment.ClientKey,
		payment.Amount,
		payment.Currency,
		payment.FailedReason,
		payment.PaymentID,
		payment.PaymentIntentID,
		payment.Status,
		payment.Type,
		payment.Provider,
		payment.RawResponse,
	)
	return err
}
func (s *PaymentTasks) UpdateOrderPaymentStatus(ctx context.Context, tx pgx.Tx, paymentIntentId, paymentId, newStatus string) error {
	const query = `
		UPDATE payment.orders o
		SET payment_status = $1, updated_at = NOW(), payment_id = $3
		SET payment_status = $1, updated_at = NOW(), payment_id = $3
		FROM payment.payments p
		WHERE p.order_id = o.id
		  AND p.payment_intent_id = $2
	`
	_, err := tx.Exec(ctx, query, newStatus, paymentIntentId, paymentId)
	return err
}

func (s *PaymentTasks) CheckExistingDownpayment(ctx context.Context, tx pgx.Tx, orderID string) (*types.ExistingDownpaymentResponse, error) {
	const query = `
		SELECT p.client_key, p.payment_intent_id
		FROM payment.orders o
		JOIN payment.payments p ON p.order_id = o.id
		WHERE o.id = $1
		  AND p.type = 'DOWNPAYMENT'
		  AND p.status = 'awaiting_payment_method'
		ORDER BY p.created_at DESC
		LIMIT 1
	`

	var clientKey, paymentIntentId string
	err := tx.QueryRow(ctx, query, orderID).Scan(&clientKey, &paymentIntentId)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &types.ExistingDownpaymentResponse{
				HasExistingDownpayment: false,
			}, nil
		}
		return &types.ExistingDownpaymentResponse{
			HasExistingDownpayment: false,
		}, err
	}

	return &types.ExistingDownpaymentResponse{
		HasExistingDownpayment: true,
		ClientKey:              &clientKey,
		PaymentIntentID:        &paymentIntentId,
	}, nil
}
