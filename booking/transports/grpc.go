package transports

import (
	"context"
	"errors"

	gt "github.com/go-kit/kit/transport/grpc"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mproyyan/grpc-shipping-microservice/booking/endpoints"
	"github.com/mproyyan/grpc-shipping-microservice/pb"
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
