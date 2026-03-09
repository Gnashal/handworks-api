package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"
	"handworks-api/utils"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/sync/errgroup"
)

type BookingTasks struct{}
type PaymentPort interface {
	GetQuotePrices(ctx context.Context, quoteId string) (*types.CleaningPrices, error)
}

func (t *BookingTasks) AllocateAll(ctx context.Context, paymentPort PaymentPort, req *types.CreateBookingRequest) (*types.BookingAllocation, error) {
	// Validate extra hours before proceeding
	if req.ExtraHours > 0 {
		if req.MainService.ServiceType != types.GeneralCleaning {
			return nil, fmt.Errorf("extra hours can only be added for General Cleaning services")
		}

		// Optional: Add maximum hours limit
		if req.ExtraHours > 4 {
			return nil, fmt.Errorf("extra hours cannot exceed 4 hours")
		}
	}

	g, c := errgroup.WithContext(ctx)

	var (
		prices           *types.CleaningPrices
		alloc            *types.CleaningAllocation
		cleaners         []types.CleanerAssigned
		extraHourCost    float32
		originalEndSched time.Time
	)

	g.Go(func() error {
		var err error
		prices, err = paymentPort.GetQuotePrices(c, req.QuoteId)
		return err
	})

	g.Go(func() error {
		var err error
		alloc, err = t.AllocateEquipmentAndResources(c, req)
		return err
	})

	g.Go(func() error {
		var err error
		cleaners, err = t.AllocateCleaners(c, req)

		// Calculate extra hour cost if extra hours requested and cleaners are allocated
		if err == nil && req.ExtraHours > 0 && len(cleaners) > 0 {
			extraHourCost = req.ExtraHours * 250.00 * float32(len(cleaners))

			// Store original end time before extension
			originalEndSched = req.Base.EndSched

			// Calculate new end time with extra hours
			duration := time.Duration(req.ExtraHours * float32(time.Hour))
			newEndSched := req.Base.EndSched.Add(duration)
			req.Base.EndSched = newEndSched
		}
		return err
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Defensive nils
	if prices == nil {
		prices = &types.CleaningPrices{}
	}
	if alloc == nil {
		alloc = &types.CleaningAllocation{}
	}

	// Add extra hour cost to prices
	if extraHourCost > 0 {
		prices.ExtraHourCost = extraHourCost
		prices.MainServicePrice += extraHourCost
	}

	return &types.BookingAllocation{
		CleaningAllocation: alloc,
		CleanerAssigned:    cleaners,
		CleaningPrices:     prices,
		ExtraHours:         req.ExtraHours,
		ExtraHourCost:      extraHourCost,
		OriginalEndSched:   originalEndSched,
	}, nil
}

func (t *BookingTasks) AllocateEquipmentAndResources(ctx context.Context, req *types.CreateBookingRequest) (*types.CleaningAllocation, error) {
	// FOR TESTING PA NI, I HAVE NOT IMPLEMENTED THE REAL LOGIC YET
	// TODO: Automation logic for resource and equipment allocation
	equipments := []types.CleaningEquipment{
		{ID: "7849f478-f70b-42a7-82d2-aadc81d3e6d6", Name: "Vacuum Cleaner", Type: "Electrical", PhotoURL: "https://example.com/vacuum.jpg"},
		{ID: "a4cd7e23-787b-4344-80d7-c50199d85ecd", Name: "Mop", Type: "Manual", PhotoURL: "https://example.com/mop.jpg"},
	}
	resources := []types.CleaningResources{
		{ID: "d1e94940-838d-4916-bf2b-bb09b77d7c46", Name: "Detergent", Type: "Chemical", PhotoURL: "https://example.com/detergent.jpg"},
	}
	return &types.CleaningAllocation{
		CleaningEquipment: equipments,
		CleaningResources: resources,
	}, nil
}

func (t *BookingTasks) AllocateCleaners(ctx context.Context, req *types.CreateBookingRequest) ([]types.CleanerAssigned, error) {
	// FOR TESTING PA NI, I HAVE NOT IMPLEMENTED THE REAL LOGIC YET
	// TODO: Automation logic for cleaner assignment

	// In real implementation, number of cleaners should be determined by:
	// - Service type
	// - Square meters (for general cleaning)
	// - Number of items (for couch/mattress/car)
	// - Extra hours requested

	cleaners := []types.CleanerAssigned{
		{ID: "7aa794fa-e3f9-446f-8368-bcb55518bc29", CleanerFirstName: "Charles", CleanerLastName: "Boquecosa"},
		{ID: "cb32d23a-31a8-4461-ba3e-228d418ba6f3", CleanerFirstName: "Clarence", CleanerLastName: "Diangco"},
	}

	return cleaners, nil
}

// makeBaseBooking inserts into booking.basebookings and returns the created BaseBookingDetails.
func (t *BookingTasks) MakeBaseBooking(
	ctx context.Context,
	tx pgx.Tx,
	custID string,
	customerFirstName string,
	customerLastName string,
	customerPhoneNo string,
	address types.Address,
	startSched time.Time,
	endSched time.Time,
	dirtyScale int32,
	photos []string,
	quoteId string,
	extraHours float32,
	extraHourCost float32,
	originalEndSched *time.Time,
) (*types.BaseBookingDetails, error) {

	var createdBaseBook types.BaseBookingDetails

	err := tx.QueryRow(ctx,
		`INSERT INTO booking.basebookings (
            custid,
            customerfirstname,
            customerlastname,
            customer_phone_no,
            address,
            startsched,
            endsched,
            dirtyscale,
            paymentstatus,
            reviewstatus,
            photos,
            createdat,
            updatedat,
            quoteid,
            extra_hours,
            extra_hour_cost,
            original_end_sched
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING id, custid, customerfirstname, customerlastname, customer_phone_no, address, 
            startsched, endsched, dirtyscale, paymentstatus, reviewstatus, 
            photos, createdat, updatedat, quoteid, extra_hours, extra_hour_cost, original_end_sched`,
		custID,
		customerFirstName,
		customerLastName,
		customerPhoneNo,
		address,
		startSched,
		endSched,
		dirtyScale,
		"UNPAID",
		"PENDING",
		photos,
		time.Now(),
		time.Now(),
		quoteId,
		extraHours,
		extraHourCost,
		originalEndSched,
	).Scan(
		&createdBaseBook.ID,
		&createdBaseBook.CustID,
		&createdBaseBook.CustomerFirstName,
		&createdBaseBook.CustomerLastName,
		&createdBaseBook.CustomerPhoneNo,
		&createdBaseBook.Address,
		&createdBaseBook.StartSched,
		&createdBaseBook.EndSched,
		&createdBaseBook.DirtyScale,
		&createdBaseBook.PaymentStatus,
		&createdBaseBook.ReviewStatus,
		&createdBaseBook.Photos,
		&createdBaseBook.CreatedAt,
		&createdBaseBook.UpdatedAt,
		&createdBaseBook.QuoteId,
		&createdBaseBook.ExtraHours,
		&createdBaseBook.ExtraHourCost,
		&createdBaseBook.OriginalEndSched,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert base booking: %w", err)
	}

	return &createdBaseBook, nil
}

// insertServiceDetails is a generic helper to persist service detail JSON and return a types.ServiceDetails with Details populated.
func insertServiceDetails[T any](ctx context.Context, tx pgx.Tx, serviceType types.DetailType, details T, log *utils.Logger) (*types.ServiceDetails, error) {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, fmt.Errorf("marshal details: %w", err)
	}

	var svc types.ServiceDetails
	var raw []byte

	err = tx.QueryRow(ctx, `
		INSERT INTO booking.services (service_type, details)
		VALUES ($1, $2)
		RETURNING id, service_type, details
	`, serviceType, detailsJSON).Scan(&svc.ID, &svc.ServiceType, &raw)
	if err != nil {
		return nil, fmt.Errorf("insert service: %w", err)
	}
	log.Debug("Raw JSON: %s", string(raw))

	// Unmarshal dynamically using factory
	factory, ok := types.DetailFactories[serviceType]
	if !ok {
		return nil, fmt.Errorf("no factory registered for service type %s", serviceType)
	}

	out := factory()
	if err := json.Unmarshal(raw, out); err != nil {
		log.Error("Unmarshal error: %v", err)
		return nil, fmt.Errorf("unmarshal details: %w", err)
	}
	svc.Details = out
	b, _ := json.MarshalIndent(svc.Details, "", "  ")
	log.Debug("Post-Unmarshal JSON: %s", string(b))
	return &svc, nil
}

// createMainServiceBooking converts types.ServiceDetail into DB records by using insertServiceDetails.
func (t *BookingTasks) CreateMainServiceBooking(
	ctx context.Context,
	tx pgx.Tx,
	logger *utils.Logger,
	mainService types.ServiceDetail,
) (*types.ServiceDetails, error) {

	// detect which union field is set
	if mainService.General != nil {
		d := types.GeneralCleaningDetails{
			HomeType: mainService.General.HomeType,
			SQM:      mainService.General.SQM,
			Hours:    mainService.General.Hours,
		}
		return insertServiceDetails(ctx, tx, types.ServiceGeneral, d, logger)
	}

	if mainService.Couch != nil {
		specs := make([]types.CouchCleaningSpecifications, len(mainService.Couch.CleaningSpecs))
		for i, ssp := range mainService.Couch.CleaningSpecs {
			specs[i] = types.CouchCleaningSpecifications{
				CouchType: ssp.CouchType,
				WidthCM:   ssp.WidthCM,
				DepthCM:   ssp.DepthCM,
				HeightCM:  ssp.HeightCM,
				Quantity:  ssp.Quantity,
			}
		}
		d := types.CouchCleaningDetails{
			BedPillows:    mainService.Couch.BedPillows,
			CleaningSpecs: specs,
		}
		return insertServiceDetails(ctx, tx, types.ServiceCouch, d, logger)
	}

	if mainService.Mattress != nil {
		specs := make([]types.MattressCleaningSpecifications, len(mainService.Mattress.CleaningSpecs))
		for i, ssp := range mainService.Mattress.CleaningSpecs {
			specs[i] = types.MattressCleaningSpecifications{
				BedType:  ssp.BedType,
				WidthCM:  ssp.WidthCM,
				DepthCM:  ssp.DepthCM,
				HeightCM: ssp.HeightCM,
				Quantity: ssp.Quantity,
			}
		}
		d := types.MattressCleaningDetails{CleaningSpecs: specs}
		return insertServiceDetails(ctx, tx, types.ServiceMattress, d, logger)
	}

	if mainService.Car != nil {
		specs := make([]types.CarCleaningSpecifications, len(mainService.Car.CleaningSpecs))
		for i, ssp := range mainService.Car.CleaningSpecs {
			specs[i] = types.CarCleaningSpecifications{
				CarType:  ssp.CarType,
				Quantity: ssp.Quantity,
			}
		}
		d := types.CarCleaningDetails{
			CleaningSpecs: specs,
			ChildSeats:    mainService.Car.ChildSeats,
		}
		return insertServiceDetails(ctx, tx, types.ServiceCar, d, logger)
	}

	if mainService.Post != nil {
		d := types.PostConstructionDetails{SQM: mainService.Post.SQM}
		return insertServiceDetails(ctx, tx, types.ServicePost, d, logger)
	}

	return nil, fmt.Errorf("unsupported main service type")
}

// createAddOn creates the service details for the addon and inserts into booking.addons.
func (t *BookingTasks) CreateAddOn(
	ctx context.Context,
	tx pgx.Tx,
	logger *utils.Logger,
	addonReq types.AddOnRequest,
	addOnPrice float32,
) (*types.AddOns, error) {
	// create underlying service row
	addOnServiceDetails, err := t.CreateMainServiceBooking(ctx, tx, logger, addonReq.ServiceDetail.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to create service details: %w", err)
	}
	// debug log of proto-like structure (optional)
	out, _ := json.MarshalIndent(addOnServiceDetails.Details, "", "  ")
	logger.Debug("Create Addon Service Details: %s", string(out))

	createdAddon := &types.AddOns{
		ServiceDetail: *addOnServiceDetails,
		Price:         addOnPrice,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO booking.addons 
		 (service_id, price)
		 VALUES ($1, $2)
		 RETURNING id, service_id, price`,
		addOnServiceDetails.ID,
		addOnPrice,
	).Scan(
		&createdAddon.ID,
		&createdAddon.ServiceDetail.ID,
		&createdAddon.Price,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert addon: %w", err)
	}

	return createdAddon, nil
}

// saveBooking persists the booking composite row and returns the booking id.
func (t *BookingTasks) SaveBooking(
	ctx context.Context,
	tx pgx.Tx,
	baseBookingID, mainServiceID string,
	addonIDs, equipmentIDs, resourceIDs, cleanerIDs []string,
	quoteTotalPrice float32, // Original quote total price
	extraHourCost float32, // Calculated extra hours cost
) (string, error) {
	var id string

	// Calculate final price
	finalTotalPrice := quoteTotalPrice + extraHourCost

	query := `
		INSERT INTO booking.bookings 
		(base_booking_id, main_service_id, addon_ids, equipment_ids, resource_ids, cleaner_ids, 
		 total_price, extra_hour_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err := tx.QueryRow(ctx, query,
		baseBookingID,
		mainServiceID,
		addonIDs,
		equipmentIDs,
		resourceIDs,
		cleanerIDs,
		finalTotalPrice,
		extraHourCost,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("saveBooking: %w", err)
	}

	return id, nil
}

func (t *BookingTasks) FetchBookingByID(ctx context.Context, tx pgx.Tx, bookingID string, logger *utils.Logger) (*types.Booking, error) {

	var rawJSON []byte
	err := tx.QueryRow(ctx,
		`SELECT booking.get_booking_by_id($1)`,
		bookingID).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_booking_by_id: %w", err)
	}

	var booking types.Booking
	if err := json.Unmarshal(rawJSON, &booking); err != nil {
		logger.Error("failed to unmarshal booking JSON: %v", err)
		return nil, fmt.Errorf("unmarhsal booking: %w", err)
	}

	return &booking, nil
}

func (t *BookingTasks) FetchAllBookings(
	ctx context.Context,
	tx pgx.Tx,
	startDate, endDate string,
	page, limit int,
	logger *utils.Logger,
) (*types.FetchAllBookingsResponse, error) {

	var rawJSON []byte

	// Convert empty strings to nil (NULL in database)
	var startDateArg, endDateArg interface{}

	if startDate == "" {
		startDateArg = nil
	} else {
		startDateArg = startDate
	}

	if endDate == "" {
		endDateArg = nil
	} else {
		endDateArg = endDate
	}

	err := tx.QueryRow(ctx,
		`SELECT booking.fetch_all_bookings($1, $2, $3, $4)`,
		startDateArg, endDateArg, page, limit).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc fetch_all_bookings: %w", err)
	}

	var response types.FetchAllBookingsResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal bookings JSON: %v", err)
		return nil, fmt.Errorf("unmarshal bookings: %w", err)
	}

	return &response, nil
}

