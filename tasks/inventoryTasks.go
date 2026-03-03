package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"handworks-api/types"

	"github.com/jackc/pgx/v5"
)

type InventoryTasks struct{}

func (t *InventoryTasks) CreateInventoryItem(
	c context.Context,
	tx pgx.Tx,
	name, itemType, unit, category, imageUrl string,
	quantity, maxQuantity int32,
) (*types.InventoryItem, error) {
	var item types.InventoryItem

	if err := tx.QueryRow(c,
		`INSERT INTO inventory.items
		 (name, type, unit, quantity, max_quantity, category ,image_url, is_available)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, true)
		 RETURNING id, name, type, status, unit, category, quantity, max_quantity, image_url, is_available, created_at, updated_at`,
		name, itemType, unit, quantity, maxQuantity, category, imageUrl,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Status,
		&item.Unit,
		&item.Category,
		&item.Quantity,
		&item.MaxQuantity,
		&item.ImageURL,
		&item.IsAvailable,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("could not create inventory item: %w", err)
	}

	return &item, nil
}
func (t *InventoryTasks) FetchInventoryItem(
	c context.Context,
	tx pgx.Tx,
	id string,
) (*types.InventoryItem, error) {
	var item types.InventoryItem

	if err := tx.QueryRow(c,
		`SELECT id, name, type, status, unit, quantity, max_quantity, is_available, created_at, updated_at
		 FROM inventory.items
		 WHERE id = $1`,
		id,
	).Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Status,
		&item.Unit,
		&item.Quantity,
		&item.MaxQuantity,
		&item.IsAvailable,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("could not fetch inventory item with id %s: %w", id, err)
	}

	return &item, nil
}
func (t *InventoryTasks) FetchItems(ctx context.Context, tx pgx.Tx, filter *types.InventoryFilter) (*types.InventoryListResponse, error) {
	var raw json.RawMessage

	err := tx.QueryRow(
		ctx,
		`SELECT inventory.get_items(
			$1, $2, $3, $4, $5
		)`,
		filter.Type,
		filter.Status,
		filter.Category,
		*filter.Page,
		*filter.Limit,
	).Scan(&raw)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch inventory: %w", err)
	}

	var resp types.InventoryListResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inventory response: %w", err)
	}

	return &resp, nil
}

func (t *InventoryTasks) UpdateInventoryItem(
	ctx context.Context,
	tx pgx.Tx,
	in *types.UpdateItemRequest,
) (*types.InventoryItem, error) {
	args := pgx.NamedArgs{
		"id":           in.ID,
		"name":         in.Name,
		"type":         in.Type,
		"status":       in.Status,
		"category":     in.Category,
		"quantity":     in.Quantity,
		"max_quantity": in.MaxQuantity,
	}

	row := tx.QueryRow(ctx, `
		UPDATE inventory.items
		SET
			name = COALESCE(NULLIF(@name, ''), name),
			type = COALESCE(NULLIF(@type, ''), type),
			status = COALESCE(NULLIF(@status, ''), status),
			category = COALESCE(NULLIF(@category, ''), category),
			quantity = COALESCE(NULLIF(@quantity, '')::int, quantity),
			max_quantity = COALESCE(NULLIF(@max_quantity, '')::int, max_quantity),
			updated_at = NOW()
		WHERE id = @id
		RETURNING id, name, type, status, unit, category, quantity, max_quantity, is_available, created_at, updated_at
	`, args)

	var item types.InventoryItem
	if err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Status,
		&item.Unit,
		&item.Category,
		&item.Quantity,
		&item.MaxQuantity,
		&item.IsAvailable,
		&item.CreatedAt,
		&item.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("could not update inventory item: %w", err)
	}

	return &item, nil
}
func (t *InventoryTasks) DeleteInventoryItem(
	ctx context.Context,
	tx pgx.Tx,
	id string,
) (*types.InventoryItem, error) {
	var item types.InventoryItem

	err := tx.QueryRow(ctx, `
		DELETE FROM inventory.items
		WHERE id = $1
		RETURNING id, name, type, status, unit, category, quantity, max_quantity, is_available, created_at, updated_at
	`, id).Scan(
		&item.ID,
		&item.Name,
		&item.Type,
		&item.Status,
		&item.Unit,
		&item.Category,
		&item.Quantity,
		&item.MaxQuantity,
		&item.IsAvailable,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("could not delete inventory item with id %s: %w", id, err)
	}

	return &item, nil
}

// for the assignment logic
// func (s *InventoryTasks) resolveEquipmentTypes(serviceType string) []string {
// 	switch serviceType {
// 	case "general":
// 		return []string{"Vacuum Cleaner", "Mop"}
// 	case "car":
// 		return []string{"Car wax", "Steam Cleaner"}
// 	default:
// 		return []string{}
// 	}
// }
