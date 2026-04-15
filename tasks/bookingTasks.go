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

type BookingTasks struct{}
type PaymentPort interface {
	FetchOrderAndPrices(ctx context.Context, orderId string) (*types.Order, *types.CleaningPrices, error)
}

func (t *BookingTasks) FetchOrderAndPrices(ctx context.Context, paymentPort PaymentPort, orderId string) (*types.Order, *types.CleaningPrices, error) {
	order, prices, err := paymentPort.FetchOrderAndPrices(ctx, orderId)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch order and prices: %w", err)
	}
	return order, prices, nil
}

func (t *BookingTasks) AllocateEquipmentAndResources(ctx context.Context, tx pgx.Tx, req *types.CreateBookingRequest) (*types.CleaningAllocation, error) {
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

func (t *BookingTasks) AllocateCleaners(ctx context.Context, tx pgx.Tx) ([]types.CleanerAssigned, error) {
	query := `
		SELECT e.id, a.first_name, a.last_name
		FROM account.employees e
		JOIN account.accounts a ON a.id = e.account_id
		WHERE e.status = 'ACTIVE'
		AND e.position = 'cleaner'
		ORDER BY
			CASE WHEN e.id = ANY(
				SELECT UNNEST(b.cleaner_ids)
				FROM booking.bookings b
				JOIN booking.basebookings bb ON bb.id = b.base_booking_id
				ORDER BY bb.startsched DESC
				LIMIT 1
			) THEN 1 ELSE 0 END ASC`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active cleaners: %w", err)
	}
	defer rows.Close()

	var cleaners []types.CleanerAssigned
	for rows.Next() {
		var cleaner types.CleanerAssigned
		if err := rows.Scan(&cleaner.ID, &cleaner.CleanerFirstName, &cleaner.CleanerLastName); err != nil {
			return nil, fmt.Errorf("failed to scan cleaner row: %w", err)
		}
		cleaners = append(cleaners, cleaner)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed iterating cleaner rows: %w", err)
	}

	if len(cleaners) == 0 {
		return nil, fmt.Errorf("no available cleaners found")
	}
	return cleaners, nil
}

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
	orderId string,
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
            status,
            reviewstatus,
            photos,
            createdat,
            updatedat,
            orderid,
            extra_hours,
            extra_hour_cost,
            original_end_sched
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
        RETURNING id, custid, customerfirstname, customerlastname, customer_phone_no, address,
            startsched, endsched, dirtyscale, status, reviewstatus,
            photos, createdat, updatedat, orderid, extra_hours, extra_hour_cost, original_end_sched`,
		custID,
		customerFirstName,
		customerLastName,
		customerPhoneNo,
		address,
		startSched,
		endSched,
		dirtyScale,
		"NOT_STARTED",
		"PENDING",
		photos,
		time.Now(),
		time.Now(),
		orderId,
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
		&createdBaseBook.Status,
		&createdBaseBook.ReviewStatus,
		&createdBaseBook.Photos,
		&createdBaseBook.CreatedAt,
		&createdBaseBook.UpdatedAt,
		&createdBaseBook.OrderId,
		&createdBaseBook.ExtraHours,
		&createdBaseBook.ExtraHourCost,
		&createdBaseBook.OriginalEndSched,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to insert base booking: %w", err)
	}

	return &createdBaseBook, nil
}

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

func (t *BookingTasks) CreateMainServiceBooking(
	ctx context.Context,
	tx pgx.Tx,
	logger *utils.Logger,
	mainService types.ServiceDetail,
) (*types.ServiceDetails, error) {

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

func (t *BookingTasks) CreateAddOn(
	ctx context.Context,
	tx pgx.Tx,
	logger *utils.Logger,
	addonReq types.AddOnRequest,
	addOnPrice float32,
) (*types.AddOns, error) {
	addOnServiceDetails, err := t.CreateMainServiceBooking(ctx, tx, logger, addonReq.ServiceDetail.Details)
	if err != nil {
		return nil, fmt.Errorf("failed to create service details: %w", err)
	}
	out, _ := json.MarshalIndent(addOnServiceDetails.Details, "", "  ")
	logger.Debug("Create Addon Service Details: %s", string(out))

	createdAddon := &types.AddOns{
		ServiceDetail: *addOnServiceDetails,
		Price:         addOnPrice,
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO booking.addons (service_id, price)
		 VALUES ($1, $2)
		 RETURNING id, service_id, price`,
		addOnServiceDetails.ID, addOnPrice,
	).Scan(&createdAddon.ID, &createdAddon.ServiceDetail.ID, &createdAddon.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to insert addon: %w", err)
	}

	return createdAddon, nil
}

func (t *BookingTasks) SaveBooking(
	ctx context.Context,
	tx pgx.Tx,
	baseBookingID, mainServiceID string,
	addonIDs, equipmentIDs, resourceIDs, cleanerIDs []string,
	quoteTotalPrice float32,
	extraHourCost float32,
) (string, error) {
	var id string

	finalTotalPrice := quoteTotalPrice + extraHourCost

	err := tx.QueryRow(ctx, `
		INSERT INTO booking.bookings
		(base_booking_id, main_service_id, addon_ids, equipment_ids, resource_ids, cleaner_ids,
		 total_price, extra_hour_cost)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`,
		baseBookingID, mainServiceID,
		addonIDs, equipmentIDs, resourceIDs, cleanerIDs,
		finalTotalPrice, extraHourCost,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("saveBooking: %w", err)
	}

	return id, nil
}

