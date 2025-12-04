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
	g, c := errgroup.WithContext(ctx)

	var (
		prices   *types.CleaningPrices
		alloc    *types.CleaningAllocation
		cleaners []types.CleanerAssigned
	)

	g.Go(func() error {
		var err error
		prices, err = paymentPort.GetQuotePrices(c, req.Base.QuoteId)
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

	return &types.BookingAllocation{
		CleaningAllocation: alloc,
		CleanerAssigned:    cleaners,
		CleaningPrices:     prices,
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
	cleaners := []types.CleanerAssigned{
		{ID: "7aa794fa-e3f9-446f-8368-bcb55518bc29", CleanerFirstName: "Charles", CleanerLastName: "Boquecosa"},
		{ID: "cb32d23a-31a8-4461-ba3e-228d418ba6f3", CleanerFirstName: "Clarence", CleanerLastName: "Diangco"},
	}
	return cleaners, nil
}

// makeBaseBooking inserts into booking.basebookings and returns the created BaseBookingDetails.
func (t *BookingTasks) MakeBaseBooking(
	c context.Context,
	tx pgx.Tx,
	custID string,
	customerFirstName string,
	customerLastName string,
	address types.Address,
	startSched time.Time,
	endSched time.Time,
	dirtyScale int32,
	photos []string,
	quoteId string,
) (*types.BaseBookingDetails, error) {

	var createdBaseBook types.BaseBookingDetails

	err := tx.QueryRow(c,
		`INSERT INTO booking.basebookings (
            cust_id,
            customer_first_name,
            customer_last_name,
            address,
            start_sched,
            end_sched,
            dirty_scale,
            payment_status,
            review_status,
            photos,
            created_at,
            updated_at,
            quote_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
        RETURNING id, cust_id, customer_first_name, customer_last_name, address, start_sched, end_sched, dirty_scale, payment_status, review_status, photos, created_at, updated_at, quote_id`,
		custID,
		customerFirstName,
		customerLastName,
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
	).Scan(
		&createdBaseBook.ID,
		&createdBaseBook.CustID,
		&createdBaseBook.CustomerFirstName,
		&createdBaseBook.CustomerLastName,
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
		return insertServiceDetails(ctx, tx, types.ServiceMattress, types.MattressCleaningDetails{CleaningSpecs: specs}, logger)
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
		return insertServiceDetails(ctx, tx, types.ServicePost, types.PostConstructionDetails{SQM: mainService.Post.SQM}, logger)
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
	totalPrice float32,
) (string, error) {
	var id string
	query := `
		INSERT INTO booking.bookings 
		(base_booking_id, main_service_id, addon_ids, equipment_ids, resource_ids, cleaner_ids, total_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	err := tx.QueryRow(ctx, query,
		baseBookingID,
		mainServiceID,
		addonIDs,
		equipmentIDs,
		resourceIDs,
		cleanerIDs,
		totalPrice,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("saveBooking: %w", err)
	}

	return id, nil
}

func (t *BookingTasks) FetchBookingById(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (*types.Booking, error) {

	var (
		baseBookingID string
		mainServiceID string
		addonIDs      []string
		equipmentIDs  []string
		resourceIDs   []string
		cleanerIDs    []string
		totalPrice    float32
	)

	query := `
		SELECT base_booking_id, main_service_id, addon_ids, 
		       equipment_ids, resource_ids, cleaner_ids, total_price
		FROM booking.bookings
		WHERE id = $1
	`

	if err := tx.QueryRow(ctx, query,
		id,
	).Scan(
		&baseBookingID, &mainServiceID, &addonIDs,
		&equipmentIDs, &resourceIDs, &cleanerIDs, &totalPrice,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("booking not found: %w", err)
		}
		return nil, fmt.Errorf("fetch booking row: %w", err)
	}

	base, err := loadBaseBooking(ctx, tx, baseBookingID)
	if err != nil {
		return nil, err
	}

	mainSvc, err := loadServiceDetails(ctx, tx, mainServiceID)
	if err != nil {
		return nil, fmt.Errorf("load main service: %w", err)
	}

	addons, err := loadAddOns(ctx, tx, addonIDs)
	if err != nil {
		return nil, err
	}

	equipments, err := loadEquipments(ctx, tx, equipmentIDs)
	if err != nil {
		return nil, err
	}

	resources, err := loadResources(ctx, tx, resourceIDs)
	if err != nil {
		return nil, err
	}

	cleaners, err := loadCleaners(ctx, tx, cleanerIDs)
	if err != nil {
		return nil, err
	}

	return &types.Booking{
		ID:          id,
		Base:        *base,
		MainService: *mainSvc,
		Addons:      addons,
		Equipments:  equipments,
		Resources:   resources,
		Cleaners:    cleaners,
		TotalPrice:  totalPrice,
	}, nil
}

func (t *BookingTasks) FetchAllBookingsByUserID(
	ctx context.Context,
	tx pgx.Tx,
	custId string,
) ([]*types.Booking, error) {

	// var exists bool
	// err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM account.accounts WHERE id = $1)", userId).Scan(&exists)
	// if err != nil {
	// 	return nil, fmt.Errorf("checking user existence: %w", err)
	// }
	// if !exists {
	// 	return nil, fmt.Errorf("user with id %s does not exist", userId)
	// }

	// Query all bookings for the user using join with basebookings

	query := `
		SELECT b.id, b.base_booking_id, b.main_service_id, b.addon_ids,
		       b.equipment_ids, b.resource_ids, b.cleaner_ids, b.total_price
		FROM booking.bookings b
		JOIN booking.basebookings bb ON b.base_booking_id = bb.id
		WHERE bb.cust_id = $1
	`

	rows, err := tx.Query(ctx, query, custId)
	if err != nil {
		return nil, fmt.Errorf("fetch bookings: %w", err)
	}
	defer rows.Close()

	type bookingRow struct {
		bookingID, baseBookingID, mainServiceID         string
		addonIDs, equipmentIDs, resourceIDs, cleanerIDs []string
		totalPrice                                      float32
	}

	var rowsData []bookingRow
	for rows.Next() {
		var r bookingRow
		if err := rows.Scan(
			&r.bookingID, &r.baseBookingID, &r.mainServiceID,
			&r.addonIDs, &r.equipmentIDs, &r.resourceIDs,
			&r.cleanerIDs, &r.totalPrice,
		); err != nil {
			return nil, fmt.Errorf("scan booking row: %w", err)
		}
		rowsData = append(rowsData, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate booking rows: %w", err)
	}

	if len(rowsData) == 0 {
		return nil, fmt.Errorf("no bookings found for user with id=%s", custId)
	}
	var bookings []*types.Booking
	for _, r := range rowsData {
		base, err := loadBaseBooking(ctx, tx, r.baseBookingID)
		if err != nil {
			return nil, err
		}

		main, err := loadServiceDetails(ctx, tx, r.mainServiceID)
		if err != nil {
			return nil, err
		}

		addons, err := loadAddOns(ctx, tx, r.addonIDs)
		if err != nil {
			return nil, err
		}

		equipments, err := loadEquipments(ctx, tx, r.equipmentIDs)
		if err != nil {
			return nil, err
		}

		resources, err := loadResources(ctx, tx, r.resourceIDs)
		if err != nil {
			return nil, err
		}

		cleaners, err := loadCleaners(ctx, tx, r.cleanerIDs)
		if err != nil {
			return nil, err
		}

		bookings = append(bookings, &types.Booking{
			ID:          r.bookingID,
			Base:        *base,
			MainService: *main,
			Addons:      addons,
			Equipments:  equipments,
			Resources:   resources,
			Cleaners:    cleaners,
			TotalPrice:  r.totalPrice,
		})
	}

	return bookings, nil
}

func (t *BookingTasks) ModifyBookingByID(
	ctx context.Context,
	tx pgx.Tx,
	bookingId string,
	newStartSched, newEndSched time.Time,
	newDirtyScale int32,
	newPhotos []string,
) (*types.Booking, error) {

	var baseBookingID string
	var mainServiceID string
	var addonIDs, equipmentIDs, resourceIDs, cleanerIDs []string
	var totalPrice float32

	getQuery := `
		SELECT base_booking_id, main_service_id, addon_ids, equipment_ids, 
		       resource_ids, cleaner_ids, total_price
		FROM booking.bookings
		WHERE id = $1
	`

	err := tx.QueryRow(ctx, getQuery, bookingId).Scan(
		&baseBookingID, &mainServiceID,
		&addonIDs, &equipmentIDs,
		&resourceIDs, &cleanerIDs,
		&totalPrice,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("booking with id %s does not exist", bookingId)
		}
		return nil, fmt.Errorf("fetch booking details: %w", err)
	}

	updateQuery := `
		UPDATE booking.basebookings
		SET start_sched = $1,
		    end_sched   = $2,
		    dirty_scale = $3,
		    photos      = $4,
		    updated_at  = NOW()
		WHERE id = $5
	`

	_, err = tx.Exec(ctx, updateQuery,
		newStartSched,
		newEndSched,
		newDirtyScale,
		newPhotos,
		baseBookingID,
	)
	if err != nil {
		return nil, fmt.Errorf("update basebooking: %w", err)
	}

	base, err := loadBaseBooking(ctx, tx, baseBookingID)
	if err != nil {
		return nil, err
	}

	main, err := loadServiceDetails(ctx, tx, mainServiceID)
	if err != nil {
		return nil, err
	}

	addons, err := loadAddOns(ctx, tx, addonIDs)
	if err != nil {
		return nil, err
	}

	equipments, err := loadEquipments(ctx, tx, equipmentIDs)
	if err != nil {
		return nil, err
	}

	resources, err := loadResources(ctx, tx, resourceIDs)
	if err != nil {
		return nil, err
	}

	cleaners, err := loadCleaners(ctx, tx, cleanerIDs)
	if err != nil {
		return nil, err
	}

	return &types.Booking{
		ID:          bookingId,
		Base:        *base,
		MainService: *main,
		Addons:      addons,
		Equipments:  equipments,
		Resources:   resources,
		Cleaners:    cleaners,
		TotalPrice:  totalPrice,
	}, nil
}

func (t *BookingTasks) DeleteBookingByID(
	ctx context.Context,
	tx pgx.Tx,
	bookingId string,
) error {
	// First, get the base_booking_id of the booking
	var baseBookingID string
	err := tx.QueryRow(ctx, `
		SELECT base_booking_id
		FROM booking.bookings
		WHERE id = $1
	`, bookingId).Scan(&baseBookingID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("booking with id=%s not found", bookingId)
		}
		return fmt.Errorf("fetch booking for deletion: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM booking.bookings
		WHERE id = $1
	`, bookingId)
	if err != nil {
		return fmt.Errorf("delete booking: %w", err)
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM booking.basebookings
		WHERE id = $1
	`, baseBookingID)
	if err != nil {
		return fmt.Errorf("delete base booking: %w", err)
	}

	return nil
}
