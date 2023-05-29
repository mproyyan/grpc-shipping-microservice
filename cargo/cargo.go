package cargo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/db"
	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/pborman/uuid"
)

// TrackingID uniquely identifies a particular cargo.
type TrackingID string

// NextTrackingID generates a new tracking ID.
// TODO: Move to infrastructure(?)
func NextTrackingID() TrackingID {
	return TrackingID(strings.Split(strings.ToUpper(uuid.New()), "-")[0])
}

// Cargo is the central class in the domain model.
type Cargo struct {
	TrackingID         TrackingID
	Origin             location.UNLocode
	RouteSpecification RouteSpecification
	Itinerary          Itinerary
	Delivery           Delivery
}

// SpecifyNewRoute specifies a new route for this cargo.
func (c *Cargo) SpecifyNewRoute(rs RouteSpecification) {
	c.RouteSpecification = rs
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

// AssignToRoute attaches a new itinerary to this cargo.
func (c *Cargo) AssignToRoute(itinerary Itinerary) {
	c.Itinerary = itinerary
	c.Delivery = c.Delivery.UpdateOnRouting(c.RouteSpecification, c.Itinerary)
}

// DeriveDeliveryProgress updates all aspects of the cargo aggregate status
// based on the current route specification, itinerary and handling of the cargo.
func (c *Cargo) DeriveDeliveryProgress(history HandlingHistory) {
	c.Delivery = DeriveDeliveryFrom(c.RouteSpecification, c.Itinerary, history)
}

// New creates a new, unrouted cargo.
func New(id TrackingID, rs RouteSpecification) *Cargo {
	itinerary := Itinerary{}
	history := HandlingHistory{make([]HandlingEvent, 0)}

	return &Cargo{
		TrackingID:         id,
		Origin:             rs.Origin,
		RouteSpecification: rs,
		Delivery:           DeriveDeliveryFrom(rs, itinerary, history),
	}
}

type CargoRepositoryContract interface {
	Upsert(ctx context.Context, dbtx db.DBTX, cargo *Cargo) (*Cargo, error)
	Find(ctx context.Context, dbtx db.DBTX, trackingID TrackingID) (*Cargo, error)
	FindAll(ctx context.Context, dbtx db.DBTX) ([]*Cargo, error)
}

type CargoRepository struct {
	ItineraryRepository ItineraryRepositoryContract
	DeliveryRepository  DeliveryRepositoryContract
}

func NewCargoRepository(itineraryRepository ItineraryRepositoryContract, deliveryRepository DeliveryRepositoryContract) CargoRepository {
	return CargoRepository{
		ItineraryRepository: itineraryRepository,
		DeliveryRepository:  deliveryRepository,
	}
}

type cargoResult struct {
	trackingID      string
	origin          string
	destination     string
	arrivalDeadline time.Time
	itineraryID     int64
	deliveryID      int64
}

func (cr cargoResult) build(itinerary Itinerary, delivery Delivery) *Cargo {
	return &Cargo{
		TrackingID: TrackingID(cr.trackingID),
		Origin:     location.UNLocode(cr.origin),
		RouteSpecification: RouteSpecification{
			Origin:          location.UNLocode(cr.origin),
			Destination:     location.UNLocode(cr.destination),
			ArrivalDeadline: cr.arrivalDeadline,
		},
		Itinerary: itinerary,
		Delivery:  delivery,
	}
}

func (cr CargoRepository) Upsert(ctx context.Context, dbtx db.DBTX, cargo *Cargo) (*Cargo, error) {
	// TODO: implement database transaction
	itinerary, err := cr.ItineraryRepository.Upsert(ctx, dbtx, cargo.Itinerary)
	if err != nil {
		return nil, err
	}

	cargo.Delivery.Itinerary.ID = itinerary.ID
	delivery, err := cr.DeliveryRepository.Upsert(ctx, dbtx, cargo.Delivery)
	if err != nil {
		return nil, err
	}

	var row *sql.Row
	if cargo.Itinerary.ID == 0 && cargo.Delivery.ID == 0 {
		query := `
			INSERT INTO cargos (tracking_id, origin, destination, arrival_deadline, itinerary_id, delivery_id)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING tracking_id, origin, destination, arrival_deadline
		`

		row = dbtx.QueryRowContext(
			ctx,
			query,
			cargo.TrackingID,
			cargo.RouteSpecification.Origin,
			cargo.RouteSpecification.Destination,
			cargo.RouteSpecification.ArrivalDeadline,
			itinerary.ID,
			delivery.ID,
		)
	} else {
		query := `
			UPDATE cargos SET origin = $2, destination = $3, arrival_deadline = $4
			WHERE tracking_id = $1 RETURNING tracking_id, origin, destination, arrival_deadline
		`

		row = dbtx.QueryRowContext(
			ctx,
			query,
			cargo.TrackingID,
			cargo.RouteSpecification.Origin,
			cargo.RouteSpecification.Destination,
			cargo.RouteSpecification.ArrivalDeadline,
		)
	}

	var result cargoResult
	err = row.Scan(&result.trackingID, &result.origin, &result.destination, &result.arrivalDeadline)
	if err != nil {
		fmt.Println("err :", err)
		return nil, err
	}

	return result.build(itinerary, delivery), nil
}

func (cr CargoRepository) Find(ctx context.Context, dbtx db.DBTX, trackingID TrackingID) (*Cargo, error) {
	query := `
		SELECT tracking_id, origin, destination, arrival_deadline, itinerary_id, delivery_id
		FROM cargos WHERE tracking_id = $1 LIMIT 1
	`

	var result cargoResult
	row := dbtx.QueryRowContext(ctx, query, trackingID)
	err := row.Scan(&result.trackingID, &result.origin, &result.destination, &result.arrivalDeadline, &result.itineraryID, &result.deliveryID)
	if err != nil {
		return nil, err
	}

	// Should i implement database transaction?
	itinerary, err := cr.ItineraryRepository.Find(ctx, dbtx, result.itineraryID)
	if err != nil {
		return nil, err
	}

	delivery, err := cr.DeliveryRepository.Find(ctx, dbtx, result.deliveryID)
	if err != nil {
		return nil, err
	}

	return result.build(itinerary, delivery), nil
}

func (cr CargoRepository) FindAll(ctx context.Context, dbtx db.DBTX) ([]*Cargo, error) {
	query := `
		SELECT tracking_id, origin, destination, arrival_deadline, itinerary_id, delivery_id
		FROM cargos
	`

	rows, err := dbtx.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var result cargoResult
	var cargos []*Cargo
	for rows.Next() {
		err = rows.Scan(&result.trackingID, &result.origin, &result.destination, &result.arrivalDeadline, &result.itineraryID, &result.deliveryID)
		if err != nil {
			return nil, err
		}

		// is it better to use multiple joins instead of calling each repository?
		itinerary, err := cr.ItineraryRepository.Find(ctx, dbtx, result.itineraryID)
		if err != nil {
			return nil, err
		}

		delivery, err := cr.DeliveryRepository.Find(ctx, dbtx, result.deliveryID)
		if err != nil {
			return nil, err
		}

		cargo := result.build(itinerary, delivery)
		cargos = append(cargos, cargo)
	}

	return cargos, nil
}
