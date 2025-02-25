package endpoints

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/mproyyan/grpc-shipping-microservice/booking/services"
	"github.com/mproyyan/grpc-shipping-microservice/cargo"
	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/mproyyan/grpc-shipping-microservice/pb"
	"github.com/mproyyan/grpc-shipping-microservice/voyage"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (s Set) BookNewCargo(ctx context.Context, origin location.UNLocode, destination location.UNLocode, deadline time.Time) (cargo.TrackingID, error) {
	resp, err := s.BookNewCargoEndpoint(ctx, BookNewCargoRequest{
		Origin:      origin,
		Destination: destination,
		Deadline:    deadline,
	})

	if err != nil {
		return "", err
	}

	res := resp.(BookNewCargoResponse)
	return res.TrackingID, nil
}

func (s Set) LoadCargo(ctx context.Context, id cargo.TrackingID) (services.Cargo, error) {
	resp, err := s.LoadCargoEndpoint(ctx, LoadCargoRequest{
		TrackingID: id,
	})

	if err != nil {
		return services.Cargo{}, err
	}

	res := resp.(LoadCargoResponse)
	return res.Cargo, nil
}

func (s Set) AssignCargoToRoute(ctx context.Context, id cargo.TrackingID, itinerary cargo.Itinerary) error {
	fmt.Println(itinerary)
	_, err := s.AssignCargoToRouteEndpoint(ctx, AssignCargoToRouteRequest{
		TrackingID: id,
		Itinerary:  itinerary,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s Set) ChangeDestination(ctx context.Context, id cargo.TrackingID, destination location.UNLocode) error {
	_, err := s.ChangeDestinationEndpoint(ctx, ChangeDestinationRequest{
		TrackingID:  id,
		Destination: destination,
	})

	if err != nil {
		return err
	}

	return nil
}

func (s Set) Cargos(ctx context.Context) ([]services.Cargo, error) {
	resp, err := s.CargosEndpoint(ctx, ListCargosRequest{})
	if err != nil {
		return nil, err
	}

	res := resp.(ListCargosResponse)
	return res.Cargos, nil
}

type BookNewCargoRequest struct {
	Origin      location.UNLocode `json:"origin"`
	Destination location.UNLocode `json:"destination"`
	Deadline    time.Time         `json:"deadline"`
}

func (bncreq BookNewCargoRequest) Build(req *pb.BookNewCargoRequest) BookNewCargoRequest {
	return BookNewCargoRequest{
		Origin:      location.UNLocode(req.GetOrigin()),
		Destination: location.UNLocode(req.GetDestination()),
		Deadline:    req.Deadline.AsTime(),
	}
}

type BookNewCargoResponse struct {
	TrackingID cargo.TrackingID `json:"tracking_id,omitempty"`
	Error      error            `json:"error,omitempty"`
}

func (bncres BookNewCargoResponse) error() error { return bncres.Error }

func (bncres BookNewCargoResponse) Protobuf() *pb.BookNewCargoResponse {
	return &pb.BookNewCargoResponse{
		TrackingId: string(bncres.TrackingID),
		Error:      err2str(bncres.Error),
	}
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
	TrackingID cargo.TrackingID `json:"tracking_id"`
}

func (lcreq LoadCargoRequest) Build(req *pb.LoadCargoRequest) LoadCargoRequest {
	return LoadCargoRequest{
		TrackingID: cargo.TrackingID(req.TrackingId),
	}
}

type LoadCargoResponse struct {
	Cargo services.Cargo `json:"cargo,omitempty"`
	Error error          `json:"error,omitempty"`
}

func (res LoadCargoResponse) error() error { return res.Error }

func (lcres LoadCargoResponse) Protobuf() *pb.LoadCargoResponse {
	var legs []*pb.Leg
	for _, l := range lcres.Cargo.Legs {
		leg := &pb.Leg{
			LoadLocation:   string(l.LoadLocation),
			LoadTime:       timestamppb.New(l.LoadTime),
			UnloadLocation: string(l.UnloadLocation),
			UnloadTime:     timestamppb.New(l.UnloadTime),
			VoyageNumber:   string(l.VoyageNumber),
		}

		legs = append(legs, leg)
	}

	return &pb.LoadCargoResponse{
		Cargo: &pb.BookingCargoModel{
			TrackingId:      string(lcres.Cargo.TrackingID),
			ArrivalDeadline: timestamppb.New(lcres.Cargo.ArrivalDeadline),
			Destination:     lcres.Cargo.Destination,
			Legs:            legs,
			Misrouted:       lcres.Cargo.Misrouted,
			Origin:          string(lcres.Cargo.Origin),
			Routed:          lcres.Cargo.Routed,
		},
		Error: err2str(lcres.Error),
	}
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
	TrackingID cargo.TrackingID `json:"tracking_id"`
	Itinerary  cargo.Itinerary  `json:"itinerary"`
}

func (r AssignCargoToRouteRequest) Build(req *pb.AssignCargoToRouteRequest) AssignCargoToRouteRequest {
	var legs []cargo.Leg
	for _, l := range req.Itinerary.Legs {
		leg := cargo.Leg{
			LoadLocation:   location.UNLocode(l.LoadLocation),
			LoadTime:       l.LoadTime.AsTime(),
			UnloadLocation: location.UNLocode(l.UnloadLocation),
			UnloadTime:     l.UnloadTime.AsTime(),
			VoyageNumber:   voyage.Number(l.VoyageNumber),
		}

		legs = append(legs, leg)
	}

	return AssignCargoToRouteRequest{
		TrackingID: cargo.TrackingID(req.TrackingId),
		Itinerary: cargo.Itinerary{
			ID:   req.Itinerary.GetId(),
			Legs: legs,
		},
	}
}

type AssignCargoToRouteResponse struct {
	Status Status `json:"status"`
	Error  error  `json:"error,omitempty"`
}

func (res AssignCargoToRouteResponse) error() error { return res.Error }

func (r AssignCargoToRouteResponse) Protobuf() *pb.AssignCargoToRouteResponse {
	return &pb.AssignCargoToRouteResponse{
		Error: err2str(r.Error),
	}
}

func MakeAssignCargoToRouteEndpoint(bs services.BookingServiceContract) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req, ok := request.(AssignCargoToRouteRequest)
		if !ok {
			return nil, errors.New("failed to convert request to AssignCargoToRouteRequest")
		}
		fmt.Println("req.itineraty :", req.Itinerary)
		err = bs.AssignCargoToRoute(ctx, req.TrackingID, req.Itinerary)
		return AssignCargoToRouteResponse{
			Status: newStatus(err),
			Error:  err,
		}, nil
	}
}

