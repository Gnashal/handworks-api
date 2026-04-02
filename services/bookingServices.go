package services

import (
	"context"
	"fmt"
	"handworks-api/types"
	"time"

	"github.com/jackc/pgx/v5"
)

func (s *BookingService) withTx(
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

func (s *BookingService) CreateBooking(ctx context.Context, req types.CreateBookingRequest) (*types.Booking, error) {
	s.Logger.Info("Creating booking for customer: %s...", req.Base.CustomerFirstName)

	if req.ExtraHours > 0 {
		if req.MainService.ServiceType != types.GeneralCleaning {
			return nil, fmt.Errorf("extra hours can only be added for General Cleaning services")
		}
		if req.ExtraHours > 4 {
			return nil, fmt.Errorf("extra hours cannot exceed 4 hours")
		}
	}

	order, prices, err := s.Tasks.FetchOrderAndPrices(ctx, s.PaymentPort, req.Base.OrderId)
	if err != nil {
		s.Logger.Error("Failed to fetch order and prices: %v", err)
		return nil, err
	}
	var createdBooking *types.Booking

	err = s.withTx(ctx, func(tx pgx.Tx) error {

		cleaners, err := s.Tasks.AllocateCleaners(ctx, tx)
		if err != nil {
			return err
		}

		allocation, err := s.Tasks.AllocateEquipmentAndResources(ctx, tx, &req)
		if err != nil {
			return err
		}

		mainService, err := s.Tasks.CreateMainServiceBooking(ctx, tx, s.Logger, req.MainService.Details)
		if err != nil {
			return err
		}

		originalEndSched := req.Base.EndSched
		var extraHourCost float32

		if req.ExtraHours > 0 && len(cleaners) > 0 {
			extraHourCost = req.ExtraHours * 250.00 * float32(len(cleaners))
			req.Base.EndSched = req.Base.EndSched.Add(time.Duration(req.ExtraHours * float32(time.Hour)))
		}

		baseBook, err := s.Tasks.MakeBaseBooking(
			ctx,
			tx,
			req.Base.CustID,
			req.Base.CustomerFirstName,
			req.Base.CustomerLastName,
			req.Base.CustomerPhoneNo,
			req.Base.Address,
			req.Base.StartSched,
			req.Base.EndSched, // ✅ set correctly by AllocateAll
			req.Base.DirtyScale,
			req.Base.Photos,
			order.ID,
			req.ExtraHours,
			extraHourCost,
			&originalEndSched,
		)
		if err != nil {
			return err
		}

		var addonModels []types.AddOns
		var addonIDs []string
		for _, addonReq := range req.Addons {
			var addonPrice float32
			for _, ap := range prices.AddonPrices {
				if ap.AddonName == string(addonReq.ServiceDetail.ServiceType) {
					addonPrice = ap.AddonPrice
					break
				}
			}
			createdAddon, err := s.Tasks.CreateAddOn(ctx, tx, s.Logger, addonReq, addonPrice)
			if err != nil {
				return err
			}
			addonModels = append(addonModels, *createdAddon)
			addonIDs = append(addonIDs, createdAddon.ID)
		}

		equipmentIDs := make([]string, 0, len(allocation.CleaningEquipment))
		for _, eq := range allocation.CleaningEquipment {
			equipmentIDs = append(equipmentIDs, eq.ID)
		}

		resourceIDs := make([]string, 0, len(allocation.CleaningResources))
		for _, r := range allocation.CleaningResources {
			resourceIDs = append(resourceIDs, r.ID)
		}

		cleanerIDs := make([]string, 0, len(cleaners))
		for _, c := range cleaners {
			cleanerIDs = append(cleanerIDs, c.ID)
		}

		bookingID, err := s.Tasks.SaveBooking(
			ctx,
			tx,
			baseBook.ID,
			mainService.ID,
			addonIDs,
			equipmentIDs,
			resourceIDs,
			cleanerIDs,
			prices.MainServicePrice,
			extraHourCost,
		)
		if err != nil {
			return err
		}

		createdBooking = &types.Booking{
			ID:          bookingID,
			Base:        *baseBook,
			MainService: *mainService,
			Addons:      addonModels,
			Equipments:  allocation.CleaningEquipment,
			Resources:   allocation.CleaningResources,
			Cleaners:    cleaners,
			TotalPrice:  prices.MainServicePrice + extraHourCost,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdBooking, nil
}

func (s *BookingService) GetBookings(
	ctx context.Context,
	startDate, endDate string,
	page, limit int,
) (*types.FetchAllBookingsResponse, error) {
	var result *types.FetchAllBookingsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllBookings(ctx, tx, startDate, endDate, page, limit, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch Bookings: %v", err)
		return nil, err
	}

	return result, nil
}

func (s *BookingService) GetCustomerBookings(
	ctx context.Context,
	customerId, startDate, endDate, status string,
	page, limit int,
) (*types.FetchAllBookingsResponse, error) {
	var result *types.FetchAllBookingsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllCustomerBookings(ctx, tx, customerId, startDate, endDate, status, page, limit, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("failed to fetch customer bookings: %v", err)
		return nil, err
	}

	return result, nil
}

func (s *BookingService) GetEmployeeAssignedBookings(
	ctx context.Context,
	employeeId, startDate, endDate string,
	page, limit int,
) (*types.FetchAllBookingsResponse, error) {
	var result *types.FetchAllBookingsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllEmployeeAssignedBookings(ctx, tx, employeeId, startDate, endDate, page, limit, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("failed to fetch employee assigned bookings: %v", err)
		return nil, err
	}

	return result, nil
}

func (s *BookingService) GetBookingByID(ctx context.Context, bookingID string) (*types.Booking, error) {
	var booking *types.Booking

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		booking, err = s.Tasks.FetchBookingByID(ctx, tx, bookingID, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("failed to fetch booking by ID: %v", err)
		return nil, err
	}

	return booking, nil
}

func (s *BookingService) GetBookedSlots(ctx context.Context, date string) (*types.FetchSlotsResponse, error) {
	var result *types.FetchSlotsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchBookingSlots(ctx, tx, date, s.Logger)
		return err
	}); err != nil {
		s.Logger.Error("failed to fetch booked slots: %v", err)
		return nil, fmt.Errorf("failed to get booked slots: %w", err)
	}

	return result, nil
}

func (s *BookingService) StartSession(ctx context.Context, bookingID string) error {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		canStart, startSched, err := s.Tasks.ValidateSessionStart(ctx, tx, bookingID)
		if err != nil {
			return err
		}
		if !canStart {
			return fmt.Errorf("booking cannot be started in its current state")
		}

		now := time.Now()
		startSchedLocal := startSched.In(now.Location())
		nowYear, nowMonth, nowDay := now.Date()
		startYear, startMonth, startDay := startSchedLocal.Date()
		if nowYear != startYear || nowMonth != startMonth || nowDay != startDay {
			return fmt.Errorf("session can only be started within today's timeframe")
		}

		return s.Tasks.StartSession(ctx, tx, bookingID)
	}); err != nil {
		s.Logger.Error("failed to start session for booking %s: %v", bookingID, err)
		return err
	}
	return nil
}

func (s *BookingService) EndSession(ctx context.Context, bookingID string) error {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		return s.Tasks.EndSession(ctx, tx, bookingID)
	}); err != nil {
		s.Logger.Error("failed to end session for booking %s: %v", bookingID, err)
		return err
	}
	return nil
}

func (s *BookingService) GetBookingsToday(ctx context.Context) (types.FetchBookingsTodayResponse, error) {
	var bookings types.FetchBookingsTodayResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		result, err := s.Tasks.FetchBookingsToday(ctx, tx, s.Logger)
		if err != nil {
			return err
		}
		bookings = *result
		return nil
	}); err != nil {
		s.Logger.Error("failed to fetch today's bookings: %v", err)
		return types.FetchBookingsTodayResponse{}, fmt.Errorf("failed to get today's bookings: %w", err)
	}
	return bookings, nil

}

func (s *BookingService) UpdateBooking(ctx context.Context) error {
	return nil
}

func (s *BookingService) DeleteBooking(ctx context.Context) error {
	return nil
}
