package transports

import (
	"context"
	"errors"
	"fmt"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mproyyan/grpc-shipping-microservice/booking/endpoints"
	"github.com/mproyyan/grpc-shipping-microservice/booking/services"
	"github.com/mproyyan/grpc-shipping-microservice/cargo"
	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/mproyyan/grpc-shipping-microservice/pb"
	"github.com/mproyyan/grpc-shipping-microservice/voyage"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type bookingGRPCServer struct {
	pb.UnimplementedBookingServer
	bookNewCargo       gt.Handler
	loadCargo          gt.Handler
	assignCargoToRoute gt.Handler
	changeDestination  gt.Handler
	listCargos         gt.Handler
}

func NewGRPCServer(endpoints endpoints.Set) pb.BookingServer {
	return bookingGRPCServer{
		bookNewCargo: gt.NewServer(
			endpoints.BookNewCargoEndpoint,
			decodeGRPCBookNewCargoRequest,
			encodeGRPCBookNewCargoResponse,
		),
		loadCargo: gt.NewServer(
			endpoints.LoadCargoEndpoint,
			decodeGRPCLoadCargoRequest,
			encodeGRPCLoadCargoResponse,
		),
		assignCargoToRoute: gt.NewServer(
			endpoints.AssignCargoToRouteEndpoint,
			decodeGRPCAssignCargoToRouteRequest,
			encodeGRPCAssignCargoToRouteResponse,
		),
		changeDestination: gt.NewServer(
			endpoints.ChangeDestinationEndpoint,
			decodeGRPCChangeDestinationRequest,
			encodeGRPCChangeDestinationResponse,
		),
		listCargos: gt.NewServer(
			endpoints.CargosEndpoint,
			decodeGRPCListCargosRequest,
			encodeGRPCListCargosResponse,
		),
	}
}

func NewGRPCClient(conn *grpc.ClientConn) services.BookingServiceContract {
	bookNewCargoEndpoint := gt.NewClient(
		conn,
		"pb.Booking",
		"BookNewCargo",
		encodeGRPCBookNewCargoRequest,
		decodeGRPCBookNewCargoResponse,
		pb.BookNewCargoResponse{},
	).Endpoint()

	loadCargoEndpoint := gt.NewClient(
		conn,
		"pb.Booking",
		"LoadCargo",
		encodeGRPCLoadCargoRequest,
		decodeGRPCLoadCargoResponse,
		pb.LoadCargoResponse{},
	).Endpoint()

	assignCargoToRouteEndpoint := gt.NewClient(
		conn,
		"pb.Booking",
		"AssignCargoToRoute",
		encodeGRPCAssignCargoToRouteRequest,
		decodeGRPCAssignCargoToRouteResponse,
		pb.AssignCargoToRouteResponse{},
	).Endpoint()

	changeDestinationEndpoint := gt.NewClient(
		conn,
		"pb.Booking",
		"ChangeDestination",
		encodeGRPCChangeDestinationRequest,
		decodeGRPCChangeDestinationResponse,
		pb.ChangeDestinationResponse{},
	).Endpoint()

	listCargosEndpoint := gt.NewClient(
		conn,
		"pb.Booking",
		"Cargos",
		encodeGRPCListCargosRequest,
		decodeGRPCListCargosResponse,
		pb.CargosResponse{},
	).Endpoint()

	return endpoints.Set{
		BookNewCargoEndpoint:       bookNewCargoEndpoint,
		LoadCargoEndpoint:          loadCargoEndpoint,
		AssignCargoToRouteEndpoint: assignCargoToRouteEndpoint,
		ChangeDestinationEndpoint:  changeDestinationEndpoint,
		CargosEndpoint:             listCargosEndpoint,
	}
}

func (bgs bookingGRPCServer) BookNewCargo(ctx context.Context, req *pb.BookNewCargoRequest) (*pb.BookNewCargoResponse, error) {
	_, resp, err := bgs.bookNewCargo.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.BookNewCargoResponse), nil
}

func (bgs bookingGRPCServer) LoadCargo(ctx context.Context, req *pb.LoadCargoRequest) (*pb.LoadCargoResponse, error) {
	_, resp, err := bgs.loadCargo.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.LoadCargoResponse), nil
}

func (bgs bookingGRPCServer) AssignCargoToRoute(ctx context.Context, req *pb.AssignCargoToRouteRequest) (*pb.AssignCargoToRouteResponse, error) {
	_, resp, err := bgs.assignCargoToRoute.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.AssignCargoToRouteResponse), nil
}

func (bgs bookingGRPCServer) ChangeDestination(ctx context.Context, req *pb.ChangeDestinationRequest) (*pb.ChangeDestinationResponse, error) {
	_, resp, err := bgs.changeDestination.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.ChangeDestinationResponse), nil
}

