package realtime

import (
	"context"
	"encoding/json"
	"handworks-api/realtime"
	"handworks-api/services"
	"handworks-api/utils"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func ListenBookingEvents(
	ctx context.Context,
	connString string,
	bookingService *services.BookingService,
	hub *realtime.AdminHub,
	logger *utils.Logger,
) {

	listener := pq.NewListener(
		connString,
		10*time.Second,
		time.Minute,
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				logger.Error("Listener error: %v", err)
			}
		},
	)
	err := listener.Listen("booking_created")
	if err != nil {
		logger.Fatal("Failed to start listening to booking_created: %v", err)
		return
	}

	logger.Info("Listening to booking_created events via lib/pq")

	go func() {
		for {
			select {
			case <-ctx.Done():
				logger.Info("Shutting down booking listener")
				_ = listener.UnlistenAll()
				return
			case n := <-listener.Notify:
				if n == nil {
					continue
				}
				logger.Debug("Listen payload: %s", n.Extra)

				var evt = struct {
					Type      string `json:"type"`
					BookingID string `json:"bookingId"`
				}{}
				if err := json.Unmarshal([]byte(n.Extra), &evt); err != nil {
					logger.Error("Failed to unmarshal booking event: %v", err)
					continue
				}

				booking, err := bookingService.GetBookingByID(ctx, evt.BookingID)
				if err != nil {
					logger.Error("Failed to get booking by ID: %v", err)
					continue
				}

				hub.SendToAdmin("booking.created", booking)
			case <-time.After(90 * time.Second):
				err := listener.Ping()
				if err != nil {
					logger.Error("Listener ping error: %v", err)
				}
			}
		}
	}()
}
