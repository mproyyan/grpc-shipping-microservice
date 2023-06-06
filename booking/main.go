package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/mproyyan/grpc-shipping-microservice/booking/endpoints"
	"github.com/mproyyan/grpc-shipping-microservice/booking/services"
	"github.com/mproyyan/grpc-shipping-microservice/booking/transports"
	"github.com/mproyyan/grpc-shipping-microservice/cargo"
	"github.com/mproyyan/grpc-shipping-microservice/config"
	"github.com/mproyyan/grpc-shipping-microservice/db"
	"github.com/mproyyan/grpc-shipping-microservice/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	var (
		id       = flag.String("id", "", "service id")
		grpcPort = flag.Int("grpcPort", 8888, "port for grpc server")
	)

	flag.Parse()

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
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
	healthProbe := health.NewServer()
	grpc_health_v1.RegisterHealthServer(baseServer, healthProbe)
	pb.RegisterBookingServer(baseServer, grpcServer)

	reflection.Register(baseServer)

	config := consulapi.DefaultConfig()
	consul, err := consulapi.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}

	registration := &consulapi.AgentServiceRegistration{
		ID:      *id,
		Name:    "bookingservice",
		Port:    *grpcPort,
		Address: "localhost",
		Check: &consulapi.AgentServiceCheck{
			Name:     "bookingservice-check",
			GRPC:     fmt.Sprintf("localhost:%d", *grpcPort),
			Interval: "10s",
			Timeout:  "30s",
		},
	}

	err = consul.Agent().ServiceRegister(registration)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("server running on localhost:%d", *grpcPort)
	err = baseServer.Serve(grpcListener)
	if err != nil {
		log.Print("cannot start grpc server :", err)
		os.Exit(1)
	}
}
