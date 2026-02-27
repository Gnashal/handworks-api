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

type Listener struct {
	ctx              context.Context
	log              *utils.Logger
	hub              *realtime.RealtimeHubs
	listener         *pq.Listener
	bookingService   *services.BookingService
	inventoryService *services.InventoryService
}

func NewListener(
	ctx context.Context,
	log *utils.Logger,
	hub *realtime.RealtimeHubs,
	connString string,
	bookingService *services.BookingService,
	inventoryService *services.InventoryService) *Listener {
	return &Listener{
		ctx: ctx,
		log: log,
		hub: hub,
		listener: pq.NewListener(
			connString,
			10*time.Second,
			time.Minute,
			func(ev pq.ListenerEventType, err error) {
				if err != nil {
					log.Error("Listener error: %v", err)
				}
			},
		),
		bookingService:   bookingService,
		inventoryService: inventoryService,
	}
}
func (l *Listener) Start() error {
	if err := l.listener.Listen("booking_created"); err != nil {
		return err
	}
	if err := l.listener.Listen("booking_accepted"); err != nil {
		return err
	}
	if err := l.listener.Listen("inventory_low"); err != nil {
		return err
	}

	l.log.Info("Started listening to events")

	go func() {
		for {
			select {
			case <-l.ctx.Done():
				l.log.Info("Shutting down PG listener")
				_ = l.listener.UnlistenAll()
				return

			case n := <-l.listener.Notify:
				if n == nil {
					continue
				}

				l.dispatch(n.Channel, n.Extra)

			case <-time.After(90 * time.Second):
				if err := l.listener.Ping(); err != nil {
					l.log.Error("Listener ping error: %v", err)
				}
			}
		}
	}()

	return nil
}
func (l *Listener) dispatch(channel string, payload string) {
	switch channel {

	case "booking_created":
		l.handleBookingCreated(payload)
	case "booking_accepted":
		l.handleBookingAccepted(payload)
	case "inventory_low":
		l.handleInventoryLow(payload)
	default:
		l.log.Warn("Unhandled channel: %s", channel)
	}
}
func (l *Listener) handleBookingAccepted(payload string) {
	l.log.Debug("booking_accepted payload: %s", payload)

	var evt = struct {
		Event      string   `json:"event"`
		BookingID  string   `json:"bookingId"`
		CleanerIDs []string `json:"cleanerIds"`
	}{}

	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		l.log.Error("Invalid booking_accepted payload: %v", err)
		return
	}

	booking, err := l.bookingService.GetBookingByID(l.ctx, evt.BookingID)
	if err != nil {
		l.log.Error("Failed to fetch booking: %v", err)
		return
	}

	for _, cleanerID := range evt.CleanerIDs {
		l.hub.EmployeeHub.SendToEmployee(cleanerID, "booking.accepted", booking)
	}
}
func (l *Listener) handleBookingCreated(payload string) {
	l.log.Debug("booking_created payload: %s", payload)
	var evt = struct {
		Event     string `json:"event"`
		BookingID string `json:"bookingId"`
	}{}
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		l.log.Error("Failed to unmarshal booking event: %v", err)
		return
	}
	booking, err := l.bookingService.GetBookingByID(l.ctx, evt.BookingID)
	if err != nil {
		l.log.Error("Failed to get booking by ID: %v", err)
		return
	}

	l.hub.AdminHub.SendToAdmin("booking.created", booking)
}

func (l *Listener) handleInventoryLow(payload string) {
	l.log.Debug("inventory_low payload: %s", payload)
	var evt = struct {
		Event  string `json:"event"`
		ItemID string `json:"itemId"`
	}{}
	if err := json.Unmarshal([]byte(payload), &evt); err != nil {
		l.log.Error("Failed to unmarshal inventory event: %v", err)
		return
	}
	inventory, err := l.inventoryService.GetItem(l.ctx, evt.ItemID)
	if err != nil {
		l.log.Error("Failed to get inventory item by ID: %v", err)
		return
	}
	l.hub.AdminHub.SendToAdmin("inventory.low", inventory)

}