type ChangeDestinationRequest struct {
	TrackingID  cargo.TrackingID  `json:"tracking_id"`
	Destination location.UNLocode `json:"destination"`
}

func (r ChangeDestinationRequest) Build(req *pb.ChangeDestinationRequest) ChangeDestinationRequest {
	return ChangeDestinationRequest{
		TrackingID:  cargo.TrackingID(req.TrackingId),
		Destination: location.UNLocode(req.Destination),
	}
}

type ChangeDestinationResponse struct {
	Status Status `json:"status"`
	Error  error  `json:"error,omitempty"`
}

func (res ChangeDestinationResponse) error() error { return res.Error }

func (r ChangeDestinationResponse) Protobuf() *pb.ChangeDestinationResponse {
	return &pb.ChangeDestinationResponse{
		Error: err2str(r.Error),
	}
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

func (res ListCargosResponse) error() error { return res.Error }

func (r ListCargosResponse) Protobuf() *pb.CargosResponse {
	var cargos []*pb.BookingCargoModel
	for _, c := range r.Cargos {
		var legs []*pb.Leg
		for _, l := range c.Legs {
			leg := &pb.Leg{
				LoadLocation:   string(l.LoadLocation),
				LoadTime:       timestamppb.New(l.LoadTime),
				UnloadLocation: string(l.UnloadLocation),
				UnloadTime:     timestamppb.New(l.UnloadTime),
				VoyageNumber:   string(l.VoyageNumber),
			}

			legs = append(legs, leg)
		}

		cargo := &pb.BookingCargoModel{
			TrackingId:      string(c.TrackingID),
			ArrivalDeadline: timestamppb.New(c.ArrivalDeadline),
			Destination:     string(c.Destination),
			Legs:            legs,
			Misrouted:       c.Misrouted,
			Origin:          c.Origin,
			Routed:          c.Routed,
		}

		cargos = append(cargos, cargo)
	}

	return &pb.CargosResponse{Cargos: cargos}
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

func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
