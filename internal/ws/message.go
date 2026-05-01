package ws

import "encoding/json"

// Message types — Client to Server
const (
	MsgTypeDriverLocation = "DRIVER_LOCATION"
	MsgTypeRideAccept     = "RIDE_ACCEPT"
	MsgTypeRideReject     = "RIDE_REJECT"
	MsgTypePing           = "PING"
	MsgTypeSubscribeRide  = "SUBSCRIBE_RIDE"
)

// Message types — Server to Client
const (
	MsgTypeRideRequest    = "RIDE_REQUEST"
	MsgTypeRideMatched    = "RIDE_MATCHED"
	MsgTypeRideStatus     = "RIDE_STATUS"
	MsgTypeLocationUpdate = "LOCATION_UPDATE"
	MsgTypeETAUpdate      = "ETA_UPDATE"
	MsgTypePong           = "PONG"
	MsgTypeError          = "ERROR"
	MsgTypeConnected      = "CONNECTED"
)

// Message is the envelope for all WebSocket communication
type Message struct {
	Type          string          `json:"type"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	TargetUserID  string          `json:"target_user_id,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	Timestamp     int64           `json:"timestamp"`
}

// LocationPayload is sent by drivers to update their position
type LocationPayload struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Speed   float64 `json:"speed"`
	Heading float64 `json:"heading"`
}

// RideRequestPayload is sent to drivers when a ride is available
type RideRequestPayload struct {
	RideID         string  `json:"ride_id"`
	PassengerName  string  `json:"passenger_name"`
	PickupLat      float64 `json:"pickup_lat"`
	PickupLng      float64 `json:"pickup_lng"`
	PickupAddress  string  `json:"pickup_address"`
	DropoffLat     float64 `json:"dropoff_lat"`
	DropoffLng     float64 `json:"dropoff_lng"`
	DropoffAddress string  `json:"dropoff_address"`
	EstimatedFare  float64 `json:"estimated_fare"`
	VehicleType    string  `json:"vehicle_type"`
	TimeoutSeconds int     `json:"timeout_seconds"`
}

// RideAcceptPayload is sent by driver to accept a ride
type RideAcceptPayload struct {
	RideID string `json:"ride_id"`
}

// RideStatusPayload is sent to both parties on status changes
type RideStatusPayload struct {
	RideID    string  `json:"ride_id"`
	Status    string  `json:"status"`
	DriverLat float64 `json:"driver_lat,omitempty"`
	DriverLng float64 `json:"driver_lng,omitempty"`
}

// ETAPayload is sent to passenger with updated ETA
type ETAPayload struct {
	RideID     string  `json:"ride_id"`
	ETASeconds int     `json:"eta_seconds"`
	DistanceM  float64 `json:"distance_m"`
}
