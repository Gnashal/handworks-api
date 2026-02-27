package types

import (
	"encoding/json"
	"time"
)

type Quote struct {
	ID                string          `json:"id"`
	CustomerID        string          `json:"customerId"`
	MainService       string          `json:"mainService"`                            //the main type of service
	MainServiceDetail json.RawMessage `json:"mainServiceDetail" swaggertype:"object"` //added
	MainServiceHours  int32           `json:"mainServiceHours"`                       //added
	Subtotal          float32         `json:"subtotal"`
	AddonTotal        float32         `json:"addonTotal"`
	TotalServiceHours int32           `json:"totalServiceHours"` //added
	TotalPrice        float32         `json:"totalPrice"`
	IsValid           bool            `json:"isValid"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
	Addons            []*QuoteAddon   `json:"addons"`
}

type QuoteAddon struct {
	ID            string          `json:"id"`
	QuoteID       string          `json:"quoteId"`
	ServiceType   string          `json:"serviceType"`
	ServiceDetail json.RawMessage `json:"serviceDetail" swaggertype:"object"` // serialized ServicesRequest
	ServiceHours  int32           `json:"serviceHours"`
	AddonPrice    float32         `json:"addonPrice"`
	CreatedAt     time.Time       `json:"createdAt"`
}
type FetchAllQuotesResponse struct {
	TotalQuotes     int     `json:"totalQuotes"`
	QuotesRequested int     `json:"quotesRequested"`
	Quotes          []Quote `json:"quotes"`
}

type QuoteAddonCleaningPrice struct {
	AddonName  string  `json:"addon_name"`
	AddonPrice float32 `json:"addon_price"`
}
type QuoteCleaningPrices struct {
	MainServicePrice float32              `json:"mainServicePrice"`
	AddonPrices      []AddonCleaningPrice `json:"addonPrices"`
}
type QuoteResponse struct {
	QuoteId           string           `json:"quoteId" db:"quote_id"`
	MainServiceName   string           `json:"mainServiceName"`
	MainServiceDetail json.RawMessage  `json:"mainServiceDetail" swaggertype:"object"`
	MainServiceTotal  float32          `json:"mainServiceTotal"`
	MainServiceHours  int32            `json:"mainServiceHours"`
	AddonTotal        float32          `json:"addonTotal"`
	TotalPrice        float32          `json:"totalPrice"`
	TotalServiceHours int32            `json:"totalServiceHours"`
	Addons            []AddOnBreakdown `json:"addons"`
}

// QuoteRequest represents the data needed to build a quotation.
type QuoteRequest struct {
	CustomerID string          `json:"customerId" db:"customer_id"`
	Service    ServicesRequest `json:"service"` // nested structs usually don't need db tags
	Addons     []AddOnRequest  `json:"addons"`  // same here
}

type AddOnBreakdown struct {
	AddonID       string          `json:"addonId" db:"addon_id"`
	ServiceType   string          `json:"serviceType" db:"service_type"`
	ServiceDetail json.RawMessage `json:"serviceDetail" db:"service_detail" swaggertype:"object"`
	ServiceHours  int32           `json:"serviceHours" db:"service_hours"`
	Price         float64         `json:"price" db:"addon_price"`
}

// CustomerRequest fetches all quotes belonging to a customer.
type CustomerRequest struct {
	CustomerID string `json:"customerId" db:"customer_id"`
}

// QuotesResponse holds a list of quotations for a customer.
type QuotesResponse struct {
	Quotes []QuoteResponse `json:"quotes"`
}

var MattressPrices = map[string]float32{
	"KING":           2000.00,
	"KING_HEADBAND":  2500.00,
	"QUEEN":          1800.00,
	"QUEEN_HEADBAND": 2300.00,
	"DOUBLE":         1500.00,
	"SINGLE":         1000.00,
}
var CarPrices = map[string]float32{
	"SEDAN":     3250.00,
	"MPV":       4000.00,
	"SUV":       4000.00,
	"VAN":       5200.00,
	"PICKUP":    3600.00,
	"CAR_SMALL": 1750.00,
}

var CouchPrices = map[string]float32{
	"SEATER_1":             500.00,
	"SEATER_2":             1000.00,
	"SEATER_3":             1300.00,
	"SEATER_3_LTYPE_SMALL": 1500.00,
	"SEATER_3_LTYPE_LARGE": 1750.00,
	"SEATER_4_LTYPE_SMALL": 1800.00,
	"SEATER_4_LTYPE_LARGE": 2000.00,
	"SEATER_5_LTYPE":       2250.00,
	"SEATER_6_LTYPE":       2500.00,
	"OTTOMAN":              500.00,
	"LAZBOY":               900.00,
	"CHAIR":                250.00,
}