func (bgs bookingGRPCServer) Cargos(ctx context.Context, _ *empty.Empty) (*pb.CargosResponse, error) {
	_, resp, err := bgs.listCargos.ServeGRPC(ctx, nil)
	if err != nil {
		return nil, err
	}

	return resp.(*pb.CargosResponse), nil
}

// booking server
// book new cargo
func decodeGRPCBookNewCargoRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {
	req, ok := grpcReq.(*pb.BookNewCargoRequest)
	if !ok {
		return nil, errors.New("failed to convert grpc request to *pb.BookNewCargoRequest")
	}

	cr := endpoints.BookNewCargoRequest{}
	return cr.Build(req), nil
}

func encodeGRPCBookNewCargoResponse(ctx context.Context, response interface{}) (interface{}, error) {
	res, ok := response.(endpoints.BookNewCargoResponse)
	if !ok {
		return nil, errors.New("failed to convert response to endpoints.BookNewCargoResponse")
	}

	return res.Protobuf(), nil
}

// load cargo
func decodeGRPCLoadCargoRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {
	req, ok := grpcReq.(*pb.LoadCargoRequest)
	if !ok {
		return nil, errors.New("failed to convert grpc request to *pb.LoadCargoRequest")
	}

	cr := endpoints.LoadCargoRequest{}
	return cr.Build(req), nil
}

func encodeGRPCLoadCargoResponse(ctx context.Context, response interface{}) (interface{}, error) {
	res, ok := response.(endpoints.LoadCargoResponse)
	if !ok {
		return nil, errors.New("failed to convert response to endpoints.LoadCargoResponse")
	}

	return res.Protobuf(), nil
}

// assign cargo to route
func decodeGRPCAssignCargoToRouteRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {
	req, ok := grpcReq.(*pb.AssignCargoToRouteRequest)
	if !ok {
		return nil, errors.New("failed to convert grpc request to *pb.AssignCargoToRouteRequest")
	}

	cr := endpoints.AssignCargoToRouteRequest{}
	return cr.Build(req), nil
}

func encodeGRPCAssignCargoToRouteResponse(ctx context.Context, response interface{}) (interface{}, error) {
	res, ok := response.(endpoints.AssignCargoToRouteResponse)
	if !ok {
		return nil, errors.New("failed to convert response to endpoints.AssignCargoToRouteResponse")
	}

	return res.Protobuf(), nil
}

// change cargo destination
func decodeGRPCChangeDestinationRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {
	req, ok := grpcReq.(*pb.ChangeDestinationRequest)
	if !ok {
		return nil, errors.New("failed to convert grpc request to *pb.ChangeDestinationRequest")
	}

	cr := endpoints.ChangeDestinationRequest{}
	return cr.Build(req), nil
}

func encodeGRPCChangeDestinationResponse(ctx context.Context, response interface{}) (interface{}, error) {
	res, ok := response.(endpoints.ChangeDestinationResponse)
	if !ok {
		return nil, errors.New("failed to convert response to endpoints.ChangeDestinationResponse")
	}

	return res.Protobuf(), nil
}

// change cargo destination
func decodeGRPCListCargosRequest(ctx context.Context, grpcReq interface{}) (interface{}, error) {
	return endpoints.ListCargosRequest{}, nil
}

func encodeGRPCListCargosResponse(ctx context.Context, response interface{}) (interface{}, error) {
	res, ok := response.(endpoints.ListCargosResponse)
	if !ok {
		return nil, errors.New("failed to convert response to endpoints.ListCargosResponse")
	}

	return res.Protobuf(), nil
}

// booking client
// book new cargo
func encodeGRPCBookNewCargoRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(endpoints.BookNewCargoRequest)
	if !ok {
		return nil, errors.New("failed to convert request to endpoints.BookNewCargoRequest")
	}

	return &pb.BookNewCargoRequest{
		Origin:      string(req.Origin),
		Destination: string(req.Destination),
		Deadline:    timestamppb.New(req.Deadline),
	}, nil
}

func decodeGRPCBookNewCargoResponse(ctx context.Context, grpcReply interface{}) (interface{}, error) {
	reply, ok := grpcReply.(*pb.BookNewCargoResponse)
	if !ok {
		return nil, errors.New("failed to convert response to *pb.BookNewCargoResponse")
	}

	return endpoints.BookNewCargoResponse{
		TrackingID: cargo.TrackingID(reply.TrackingId),
		Error:      str2err(reply.Error),
	}, nil
}

// load cargo
func encodeGRPCLoadCargoRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(endpoints.LoadCargoRequest)
	if !ok {
		return nil, errors.New("failed to convert request to endpoints.LoadCargoRequest")
	}

	return &pb.LoadCargoRequest{
		TrackingId: string(req.TrackingID),
	}, nil
}

