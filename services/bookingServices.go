package services

import (
	"context"
	"fmt"
	"handworks-api/types"

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

	// ✅ Validate once only
	if req.ExtraHours > 0 {
		if req.MainService.ServiceType != types.GeneralCleaning {
			return nil, fmt.Errorf("extra hours can only be added for General Cleaning services")
		}
		if req.ExtraHours > 4 {
			return nil, fmt.Errorf("extra hours cannot exceed 4 hours")
		}
	}

	// ✅ AllocateAll is the single source of truth for EndSched, extraHourCost, originalEndSched
	alloc, err := s.Tasks.AllocateAll(ctx, s.PaymentPort, &req)
	if err != nil {
		s.Logger.Error("Allocation failed: %v", err)
		return nil, err
	}

	s.Logger.Debug("=== SCHEDULE DEBUG ===")
	s.Logger.Debug("StartSched:        %v", req.Base.StartSched)
	s.Logger.Debug("TotalServiceHours: %v", req.TotalServiceHours)
	s.Logger.Debug("ExtraHours:        %v", req.ExtraHours)
	s.Logger.Debug("EndSched after AllocateAll: %v", req.Base.EndSched)
	s.Logger.Debug("OriginalEndSched:  %v", alloc.OriginalEndSched)
	s.Logger.Debug("======================")

	var createdBooking *types.Booking

	err = s.withTx(ctx, func(tx pgx.Tx) error {
		mainService, err := s.Tasks.CreateMainServiceBooking(ctx, tx, s.Logger, req.MainService.Details)
		if err != nil {
			return err
		}

		// ✅ Use values directly from alloc — no recomputation, no EndSched mutation
		extraHourCost := alloc.ExtraHourCost
		originalEndSched := alloc.OriginalEndSched

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
			req.Base.QuoteId,
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
			for _, ap := range alloc.CleaningPrices.AddonPrices {
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

		equipmentIDs := make([]string, 0, len(alloc.CleaningAllocation.CleaningEquipment))
		for _, eq := range alloc.CleaningAllocation.CleaningEquipment {
			equipmentIDs = append(equipmentIDs, eq.ID)
		}

		resourceIDs := make([]string, 0, len(alloc.CleaningAllocation.CleaningResources))
		for _, r := range alloc.CleaningAllocation.CleaningResources {
			resourceIDs = append(resourceIDs, r.ID)
		}

		cleanerIDs := make([]string, 0, len(alloc.CleanerAssigned))
		for _, c := range alloc.CleanerAssigned {
			cleanerIDs = append(cleanerIDs, c.ID)
		}

		// ✅ Subtract extraHourCost to get base quote price (AllocateAll already added it to MainServicePrice)
		originalQuotePrice := alloc.CleaningPrices.MainServicePrice - extraHourCost

		bookingID, err := s.Tasks.SaveBooking(
			ctx, tx,
			baseBook.ID, mainService.ID,
			addonIDs, equipmentIDs, resourceIDs, cleanerIDs,
			originalQuotePrice,
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
			Equipments:  alloc.CleaningAllocation.CleaningEquipment,
			Resources:   alloc.CleaningAllocation.CleaningResources,
			Cleaners:    alloc.CleanerAssigned,
			TotalPrice:  originalQuotePrice + extraHourCost,
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
	customerId, startDate, endDate string,
	page, limit int,
) (*types.FetchAllBookingsResponse, error) {
	var result *types.FetchAllBookingsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		result, err = s.Tasks.FetchAllCustomerBookings(ctx, tx, customerId, startDate, endDate, page, limit, s.Logger)
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

func (s *BookingService) UpdateBooking(ctx context.Context) error {
	return nil
}

func (s *BookingService) DeleteBooking(ctx context.Context) error {
	return nil
}
