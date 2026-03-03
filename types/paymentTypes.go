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

// --- Order Types ---
type Order struct {
	ID          string `db:"id" json:"id"`
	OrderNumber string `db:"order_number" json:"order_number"`
	CustomerID  string `db:"customer_id" json:"customer_id"`
	QuoteID     string `db:"quote_id" json:"quote_id"`

	Currency string `db:"currency" json:"currency"`

	Subtotal    float32 `db:"subtotal" json:"subtotal"`
	AddonTotal  float32 `db:"addon_total" json:"addon_total"`
	TotalAmount float32 `db:"total_amount" json:"total_amount"`

	DownpaymentRequired float32 `db:"downpayment_required" json:"downpayment_required"`
	RemainingBalance    float32 `db:"remaining_balance" json:"remaining_balance"`

	PaymentStatus string `db:"payment_status" json:"payment_status"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type CreateOrderRequest struct {
	QuoteID     string  `json:"quoteId" binding:"required"`
	CustomerID  string  `json:"customerId" binding:"required"`
	Subtotal    float32 `json:"subtotal" binding:"required"`
	AddonTotal  float32 `json:"addonTotal" binding:"required"`
	TotalAmount float32 `json:"totalAmount" binding:"required"`
}
type CreateOrderResponse struct {
	OrderID     string `json:"orderId"`
	OrderNumber string `json:"orderNumber"`
}
type GetOrdersResponse struct {
	OrdersRequested int     `json:"ordersRequested"`
	TotalOrders     int     `json:"totalOrders"`
	Orders          []Order `json:"orders"`
}

// --- Payment Types ---
type Payment struct {
	ID      string `db:"id" json:"id"`
	OrderID string `db:"order_id" json:"order_id"`

	Type     string `db:"type" json:"type"`         // DOWNPAYMENT | FULLPAYMENT |BALANCE | REFUND
	Provider string `db:"provider" json:"provider"` // PAYMONGO | CASH | MANUAL

	PaymentIntentID *string `db:"payment_intent_id" json:"payment_intent_id,omitempty"`
	PaymentID       *string `db:"payment_id" json:"payment_id,omitempty"`
	PaymentMethodID *string `db:"payment_method_id" json:"payment_method_id,omitempty"`

	Amount   float32 `db:"amount" json:"amount"`
	Currency string  `db:"currency" json:"currency"`

	Status string `db:"status" json:"status"`

	PaidAt       *time.Time `db:"paid_at" json:"paid_at,omitempty"`
	FailedReason *string    `db:"failed_reason" json:"failed_reason,omitempty"`

	RawResponse []byte `db:"raw_response" json:"-"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type GetPaymentsResponse struct {
	PaymentsRequested int       `json:"paymentsRequested"`
	TotalPayments     int       `json:"totalPayments"`
	Payments          []Payment `json:"payments"`
}

// --- Paymongo Types ---
type PaymentIntentResponse struct {
	Data PaymentIntentData `json:"data"`
}

type PaymentIntentData struct {
	ID         string                  `json:"id"`
	Type       string                  `json:"type"`
	Attributes PaymentIntentAttributes `json:"attributes"`
}

type PaymentIntentAttributes struct {
	Amount               int64                             `json:"amount"`
	Currency             string                            `json:"currency"`
	Description          string                            `json:"description"`
	StatementDescriptor  string                            `json:"statement_descriptor"`
	Status               string                            `json:"status"`
	Livemode             bool                              `json:"livemode"`
	ClientKey            string                            `json:"client_key"`
	CreatedAt            int64                             `json:"created_at"`
	UpdatedAt            int64                             `json:"updated_at"`
	LastPaymentError     any                               `json:"last_payment_error"` // null or object
	PaymentMethodAllowed []string                          `json:"payment_method_allowed"`
	Payments             []any                             `json:"payments"`    // empty or payment objects
	NextAction           any                               `json:"next_action"` // null or next action object
	PaymentMethodOptions PaymentIntentPaymentMethodOptions `json:"payment_method_options"`
	Metadata             map[string]string                 `json:"metadata"`
}

type PaymentIntentPaymentMethodOptions struct {
	Card PaymentIntentCardOptions `json:"card"`
}

type PaymentIntentCardOptions struct {
	RequestThreeDSecure string `json:"request_three_d_secure"`
}
