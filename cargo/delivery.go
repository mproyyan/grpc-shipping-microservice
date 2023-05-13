package cargo

import (
	"context"
	"database/sql"
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/db"
	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/mproyyan/grpc-shipping-microservice/voyage"
)

// Delivery is the actual transportation of the cargo, as opposed to the
// customer requirement (RouteSpecification) and the plan (Itinerary).
type Delivery struct {
	ID                      int64
	Itinerary               Itinerary
	RouteSpecification      RouteSpecification
	RoutingStatus           RoutingStatus
	TransportStatus         TransportStatus
	NextExpectedActivity    HandlingActivity
	LastEvent               HandlingEvent
	LastKnownLocation       location.UNLocode
	CurrentVoyage           voyage.Number
	ETA                     time.Time
	IsMisdirected           bool
	IsUnloadedAtDestination bool
}

// UpdateOnRouting creates a new delivery snapshot to reflect changes in
// routing, i.e. when the route specification or the itinerary has changed but
// no additional handling of the cargo has been performed.
func (d Delivery) UpdateOnRouting(rs RouteSpecification, itinerary Itinerary) Delivery {
	return newDelivery(d.LastEvent, itinerary, rs)
}

// IsOnTrack checks if the delivery is on track.
func (d Delivery) IsOnTrack() bool {
	return d.RoutingStatus == Routed && !d.IsMisdirected
}

// DeriveDeliveryFrom creates a new delivery snapshot based on the complete
// handling history of a cargo, as well as its route specification and
// itinerary.
func DeriveDeliveryFrom(rs RouteSpecification, itinerary Itinerary, history HandlingHistory) Delivery {
	lastEvent, _ := history.MostRecentlyCompletedEvent()
	return newDelivery(lastEvent, itinerary, rs)
}

// newDelivery creates a up-to-date delivery based on an handling event,
// itinerary and a route specification.
func newDelivery(lastEvent HandlingEvent, itinerary Itinerary, rs RouteSpecification) Delivery {
	var (
		routingStatus           = calculateRoutingStatus(itinerary, rs)
		transportStatus         = calculateTransportStatus(lastEvent)
		lastKnownLocation       = calculateLastKnownLocation(lastEvent)
		isMisdirected           = calculateMisdirectedStatus(lastEvent, itinerary)
		isUnloadedAtDestination = calculateUnloadedAtDestination(lastEvent, rs)
		currentVoyage           = calculateCurrentVoyage(transportStatus, lastEvent)
	)

	d := Delivery{
		LastEvent:               lastEvent,
		Itinerary:               itinerary,
		RouteSpecification:      rs,
		RoutingStatus:           routingStatus,
		TransportStatus:         transportStatus,
		LastKnownLocation:       lastKnownLocation,
		IsMisdirected:           isMisdirected,
		IsUnloadedAtDestination: isUnloadedAtDestination,
		CurrentVoyage:           currentVoyage,
	}

	d.NextExpectedActivity = calculateNextExpectedActivity(d)
	d.ETA = calculateETA(d)

	return d
}

// Below are internal functions used when creating a new delivery.

func calculateRoutingStatus(itinerary Itinerary, rs RouteSpecification) RoutingStatus {
	if itinerary.Legs == nil {
		return NotRouted
	}

	if rs.IsSatisfiedBy(itinerary) {
		return Routed
	}

	return Misrouted
}

func calculateMisdirectedStatus(event HandlingEvent, itinerary Itinerary) bool {
	if event.Activity.Type == NotHandled {
		return false
	}

	return !itinerary.IsExpected(event)
}

func calculateUnloadedAtDestination(event HandlingEvent, rs RouteSpecification) bool {
	if event.Activity.Type == NotHandled {
		return false
	}

	return event.Activity.Type == Unload && rs.Destination == event.Activity.Location
}

func calculateTransportStatus(event HandlingEvent) TransportStatus {
	switch event.Activity.Type {
	case NotHandled:
		return NotReceived
	case Load:
		return OnboardCarrier
	case Unload:
		return InPort
	case Receive:
		return InPort
	case Customs:
		return InPort
	case Claim:
		return Claimed
	}
	return Unknown
}

func calculateLastKnownLocation(event HandlingEvent) location.UNLocode {
	return event.Activity.Location
}

