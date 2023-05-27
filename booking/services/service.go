package services

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/cargo"
	"github.com/mproyyan/grpc-shipping-microservice/location"
)

var ErrInvalidArgument = errors.New("invalid argument")

type BookingServiceContract interface {
	BookNewCargo(ctx context.Context, origin location.UNLocode, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error)
	LoadCargo(ctx context.Context, id cargo.TrackingID) (Cargo, error)
	AssignCargoToRoute(ctx context.Context, id cargo.TrackingID, itinerary cargo.Itinerary) error
	ChangeDestination(ctx context.Context, id cargo.TrackingID, destination location.UNLocode) error
	Cargos(ctx context.Context) ([]Cargo, error)
}

type BookingService struct {
	db     *sql.DB
	cargos cargo.CargoRepositoryContract
	events cargo.EventRepositoryContract
}

func NewBookingService(cargos cargo.CargoRepositoryContract, events cargo.EventRepositoryContract) BookingService {
	return BookingService{
		cargos: cargos,
		events: events,
	}
}

func (bs *BookingService) BookNewCargo(ctx context.Context, origin location.UNLocode, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error) {
	if origin == "" || destination == "" || deadline.IsZero() {
		return "", ErrInvalidArgument
	}

	id := cargo.NextTrackingID()
	rs := cargo.RouteSpecification{
		Origin:          origin,
		Destination:     destination,
		ArrivalDeadline: deadline,
	}

	c := cargo.New(id, rs)
	c, err := bs.cargos.Upsert(ctx, bs.db, c)
	if err != nil {
		return "", err
	}

	return c.TrackingID, nil
}

func (bs *BookingService) LoadCargo(ctx context.Context, id cargo.TrackingID) (Cargo, error) {
	if id == "" {
		return Cargo{}, ErrInvalidArgument
	}

	c, err := bs.cargos.Find(ctx, bs.db, id)
	if err != nil {
		return Cargo{}, err
	}

	return assemble(c, bs.events), nil
}

func (bs *BookingService) AssignCargoToRoute(ctx context.Context, id cargo.TrackingID, itinerary cargo.Itinerary) error {
	if id == "" || len(itinerary.Legs) == 0 {
		return ErrInvalidArgument
	}

	c, err := bs.cargos.Find(ctx, bs.db, id)
	if err != nil {
		return err
	}

	c.AssignToRoute(itinerary)
	_, err = bs.cargos.Upsert(ctx, bs.db, c)
	if err != nil {
		return err
	}

	return nil
}

func (bs *BookingService) ChangeDestination(ctx context.Context, id cargo.TrackingID, destination location.UNLocode) error {
	if id == "" || destination == "" {
		return ErrInvalidArgument
	}

	c, err := bs.cargos.Find(ctx, bs.db, id)
	if err != nil {
		return err
	}

	c.SpecifyNewRoute(cargo.RouteSpecification{
		Origin:          c.Origin,
		Destination:     destination,
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
	})

	_, err = bs.cargos.Upsert(ctx, bs.db, c)
	if err != nil {
		return err
	}

	return nil
}

func (bs *BookingService) Cargos(ctx context.Context) ([]Cargo, error) {
	var results []Cargo
	cargos, err := bs.cargos.FindAll(ctx, bs.db)
	if err != nil {
		return results, err
	}

	for _, c := range cargos {
		results = append(results, assemble(c, bs.events))
	}

	return results, nil
}

type Cargo struct {
	ArrivalDeadline time.Time   `json:"arrival_deadline"`
	Destination     string      `json:"destination"`
	Legs            []cargo.Leg `json:"legs,omitempty"`
	Misrouted       bool        `json:"misrouted"`
	Origin          string      `json:"origin"`
	Routed          bool        `json:"routed"`
	TrackingID      string      `json:"tracking_id"`
}

func assemble(c *cargo.Cargo, events cargo.EventRepositoryContract) Cargo {
	return Cargo{
		TrackingID:      string(c.TrackingID),
		Origin:          string(c.Origin),
		Destination:     string(c.RouteSpecification.Destination),
		Misrouted:       c.Delivery.RoutingStatus == cargo.Misrouted,
		Routed:          !c.Itinerary.IsEmpty(),
		ArrivalDeadline: c.RouteSpecification.ArrivalDeadline,
		Legs:            c.Itinerary.Legs,
	}
}
