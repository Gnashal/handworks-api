package realtime

import (
	"context"
	"encoding/json"
	"handworks-api/services"
	"handworks-api/types"
	"handworks-api/utils"

	"github.com/jackc/pgx/v5"
)

func ListenBookingEvents(
	ctx context.Context,
	dbUrl string,
	bookingService *services.BookingService,
	hub *AdminHub,
	logger *utils.Logger,
) {
	conn, err := pgx.Connect(ctx, dbUrl)
	if err != nil {
		logger.Fatal("Failed to connect to database: %s", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "LISTEN booking_created")
	if err != nil {
		logger.Fatal("Failed to listen for booking_created events: %s", err)
	}

	for {
		n, err := conn.WaitForNotification(ctx)
		if err != nil {
			logger.Error("Failed to listen for booking_created notif: %s",err)
			continue
		}

		var evt types.BookingEvent

		if err := json.Unmarshal([]byte(n.Payload), &evt); err != nil {
			continue
		}

		booking, err := bookingService.GetBookingByID(ctx, evt.BookingID)
		if err != nil {
			logger.Error("Failed to get booking by ID: %s", err)
			continue
		}

		hub.Broadcast("booking.created", booking)
	}
}