func calculateNextExpectedActivity(d Delivery) HandlingActivity {
	if !d.IsOnTrack() {
		return HandlingActivity{}
	}

	switch d.LastEvent.Activity.Type {
	case NotHandled:
		return HandlingActivity{Type: Receive, Location: d.RouteSpecification.Origin}
	case Receive:
		l := d.Itinerary.Legs[0]
		return HandlingActivity{Type: Load, Location: l.LoadLocation, VoyageNumber: l.VoyageNumber}
	case Load:
		for _, l := range d.Itinerary.Legs {
			if l.LoadLocation == d.LastEvent.Activity.Location {
				return HandlingActivity{Type: Unload, Location: l.UnloadLocation, VoyageNumber: l.VoyageNumber}
			}
		}
	case Unload:
		for i, l := range d.Itinerary.Legs {
			if l.UnloadLocation == d.LastEvent.Activity.Location {
				if i < len(d.Itinerary.Legs)-1 {
					return HandlingActivity{Type: Load, Location: d.Itinerary.Legs[i+1].LoadLocation, VoyageNumber: d.Itinerary.Legs[i+1].VoyageNumber}
				}

				return HandlingActivity{Type: Claim, Location: l.UnloadLocation}
			}
		}
	}

	return HandlingActivity{}
}

func calculateCurrentVoyage(transportStatus TransportStatus, event HandlingEvent) voyage.Number {
	if transportStatus == OnboardCarrier && event.Activity.Type != NotHandled {
		return event.Activity.VoyageNumber
	}

	return voyage.Number("")
}

func calculateETA(d Delivery) time.Time {
	if !d.IsOnTrack() {
		return time.Time{}
	}

	return d.Itinerary.FinalArrivalTime()
}

type DeliveryRepositoryContract interface {
	Upsert(ctx context.Context, dbtx db.DBTX, delivery Delivery) (Delivery, error)
	Find(ctx context.Context, dbtx db.DBTX, id int64) (Delivery, error)
}

type DeliveryRepository struct {
}

type deliveryResult struct {
	id              int64
	origin          string
	destination     string
	arrivalDeadline sql.NullTime
}

func (dr deliveryResult) build(itinerary Itinerary, event HandlingEvent) Delivery {
	rs := RouteSpecification{
		Origin:          location.UNLocode(dr.origin),
		Destination:     location.UNLocode(dr.destination),
		ArrivalDeadline: dr.arrivalDeadline.Time,
	}

	delivery := newDelivery(event, itinerary, rs)
	delivery.ID = dr.id

	return delivery
}

func (dr DeliveryRepository) Upsert(ctx context.Context, dbtx db.DBTX, delivery Delivery) (Delivery, error) {
	var row *sql.Row
	if delivery.ID == 0 {
		query := `
			INSERT INTO deliveries (itinerary_id, origin, destination, arrival_deadline)
			VALUES ($1, $2, $3, $4)
			RETURNING id, origin, destination, arrival_deadline
		`

		row = dbtx.QueryRowContext(
			ctx,
			query,
			delivery.Itinerary.ID,
			delivery.RouteSpecification.Origin,
			delivery.RouteSpecification.Destination,
			delivery.RouteSpecification.ArrivalDeadline,
		)
	} else {
		query := `
			UPDATE deliveries SET origin = $2, destination = $3, arrival_deadline = $4, last_event = $5
			WHERE id = $1 RETURNING id, origin, destination, arrival_deadline
		`

		var eventType *int
		if delivery.LastEvent.Activity.Type == NotHandled {
			eventType = (*int)(&delivery.LastEvent.Activity.Type)
		}

		row = dbtx.QueryRowContext(
			ctx,
			query,
			delivery.ID,
			delivery.RouteSpecification.Origin,
			delivery.RouteSpecification.Destination,
			delivery.RouteSpecification.ArrivalDeadline,
			eventType,
		)
	}

	var result deliveryResult
	err := row.Scan(&result.id, &result.origin, &result.destination, &result.arrivalDeadline)
	if err != nil {
		return Delivery{}, err
	}

	return result.build(delivery.Itinerary, delivery.LastEvent), nil
}

func (dr DeliveryRepository) Find(ctx context.Context, dbtx db.DBTX, id int64) (Delivery, error) {
	query := `
		SELECT i.id AS itinerary_id, i.legs AS itinerary_legs, d.id AS delivery_id,
		d.origin AS rs_origin, d.destination AS rs_destination, d.arrival_deadline AS rs_arrival_deadline,
		e.id AS event_id, e.tracking_id AS event_tracking_id, e.event_type AS event_type, e.location AS event_location, e.voyage_number AS event_voyage_number
		FROM deliveries AS d
		LEFT JOIN itineraries AS i ON d.itinerary_id = i.id
		LEFT JOIN events AS e ON d.last_event = e.id
		WHERE d.id = $1 LIMIT 1
	`

	var dResult deliveryResult
	var iResult itineraryResult
	var eResult eventResult
	row := dbtx.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&iResult.id,
		&iResult.legs,
		&dResult.id,
		&dResult.origin,
		&dResult.destination,
		&dResult.arrivalDeadline,
		&eResult.id,
		&eResult.trackingId,
		&eResult.eventType,
		&eResult.location,
		&eResult.voyageNumber,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return Delivery{}, nil
		}

		return Delivery{}, err
	}

	itinerary, _ := iResult.build()
	event := eResult.build()
	return dResult.build(itinerary, event), nil
}
