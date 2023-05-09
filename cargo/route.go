package cargo

import (
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/location"
)

// RouteSpecification Contains information about a route: its origin,
// destination and arrival deadline.
type RouteSpecification struct {
	Origin          location.UNLocode
	Destination     location.UNLocode
	ArrivalDeadline time.Time
}

// IsSatisfiedBy checks whether provided itinerary satisfies this
// specification.
func (s RouteSpecification) IsSatisfiedBy(itinerary Itinerary) bool {
	return itinerary.Legs != nil &&
		s.Origin == itinerary.InitialDepartureLocation() &&
		s.Destination == itinerary.FinalArrivalLocation()
}

// RoutingStatus describes status of cargo routing.
type RoutingStatus int

// Valid routing statuses.
const (
	NotRouted RoutingStatus = iota
	Misrouted
	Routed
)

func (s RoutingStatus) String() string {
	switch s {
	case NotRouted:
		return "Not routed"
	case Misrouted:
		return "Misrouted"
	case Routed:
		return "Routed"
	}
	return ""
}

// TransportStatus describes status of cargo transportation.
type TransportStatus int

// Valid transport statuses.
const (
	NotReceived TransportStatus = iota
	InPort
	OnboardCarrier
	Claimed
	Unknown
)

func (s TransportStatus) String() string {
	switch s {
	case NotReceived:
		return "Not received"
	case InPort:
		return "In port"
	case OnboardCarrier:
		return "Onboard carrier"
	case Claimed:
		return "Claimed"
	case Unknown:
		return "Unknown"
	}
	return ""
}
