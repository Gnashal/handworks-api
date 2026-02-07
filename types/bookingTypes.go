package types

import "time"

type BookingAllocation struct {
	CleaningAllocation *CleaningAllocation
	CleanerAssigned    []CleanerAssigned
	CleaningPrices     *CleaningPrices
}
type CleaningAllocation struct {
	CleaningEquipment []CleaningEquipment
	CleaningResources []CleaningResources
}
type CleaningEquipment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	PhotoURL string `json:"photoUrl"`
}

type CleaningResources struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	PhotoURL string `json:"photoUrl"`
}

type CleanerAssigned struct {
	ID               string `json:"id"`
	CleanerFirstName string `json:"cleanerFirstName"`
	CleanerLastName  string `json:"cleanerLastName"`
	PFPUrl           string `json:"pfpUrl"`
}
type AddonCleaningPrice struct {
	AddonName  string  `json:"addonName"`
	AddonPrice float32 `json:"addonPrice"`
}
type CleaningPrices struct {
	MainServicePrice float32              `json:"mainServicePrice"`
	AddonPrices      []AddonCleaningPrice `json:"addonPrices"`
}

type ServiceDetail struct {
	General  *GeneralCleaningDetails  `json:"general,omitempty"`
	Couch    *CouchCleaningDetails    `json:"couch,omitempty"`
	Mattress *MattressCleaningDetails `json:"mattress,omitempty"`
	Car      *CarCleaningDetails      `json:"car,omitempty"`
	Post     *PostConstructionDetails `json:"post,omitempty"`
}

type ServiceDetails struct {
	ID          string `json:"id"`
	ServiceType string `json:"serviceType"`
	Details     any    `json:"details"`
}

// detail factory types
type DetailType string

const (
	ServiceGeneral  DetailType = "GENERAL_CLEANING"
	ServiceCouch    DetailType = "COUCH"
	ServiceMattress DetailType = "MATTRESS"
	ServiceCar      DetailType = "CAR"
	ServicePost     DetailType = "POST"
)

// Used for unmarshaling dynamically
var DetailFactories = map[DetailType]func() any{
	ServiceGeneral:  func() any { return &GeneralCleaningDetails{} },
	ServiceCouch:    func() any { return &CouchCleaningDetails{} },
	ServiceMattress: func() any { return &MattressCleaningDetails{} },
	ServiceCar:      func() any { return &CarCleaningDetails{} },
	ServicePost:     func() any { return &PostConstructionDetails{} },
}

type GeneralCleaningDetails struct {
	HomeType string `json:"homeType"`
	SQM      int32  `json:"sqm"`
	Hours    int32  `json:"hours"`
}

// Couch cleaning
type CouchCleaningSpecifications struct {
	CouchType string `json:"couchType"`
	WidthCM   int32  `json:"widthCm"`
	DepthCM   int32  `json:"depthCm"`
	HeightCM  int32  `json:"heightCm"`
	Quantity  int32  `json:"quantity"`
}

type CouchCleaningDetails struct {
	CleaningSpecs []CouchCleaningSpecifications `json:"cleaningSpecs"`
	BedPillows    int32                         `json:"bedPillows"`
}

// Mattress cleaning
type MattressCleaningSpecifications struct {
	BedType  string `json:"bedType"`
	WidthCM  int32  `json:"widthCm"`
	DepthCM  int32  `json:"depthCm"`
	HeightCM int32  `json:"heightCm"`
	Quantity int32  `json:"quantity"`
}

type MattressCleaningDetails struct {
	CleaningSpecs []MattressCleaningSpecifications `json:"cleaningSpecs"`
}

// Car cleaning
type CarCleaningSpecifications struct {
	CarType  string `json:"carType"`
	Quantity int32  `json:"quantity"`
}

type CarCleaningDetails struct {
	CleaningSpecs []CarCleaningSpecifications `json:"cleaningSpecs"`
	ChildSeats    int32                       `json:"childSeats"`
}
type PostConstructionDetails struct {
	SQM int32 `json:"sqm"`
}
type BaseBookingDetails struct {
	ID                string     `json:"id"`
	CustID            string     `json:"custId"`
	CustomerFirstName string     `json:"customerFirstName"`
	CustomerLastName  string     `json:"customerLastName"`
	Address           Address    `json:"address"`
	StartSched        time.Time  `json:"startSched"`
	EndSched          time.Time  `json:"endSched"`
	DirtyScale        int32      `json:"dirtyScale"`
	PaymentStatus     string     `json:"paymentStatus"`
	ReviewStatus      string     `json:"reviewStatus"`
	Photos            []string   `json:"photos"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	QuoteId           string     `json:"quoteId"`
}
type BaseBookingDetailsRequest struct {
	CustID            string     `json:"custId"`
	CustomerFirstName string     `json:"customerFirstName"`
	CustomerLastName  string     `json:"customerLastName"`
	Address           Address    `json:"address"`
	StartSched        time.Time  `json:"startSched"`
	EndSched          time.Time  `json:"endSched"`
	DirtyScale        int32      `json:"dirtyScale"`
	Photos            []string   `json:"photos"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         *time.Time `json:"updatedAt,omitempty"`
	QuoteId           string     `json:"quoteId"`
}
type Address struct {
	AddressHuman string  `json:"addressHuman"`
	AddressLat   float64 `json:"addressLat"`
	AddressLng   float64 `json:"addressLng"`
}
type BookingReply struct {
	Source     string              `json:"source"`
	Equipments []CleaningEquipment `json:"equipments,omitempty"`
	Resources  []CleaningResources `json:"resources,omitempty"`
	Cleaners   []CleanerAssigned   `json:"cleaners,omitempty"`
	Prices     CleaningPrices      `json:"prices,omitempty"`
	Error      string              `json:"error,omitempty"`
}

type MainServiceType string

const (
	ServiceTypeUnspecified MainServiceType = "SERVICE_TYPE_UNSPECIFIED"
	GeneralCleaning        MainServiceType = "GENERAL_CLEANING"
	CouchCleaning          MainServiceType = "COUCH"
	MattressCleaning       MainServiceType = "MATTRESS"
	CarCleaning            MainServiceType = "CAR"
	PostCleaning           MainServiceType = "POST"
)

type ServicesRequest struct {
	ServiceType MainServiceType `json:"serviceType"`
	Details     ServiceDetail   `json:"details"`
}
type AddOnRequest struct {
	ServiceDetail ServicesRequest `json:"serviceDetail"`
}

type CreateBookingRequest struct {
	AccountID   string                    `json:"accountId"`
	Base        BaseBookingDetailsRequest `json:"base"`
	MainService ServicesRequest           `json:"mainService"`
	Addons      []AddOnRequest            `json:"addons"`
}
type AddOns struct {
	ID            string         `json:"id"`
	ServiceDetail ServiceDetails `json:"serviceDetail"`
	Price         float32        `json:"price"`
}
type Booking struct {
	ID          string              `json:"id"`
	Base        BaseBookingDetails  `json:"base"`
	MainService ServiceDetails      `json:"mainService"`
	Addons      []AddOns            `json:"addons,omitempty"`
	Equipments  []CleaningEquipment `json:"equipments,omitempty"`
	Resources   []CleaningResources `json:"resources,omitempty"`
	Cleaners    []CleanerAssigned   `json:"cleaners,omitempty"`
	TotalPrice  float32             `json:"totalPrice"`
}

type FetchAllBookingsResponse struct {
	TotalBookings     int       `json:"totalBookings"`
	BookingsRequested int       `json:"bookingsRequested"`
	Bookings          []Booking `json:"bookings"`
}