func (t *BookingTasks) FetchAllCustomerBookings(
	ctx context.Context,
	tx pgx.Tx,
	customerId, startDate, endDate string,
	page, limit int,
	logger *utils.Logger,
) (*types.FetchAllBookingsResponse, error) {
	var rawJSON []byte

	err := tx.QueryRow(ctx,
		`SELECT booking.get_bookings_by_customer($1, $2::date, $3::date, $4, $5)`,
		customerId, startDate, endDate, page, limit,
	).Scan(&rawJSON)

	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_bookings_by_customer: %w", err)
	}

	var response types.FetchAllBookingsResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookings: %w", err)
	}

	return &response, nil
}

func (t *BookingTasks) FetchAllEmployeeAssignedBookings(
	ctx context.Context,
	tx pgx.Tx,
	employeeId, startDate, endDate string,
	page, limit int,
	logger *utils.Logger,
) (*types.FetchAllBookingsResponse, error) {
	var rawJSON []byte
	var response types.FetchAllBookingsResponse
	err := tx.QueryRow(ctx,
		`SELECT booking.get_bookings_by_cleaner($1, $2, $3, $4, $5)`,
		employeeId, startDate, endDate, page, limit,
	).Scan(&rawJSON)
	if err != nil {
		return nil, fmt.Errorf("failed calling sproc get_bookings_by_employee: %w", err)
	}

	if err := json.Unmarshal(rawJSON, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bookings: %w", err)
	}

	return &response, nil
}

func (t *BookingTasks) FetchBookingSlots(
	ctx context.Context,
	tx pgx.Tx,
	selectedDate string,
	logger *utils.Logger,
) (*types.FetchSlotsResponse, error) {

	var rawJSON []byte

	err := tx.QueryRow(ctx,
		`SELECT booking.get_daily_booking_slots($1)`,
		selectedDate,
	).Scan(&rawJSON)

	if err != nil {
		logger.Error("failed to call get_daily_booking_slots sproc: %v", err)
		return nil, fmt.Errorf("failed calling sproc get_daily_booking_slots: %w", err)
	}

	var response types.FetchSlotsResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal booking slots JSON: %v", err)
		return nil, fmt.Errorf("unmarshal booking slots: %w", err)
	}

	return &response, nil
}
