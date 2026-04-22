package services

import (
	"context"
	"errors"
	"fmt"
	"handworks-api/tasks"
	"handworks-api/types"

	"github.com/jackc/pgx/v5"
)

func (s *AdminService) withTx(
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

func (s *AdminService) GetAdminDashboard(ctx context.Context, req *types.AdminDashboardRequest) (*types.AdminDashboardResponse, error) {
	var res *types.AdminDashboardResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		res, err = s.Tasks.FetchAdminDashboardData(ctx, tx, s.Logger, req.DateFilter)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch admin dashboard analytics: %v", err)
		return nil, err
	}

	return res, nil
}

func (s *AdminService) OnboardEmployee(ctx context.Context, req *types.OnboardEmployeeRequest) (*types.SignUpEmployeeResponse, error) {
	var emp *types.SignUpEmployeeResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		clerkUser, err := s.Tasks.CreateClerkUser(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to create clerk user: %w", err)
		}
		newEmp := &types.SignUpEmployeeRequest{
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Role:      req.Role,
			Provider:  "email/password",
			ClerkID:   clerkUser.ID,
			Position:  req.Position,
			HireDate:  req.HireDate,
		}
		emp, err = s.AccountPort.SignUpEmployee(ctx, *newEmp)
		if err != nil {
			return fmt.Errorf("failed to onboard employee: %w", err)
		}
		return err
	}); err != nil {
		s.Logger.Error("Failed to onboard employee: %v", err)
		return nil, err
	}

	return emp, nil
}

func (s *AdminService) AcceptBooking(ctx context.Context, bookingId string) (*types.AcceptBookingResponse, error) {
	s.Logger.Info("Accepting booking with ID: %s", bookingId)
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		return s.Tasks.AcceptBooking(ctx, tx, bookingId)
	}); err != nil {
		s.Logger.Error("Failed to accept booking: %v", err)
		return nil, err
	}

	return &types.AcceptBookingResponse{
		BookingID: bookingId,
		Status:    "SCHEDULED",
	}, nil
}

func (s *AdminService) AssignResourcesToBooking(ctx context.Context, req *types.AssignResourcesToBookingRequest) (*types.AssignInventoryResponse, error) {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		return s.Tasks.AssignResourcesToBooking(ctx, tx, req.BookingID, req.Resources)
	}); err != nil {
		s.Logger.Error("Failed to assign resources to booking: %v", err)
		return nil, err
	}

	return &types.AssignInventoryResponse{
		BookingID: req.BookingID,
		Message:   "Resources assigned successfully",
	}, nil
}

func (s *AdminService) AssignEquipmentToBooking(ctx context.Context, req *types.AssignEquipmentToBookingRequest) (*types.AssignInventoryResponse, error) {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		return s.Tasks.AssignEquipmentToBooking(ctx, tx, req.BookingID, req.Equipment)
	}); err != nil {
		s.Logger.Error("Failed to assign equipment to booking: %v", err)
		return nil, err
	}

	return &types.AssignInventoryResponse{
		BookingID: req.BookingID,
		Message:   "Equipment assigned successfully",
	}, nil
}

func (s *AdminService) GetCalendarBookings(ctx context.Context, month string) (*types.CalendarBookingResponse, error) {
	var res *types.CalendarBookingResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		res, err = s.Tasks.FetchCalendarBookings(ctx, tx, month)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch calendar bookings: %v", err)
		return nil, err
	}

	return res, nil
}
func (s *AdminService) GetBookingTrends(ctx context.Context) (*types.BookingTrendsResponse, error) {
	var res *types.BookingTrendsResponse

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		res, err = s.Tasks.FetchBookingTrends(ctx, tx)
		return err
	}); err != nil {
		s.Logger.Error("Failed to fetch booking trends: %v", err)
		return nil, err
	}

	return res, nil
}

func (s *AdminService) GetAvailableCleaners(ctx context.Context, bookingId string) (*types.AvailableCleanersResponse, error) {
	var (
		window   *tasks.BookingScheduleWindow
		cleaners []types.AvailableCleaner
	)

	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		var err error
		window, err = s.Tasks.GetBookingScheduleWindow(ctx, tx, bookingId)
		if err != nil {
			return err
		}

		cleaners, err = s.Tasks.FetchAvailableCleanersByBooking(ctx, tx, bookingId, window.StartSched, window.EndSched)
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		s.Logger.Error("Failed to fetch available cleaners: %v", err)
		return nil, err
	}

	return &types.AvailableCleanersResponse{
		BookingID:    bookingId,
		StartSched:   window.StartSched,
		EndSched:     window.EndSched,
		Cleaners:     cleaners,
		CleanerCount: len(cleaners),
	}, nil
}

func (s *AdminService) AssignEmployeeToBooking(ctx context.Context, req *types.AssignEmployeeToBookingRequest) (*types.AssignEmployeeToBookingResponse, error) {
	if err := s.withTx(ctx, func(tx pgx.Tx) error {
		window, err := s.Tasks.GetBookingScheduleWindow(ctx, tx, req.BookingID)
		if err != nil {
			return err
		}

		if err := s.Tasks.ValidateActiveCleaner(ctx, tx, req.EmployeeID); err != nil {
			return err
		}

		if req.Action == types.AssignEmployeeActionAdd {
			hasConflict, err := s.Tasks.CleanerHasScheduleConflict(ctx, tx, req.EmployeeID, window.StartSched, window.EndSched, req.BookingID)
			if err != nil {
				return err
			}
			if hasConflict {
				return tasks.ErrCleanerHasConflict
			}
		}

		return s.Tasks.AssignEmployeeToBooking(ctx, tx, req.BookingID, req.EmployeeID, req.Action)
	}); err != nil {
		if errors.Is(err, tasks.ErrBookingNotFound) ||
			errors.Is(err, tasks.ErrEmployeeNotFoundOrInactive) ||
			errors.Is(err, tasks.ErrCleanerAlreadyAssigned) ||
			errors.Is(err, tasks.ErrCleanerNotAssigned) ||
			errors.Is(err, tasks.ErrCleanerHasConflict) {
			return nil, err
		}

		s.Logger.Error("Failed to assign employee to booking: %v", err)
		return nil, err
	}

	actionText := string(req.Action)
	message := "Cleaner assigned to booking successfully"
	if req.Action == types.AssignEmployeeActionRemove {
		actionText = string(types.AssignEmployeeActionRemove)
		message = "Cleaner removed from booking successfully"
	}

	return &types.AssignEmployeeToBookingResponse{
		BookingID:  req.BookingID,
		EmployeeID: req.EmployeeID,
		Action:     actionText,
		Message:    message,
	}, nil
}