func (t *BookingTasks) FetchBookingByID(ctx context.Context, tx pgx.Tx, bookingID string, logger *utils.Logger) (*types.Booking, error) {
	var rawJSON []byte
	err := tx.QueryRow(ctx, `SELECT booking.get_booking_by_id($1)`, bookingID).Scan(&rawJSON)
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
	customerId, startDate, endDate, status string,
	page, limit int,
	logger *utils.Logger,
) (*types.FetchAllBookingsResponse, error) {
	var rawJSON []byte

	err := tx.QueryRow(ctx,
		`SELECT booking.get_bookings_by_customer($1, $2, $3, $4, $5, $6)`,
		customerId, startDate, endDate, status, page, limit,
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

func (t *BookingTasks) ValidateSessionStart(ctx context.Context, tx pgx.Tx, bookingID string) (bool, time.Time, error) {
	query := `SELECT bb.startsched
		FROM booking.basebookings bb
		JOIN booking.bookings b ON b.base_booking_id = bb.id
		WHERE b.id = $1
		AND bb.status = 'NOT_STARTED'
		AND bb.reviewstatus = 'SCHEDULED'`

	var startSched time.Time
	err := tx.QueryRow(ctx, query, bookingID).Scan(&startSched)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, time.Time{}, nil
		}
		return false, time.Time{}, fmt.Errorf("failed to validate session start: %w", err)
	}

	return true, startSched, nil
}

func (t *BookingTasks) StartSession(ctx context.Context, tx pgx.Tx, bookingID string, startPhotos []string) error {
	if _, err := tx.Exec(ctx,
		`INSERT INTO booking.sessions (booking_id, start_photos, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 ON CONFLICT (booking_id)
		 DO UPDATE SET
			start_photos = EXCLUDED.start_photos,
			updated_at = NOW()`,
		bookingID,
		startPhotos,
	); err != nil {
		return err
	}

	_, err := tx.Exec(ctx,
		`UPDATE booking.basebookings
		 SET status = 'ONGOING', updatedat = now()
		 WHERE id = (SELECT base_booking_id FROM booking.bookings WHERE id = $1)`,
		bookingID,
	)
	return err
}

func (t *BookingTasks) EndSession(ctx context.Context, tx pgx.Tx, bookingID string, endPhotos []string) error {
	if _, err := tx.Exec(ctx,
		`INSERT INTO booking.sessions (booking_id, end_photos, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 ON CONFLICT (booking_id)
		 DO UPDATE SET
			end_photos = EXCLUDED.end_photos,
			updated_at = NOW()`,
		bookingID,
		endPhotos,
	); err != nil {
		return err
	}

	_, err := tx.Exec(ctx,
		`UPDATE booking.basebookings
		 SET status = 'COMPLETED', updatedat = now()
		 WHERE id = (SELECT base_booking_id FROM booking.bookings WHERE id = $1)`,
		bookingID,
	)
	return err
}

func (t *BookingTasks) UpdateCleanerStatusesForBooking(ctx context.Context, tx pgx.Tx, bookingID, status string) error {
	_, err := tx.Exec(ctx,
		`UPDATE account.employees e
		 SET status = $1,
		     updated_at = NOW()
		 WHERE e.id = ANY(
		 	COALESCE(
		 		(SELECT b.cleaner_ids FROM booking.bookings b WHERE b.id = $2),
		 		ARRAY[]::uuid[]
		 	)
		 )`,
		status,
		bookingID,
	)
	return err
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

func (t *BookingTasks) FetchUsedInventoryByBooking(ctx context.Context, tx pgx.Tx, bookingID string, itemType string) ([]types.UsedInventoryItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT biu.item_id, COALESCE(i.image_url, ''), biu.quantity_used
		 FROM booking.bookings b
		 JOIN booking.booking_inventory_used biu
		   ON biu.id = ANY(
				CASE
					WHEN $2 = 'RESOURCE' THEN COALESCE(b.resource_ids, ARRAY[]::uuid[])
					WHEN $2 = 'EQUIPMENT' THEN COALESCE(b.equipment_ids, ARRAY[]::uuid[])
					ELSE ARRAY[]::uuid[]
				END
			)
		 JOIN inventory.items i
		   ON i.id = biu.item_id
		 WHERE b.id = $1
		   AND biu.item_type = $2`,
		bookingID,
		itemType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch used inventory for booking: %w", err)
	}
	defer rows.Close()

	items := make([]types.UsedInventoryItem, 0)
	for rows.Next() {
		var item types.UsedInventoryItem
		if scanErr := rows.Scan(&item.ID, &item.ImageURL, &item.Quantity); scanErr != nil {
			return nil, fmt.Errorf("failed to scan used inventory row: %w", scanErr)
		}
		items = append(items, item)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("failed iterating used inventory rows: %w", rows.Err())
	}

	return items, nil
}

func (t *BookingTasks) FetchBookingsToday(ctx context.Context, tx pgx.Tx, logger *utils.Logger) (*types.FetchBookingsTodayResponse, error) {
	var rawJSON []byte

	err := tx.QueryRow(ctx,
		`SELECT booking.get_bookings_today()`,
	).Scan(&rawJSON)
	if err != nil {
		logger.Error("failed to call get_bookings_today sproc: %v", err)
		return nil, fmt.Errorf("failed calling sproc get_bookings_today: %w", err)
	}

	var response types.FetchBookingsTodayResponse
	if err := json.Unmarshal(rawJSON, &response); err != nil {
		logger.Error("failed to unmarshal today's bookings JSON: %v", err)
		return nil, fmt.Errorf("unmarshal today's bookings: %w", err)
	}
	return &response, nil
}
