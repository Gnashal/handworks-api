package types

import "time"


type ItemType string
type ItemStatus string
type ItemCategory string

const (
	// Item Types
	ItemTypeResource  ItemType = "RESOURCE"
	ItemTypeEquipment ItemType = "EQUIPMENT"

	// Item Status
	ItemStatusHigh       ItemStatus = "HIGH"
	ItemStatusLow        ItemStatus = "LOW"
	ItemStatusDanger     ItemStatus = "DANGER"
	ItemStatusOutOfStock ItemStatus = "OUT_OF_STOCK"

	// Item Categories
	CategoryGeneral     ItemCategory = "GENERAL"
	CategoryElectronics ItemCategory = "ELECTRONICS"
	CategoryFurniture   ItemCategory = "FURNITURE"
	CategoryAppliances  ItemCategory = "APPLIANCES"
	CategoryVehicles    ItemCategory = "VEHICLES"
	CategoryOther       ItemCategory = "OTHER"
)



// Db / Resonse Model
type InventoryItem struct {
	ID          string        `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	Type        ItemType      `json:"type" db:"type"`
	Status      ItemStatus    `json:"status" db:"status"`
	Category    ItemCategory  `json:"category" db:"category"`
	Quantity    int32         `json:"quantity" db:"quantity"`
	MaxQuantity int32         `json:"max_quantity" db:"max_quantity"`
	Unit        string        `json:"unit" db:"unit"`
	IsAvailable bool          `json:"is_available" db:"is_available"`
	ImageURL    string        `json:"image_url" db:"image_url"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}


type CreateItemRequest struct {
	Name     string `json:"name" binding:"required"`
	Type     string `json:"type" binding:"required"`     // RESOURCE / EQUIPMENT
	Category string `json:"category" binding:"required"` // GENERAL / ELECTRONICS / ...
	Quantity int32  `json:"quantity" binding:"required"`
	Unit     string `json:"unit" binding:"required"`
	ImageURL string `json:"image_url"`
}

type UpdateItemRequest struct {
	ID 			string `json:"id" binding:"required"`
	Name        string `json:"name"`
	Type        string `json:"type"`        // optional
	Status      string `json:"status"`      // HIGH / LOW / DANGER / OUT_OF_STOCK
	Category    string `json:"category"`    // optional
	Quantity    int32  `json:"quantity"`
	MaxQuantity int32  `json:"max_quantity"`
	Unit        string `json:"unit"`
	ImageURL    string `json:"image_url"`
}

type InventoryListResponse struct {
	TotalItems    int32                  `json:"totalItems"`
	ItemsReturned int32                  `json:"itemsReturned"`
	Items         []*InventoryItem 		`json:"items"`
}


type InventoryFilter struct {
	Type     *string `json:"type,omitempty" form:"type"`
	Status   *string `json:"status,omitempty" form:"status"`
	Category *string `json:"category,omitempty" form:"category"`
	StartDate *string `json:"start_date,omitempty" form:"startDate"`
	EndDate   *string `json:"end_date,omitempty" form:"endDate"`
	Page      *int    `json:"page,omitempty" form:"page"`
	Limit     *int    `json:"limit,omitempty" form:"limit"`
}
