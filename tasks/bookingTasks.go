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

type BookingService interface {
	CreateBooking(ctx context.Context, evt types.CreateBookingEvent) (*types.Booking, error)
}

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

func loadBaseBooking(ctx context.Context, tx pgx.Tx, id string) (*types.BaseBookingDetails, error) {
	var base types.BaseBookingDetails

	query := `
		SELECT id, cust_id, customer_first_name, customer_last_name,
		       address, start_sched, end_sched, dirty_scale,
		       payment_status, review_status, photos,
		       created_at, updated_at, quote_id
		FROM booking.basebookings
		WHERE id = $1
	`

	if err := tx.QueryRow(ctx, query, id).Scan(
		&base.ID, &base.CustID, &base.CustomerFirstName,
		&base.CustomerLastName, &base.Address,
		&base.StartSched, &base.EndSched, &base.DirtyScale,
		&base.PaymentStatus, &base.ReviewStatus, &base.Photos,
		&base.CreatedAt, &base.UpdatedAt, &base.QuoteId,
	); err != nil {
		return nil, fmt.Errorf("load base booking: %w", err)
	}

	return &base, nil
}

func loadServiceDetails(ctx context.Context, tx pgx.Tx, svcID string) (*types.ServiceDetails, error) {
	var svc types.ServiceDetails
	var raw []byte
	var svcType string

	err := tx.QueryRow(ctx, `
		SELECT id, service_type, details
		FROM booking.services
		WHERE id = $1
	`, svcID).Scan(&svc.ID, &svcType, &raw)

	if err != nil {
		return nil, fmt.Errorf("load service %s: %w", svcID, err)
	}

	svc.ServiceType = svcType

	factory, ok := types.DetailFactories[types.DetailType(svcType)]
	if !ok {
		// fallback to map[string]any
		var m any
		if err := json.Unmarshal(raw, &m); err != nil {
			return nil, fmt.Errorf("service unmarshal fallback: %w", err)
		}
		svc.Details = m
		return &svc, nil
	}

	out := factory()
	if err := json.Unmarshal(raw, out); err != nil {
		return nil, fmt.Errorf("service unmarshal: %w", err)
	}

	svc.Details = out
	return &svc, nil
}

func loadAddOns(ctx context.Context, tx pgx.Tx, ids []string) ([]types.AddOns, error) {
	if len(ids) == 0 {
		return []types.AddOns{}, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT a.id, a.service_id, a.price, 
		       s.id, s.service_type, s.details
		FROM booking.addons a
		JOIN booking.services s ON a.service_id = s.id
		WHERE a.id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query addons: %w", err)
	}
	defer rows.Close()

	var addons []types.AddOns

	for rows.Next() {
		var addID, serviceID, svcType string
		var price float32
		var raw []byte

		if err := rows.Scan(&addID, &serviceID, &price, &serviceID, &svcType, &raw); err != nil {
			return nil, fmt.Errorf("scan addon: %w", err)
		}

		svc, err := parseServiceFromRow(serviceID, svcType, raw)
		if err != nil {
			return nil, err
		}

		addons = append(addons, types.AddOns{
			ID:            addID,
			ServiceDetail: *svc,
			Price:         price,
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("iterate addons: %w", rows.Err())
	}

	return addons, nil
}

func parseServiceFromRow(id, svcType string, raw []byte) (*types.ServiceDetails, error) {
	svc := types.ServiceDetails{
		ID:          id,
		ServiceType: svcType,
	}

	if factory, ok := types.DetailFactories[types.DetailType(svcType)]; ok {
		out := factory()
		if err := json.Unmarshal(raw, out); err != nil {
			return nil, fmt.Errorf("service unmarshal: %w", err)
		}
		svc.Details = out
		return &svc, nil
	}

	// fallback
	var m any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("service fallback unmarshal: %w", err)
	}
	svc.Details = m
	return &svc, nil
}

func loadEquipments(ctx context.Context, tx pgx.Tx, ids []string) ([]types.CleaningEquipment, error) {
	if len(ids) == 0 {
		return []types.CleaningEquipment{}, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT id, name, type, photoUrl
		FROM booking.equipments
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query equipments: %w", err)
	}
	defer rows.Close()

	var eq []types.CleaningEquipment
	for rows.Next() {
		var e types.CleaningEquipment
		if err := rows.Scan(&e.ID, &e.Name, &e.Type, &e.PhotoURL); err != nil {
			return nil, fmt.Errorf("scan equipment: %w", err)
		}
		eq = append(eq, e)
	}
	return eq, rows.Err()
}

func loadResources(ctx context.Context, tx pgx.Tx, ids []string) ([]types.CleaningResources, error) {
	if len(ids) == 0 {
		return []types.CleaningResources{}, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT id, name, type, photoUrl
		FROM booking.resources
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query resources: %w", err)
	}
	defer rows.Close()

	var out []types.CleaningResources
	for rows.Next() {
		var r types.CleaningResources
		if err := rows.Scan(&r.ID, &r.Name, &r.Type, &r.PhotoURL); err != nil {
			return nil, fmt.Errorf("scan resource: %w", err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func loadCleaners(ctx context.Context, tx pgx.Tx, ids []string) ([]types.CleanerAssigned, error) {
	if len(ids) == 0 {
		return []types.CleanerAssigned{}, nil
	}

	rows, err := tx.Query(ctx, `
		SELECT id, cleanerFirstName, cleanerLastName, pfpUrl
		FROM booking.cleaners
		WHERE id = ANY($1)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("query cleaners: %w", err)
	}
	defer rows.Close()

	var out []types.CleanerAssigned
	for rows.Next() {
		var c types.CleanerAssigned
		if err := rows.Scan(&c.ID, &c.CleanerFirstName, &c.CleanerLastName, &c.PFPUrl); err != nil {
			return nil, fmt.Errorf("scan cleaner: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (t *BookingTasks) FetchBookingById(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (*types.Booking, error) {

	// Load main row
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

	// Load base booking
	base, err := loadBaseBooking(ctx, tx, baseBookingID)
	if err != nil {
		return nil, err
	}

	// Load services
	mainSvc, err := loadServiceDetails(ctx, tx, mainServiceID)
	if err != nil {
		return nil, fmt.Errorf("load main service: %w", err)
	}

	// Load extras
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