func decodeGRPCLoadCargoResponse(ctx context.Context, grpcReply interface{}) (interface{}, error) {
	reply, ok := grpcReply.(*pb.LoadCargoResponse)
	if !ok {
		return nil, errors.New("failed to convert response to *pb.LoadCargoResponse")
	}

	var legs []cargo.Leg
	for _, l := range reply.Cargo.Legs {
		leg := cargo.Leg{
			LoadLocation:   location.UNLocode(l.LoadLocation),
			UnloadLocation: location.UNLocode(l.UnloadLocation),
			LoadTime:       l.LoadTime.AsTime(),
			UnloadTime:     l.UnloadTime.AsTime(),
			VoyageNumber:   voyage.Number(l.VoyageNumber),
		}

		legs = append(legs, leg)
	}

	return endpoints.LoadCargoResponse{
		Cargo: services.Cargo{
			ArrivalDeadline: reply.GetCargo().ArrivalDeadline.AsTime(),
			Destination:     reply.Cargo.Destination,
			Legs:            legs,
			Misrouted:       reply.GetCargo().Misrouted,
			Origin:          reply.Cargo.Origin,
			Routed:          reply.Cargo.Routed,
			TrackingID:      reply.Cargo.TrackingId,
		},
	}, nil
}

// assign cargo to route
func encodeGRPCAssignCargoToRouteRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(endpoints.AssignCargoToRouteRequest)
	if !ok {
		return nil, errors.New("failed to convert request to endpoints.AssignCargoToRouteRequest")
	}
	fmt.Println(req)
	var legs []*pb.Leg
	for _, l := range req.Itinerary.Legs {
		leg := &pb.Leg{
			LoadLocation:   string(l.LoadLocation),
			LoadTime:       timestamppb.New(l.LoadTime),
			UnloadLocation: string(l.UnloadLocation),
			UnloadTime:     timestamppb.New(l.UnloadTime),
			VoyageNumber:   string(l.VoyageNumber),
		}

		legs = append(legs, leg)
	}

	return &pb.AssignCargoToRouteRequest{
		TrackingId: string(req.TrackingID),
		Itinerary: &pb.Itinerary{
			Id:   req.Itinerary.ID,
			Legs: legs,
		},
	}, nil
}

func decodeGRPCAssignCargoToRouteResponse(ctx context.Context, grpcReply interface{}) (interface{}, error) {
	reply, ok := grpcReply.(*pb.AssignCargoToRouteResponse)
	if !ok {
		return nil, errors.New("failed to convert response to *pb.AssignCargoToRouteResponse")
	}

	return endpoints.AssignCargoToRouteResponse{
		Error: str2err(reply.Error),
	}, nil
}

// change destination
func encodeGRPCChangeDestinationRequest(ctx context.Context, request interface{}) (interface{}, error) {
	req, ok := request.(endpoints.ChangeDestinationRequest)
	if !ok {
		return nil, errors.New("failed to convert request to endpoints.ChangeDestinationRequest")
	}

	return &pb.ChangeDestinationRequest{
		TrackingId:  string(req.TrackingID),
		Destination: string(req.Destination),
	}, nil
}

func decodeGRPCChangeDestinationResponse(ctx context.Context, grpcReply interface{}) (interface{}, error) {
	reply, ok := grpcReply.(*pb.ChangeDestinationResponse)
	if !ok {
		return nil, errors.New("failed to convert response to *pb.ChangeDestinationResponse")
	}

	return endpoints.ChangeDestinationResponse{
		Error: str2err(reply.Error),
	}, nil
}

// list cargos
// book new cargo
func encodeGRPCListCargosRequest(ctx context.Context, request interface{}) (interface{}, error) {
	_, ok := request.(endpoints.ListCargosRequest)
	if !ok {
		return nil, errors.New("failed to convert request to endpoints.ListCargosRequest")
	}

	return nil, nil
}

func decodeGRPCListCargosResponse(ctx context.Context, grpcReply interface{}) (interface{}, error) {
	reply, ok := grpcReply.(*pb.CargosResponse)
	if !ok {
		return nil, errors.New("failed to convert response to *pb.CargosResponse")
	}

	var cargos []services.Cargo
	for _, c := range reply.Cargos {
		var legs []cargo.Leg
		for _, l := range c.Legs {
			leg := cargo.Leg{
				LoadLocation:   location.UNLocode(l.LoadLocation),
				LoadTime:       l.LoadTime.AsTime(),
				UnloadLocation: location.UNLocode(l.UnloadLocation),
				UnloadTime:     l.UnloadTime.AsTime(),
				VoyageNumber:   voyage.Number(l.VoyageNumber),
			}

			legs = append(legs, leg)
		}

		cargo := services.Cargo{
			ArrivalDeadline: c.ArrivalDeadline.AsTime(),
			Destination:     c.Destination,
			Legs:            legs,
			Misrouted:       c.Misrouted,
			Origin:          c.Origin,
			Routed:          c.Routed,
			TrackingID:      c.TrackingId,
		}

		cargos = append(cargos, cargo)
	}

	return endpoints.ListCargosResponse{
		Cargos: cargos,
	}, nil
}

func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}
