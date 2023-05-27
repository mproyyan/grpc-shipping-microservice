package endpoints

import (
	"context"
	"errors"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/mproyyan/grpc-shipping-microservice/booking/services"
	"github.com/mproyyan/grpc-shipping-microservice/cargo"
	"github.com/mproyyan/grpc-shipping-microservice/location"
)

type Set struct {
	BookNewCargoEndpoint       endpoint.Endpoint
	LoadCargoEndpoint          endpoint.Endpoint
	AssignCargoToRouteEndpoint endpoint.Endpoint
	ChangeDestinationEndpoint  endpoint.Endpoint
	CargosEndpoint             endpoint.Endpoint
}

func NewBookingEndpoints(bs services.BookingServiceContract) Set {
	var bookNewCargoEndpoint = MakeBookNewCargoEndpoint(bs)
	var loadCargoEndpoint = MakeLoadCargoEndpoint(bs)
	var assignCargoToRouteEndpoint = MakeAssignCargoToRouteEndpoint(bs)
	var changeDestinationEndpoint = MakeChangeDestinationEndpoint(bs)
	var listCargosEndpoint = MakeListCargosEndpoint(bs)

	return Set{
		BookNewCargoEndpoint:       bookNewCargoEndpoint,
		LoadCargoEndpoint:          loadCargoEndpoint,
		AssignCargoToRouteEndpoint: assignCargoToRouteEndpoint,
		ChangeDestinationEndpoint:  changeDestinationEndpoint,
		CargosEndpoint:             listCargosEndpoint,
	}
}

type BookNewCargoRequest struct {
	Origin      location.UNLocode
	Destination location.UNLocode
	Deadline    time.Time
}

type BookNewCargoResponse struct {
	TrackingID cargo.TrackingID `json:"tracking_id,omitempty"`
	Error      error            `json:"error,omitempty"`
}

func MakeBookNewCargoEndpoint(bs services.BookingServiceContract) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(BookNewCargoRequest)
		if !ok {
			return nil, errors.New("failed to convert request to BookNewCargoRequest")
		}

		id, err := bs.BookNewCargo(ctx, req.Origin, req.Destination, req.Deadline)
		return BookNewCargoResponse{
			TrackingID: id,
			Error:      err,
		}, nil
	}
}

type LoadCargoRequest struct {
	TrackingID cargo.TrackingID
}

type LoadCargoResponse struct {
	Cargo services.Cargo `json:"cargo,omitempty"`
	Error error          `json:"error,omitempty"`
}

func MakeLoadCargoEndpoint(bs services.BookingServiceContract) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(LoadCargoRequest)
		if !ok {
			return nil, errors.New("failed to convert request to LoadCargoRequest")
		}

		cargo, err := bs.LoadCargo(ctx, req.TrackingID)
		return LoadCargoResponse{
			Cargo: cargo,
			Error: err,
		}, nil
	}
}

type Status string

func newStatus(err error) Status {
	if err != nil {
		return "failed"
	}

	return "success"
}

type AssignCargoToRouteRequest struct {
	TrackingID cargo.TrackingID
	Itinerary  cargo.Itinerary
}

type AssignCargoToRouteResponse struct {
	Status Status `json:"status"`
	Error  error  `json:"error,omitempty"`
}

func MakeAssignCargoToRouteEndpoint(bs services.BookingServiceContract) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(AssignCargoToRouteRequest)
		if !ok {
			return nil, errors.New("failed to convert request to AssignCargoToRouteRequest")
		}

		err = bs.AssignCargoToRoute(ctx, req.TrackingID, req.Itinerary)
		return AssignCargoToRouteResponse{
			Status: newStatus(err),
			Error:  err,
		}, nil
	}
}

type ChangeDestinationRequest struct {
	TrackingID  cargo.TrackingID
	Destination location.UNLocode
}

type ChangeDestinationResponse struct {
	Status Status `json:"status"`
	Error  error  `json:"error,omitempty"`
}

func MakeChangeDestinationEndpoint(bs services.BookingServiceContract) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(ChangeDestinationRequest)
		if !ok {
			return nil, errors.New("failed to convert request to ChangeDestinationRequest")
		}

		err = bs.ChangeDestination(ctx, req.TrackingID, req.Destination)
		return ChangeDestinationResponse{
			Status: newStatus(err),
			Error:  err,
		}, nil
	}
}

type ListCargosRequest struct{}

type ListCargosResponse struct {
	Cargos []services.Cargo `json:"cargos"`
	Error  error            `json:"error,omitempty"`
}

func MakeListCargosEndpoint(bs services.BookingServiceContract) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		_, ok := request.(ListCargosRequest)
		if !ok {
			return nil, errors.New("failed to convert request to ListCargoRequest")
		}

		cargos, err := bs.Cargos(ctx)
		return ListCargosResponse{
			Cargos: cargos,
			Error:  err,
		}, nil
	}
}
