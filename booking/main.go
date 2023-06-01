package main

import (
	"log"
	"net"
	"os"

	"github.com/mproyyan/grpc-shipping-microservice/booking/endpoints"
	"github.com/mproyyan/grpc-shipping-microservice/booking/services"
	"github.com/mproyyan/grpc-shipping-microservice/booking/transports"
	"github.com/mproyyan/grpc-shipping-microservice/cargo"
	"github.com/mproyyan/grpc-shipping-microservice/config"
	"github.com/mproyyan/grpc-shipping-microservice/db"
	"github.com/mproyyan/grpc-shipping-microservice/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	grpcListener, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Print("cannot start grpc listener :", err)
		os.Exit(1)
	}

	env, err := config.LoadEnv(".", "app")
	if err != nil {
		log.Print("failed to load environment file :", err)
		os.Exit(1)
	}

	db, err := db.NewPostgreSQL(env).Connect()
	if err != nil {
		log.Print("failed to open database connection :", err)
		os.Exit(1)
	}

	var (
		itineraries = cargo.NewItineraryRepository()
		deliveries  = cargo.NewDeliveryRepository()
		cargos      = cargo.NewCargoRepository(itineraries, deliveries)
		events      = cargo.NewEventRepository()
	)

	var (
		service    = services.NewBookingService(db, cargos, events)
		ep         = endpoints.NewBookingEndpoints(service)
		grpcServer = transports.NewGRPCServer(ep)
	)

	baseServer := grpc.NewServer()
	pb.RegisterBookingServer(baseServer, grpcServer)
	reflection.Register(baseServer)

	log.Print("server running on :8888")
	err = baseServer.Serve(grpcListener)
	if err != nil {
		log.Print("cannot start grpc server :", err)
		os.Exit(1)
	}
}
