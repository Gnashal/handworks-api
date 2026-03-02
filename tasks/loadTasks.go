package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"

	"github.com/jackc/pgx/v5"
)

type LoadTasks struct{}

// Make functions public by capitalizing them if they're used elsewhere
// Or comment them out if they're truly unused

func LoadBaseBooking(ctx context.Context, tx pgx.Tx, id string) (*types.BaseBookingDetails, error) {
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

func LoadServiceDetails(ctx context.Context, tx pgx.Tx, svcID string) (*types.ServiceDetails, error) {
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

func LoadAddOns(ctx context.Context, tx pgx.Tx, ids []string) ([]types.AddOns, error) {
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

		svc, err := ParseServiceFromRow(serviceID, svcType, raw)
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

func ParseServiceFromRow(id, svcType string, raw []byte) (*types.ServiceDetails, error) {
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

func LoadEquipments(ctx context.Context, tx pgx.Tx, ids []string) ([]types.CleaningEquipment, error) {
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

func LoadResources(ctx context.Context, tx pgx.Tx, ids []string) ([]types.CleaningResources, error) {
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

func LoadCleaners(ctx context.Context, tx pgx.Tx, ids []string) ([]types.CleanerAssigned, error) {
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
