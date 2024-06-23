package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	consulsd "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	kitlog "github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/hashicorp/consul/api"
	be "github.com/mproyyan/grpc-shipping-microservice/booking/endpoints"
	bs "github.com/mproyyan/grpc-shipping-microservice/booking/services"
	bt "github.com/mproyyan/grpc-shipping-microservice/booking/transports"
	"google.golang.org/grpc"
)

func main() {
	var (
		httpAddr     = flag.String("http.addr", ":8000", "Address for HTTP (JSON) server")
		consulAddr   = flag.String("consul.addr", "", "Consul agent address")
		retryMax     = flag.Int("retry.max", 3, "per-request retries to different instances")
		retryTimeout = flag.Duration("retry.timeout", 500*time.Millisecond, "per-request timeout, including retries")
	)

	flag.Parse()

	var logger kitlog.Logger
	{
		logger = kitlog.NewLogfmtLogger(os.Stderr)
		logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
		logger = kitlog.With(logger, "caller", kitlog.DefaultCaller)
	}

	// Service discovery domain. In this example we use Consul.
	var client consulsd.Client
	{
		consulConfig := api.DefaultConfig()
		if len(*consulAddr) > 0 {
			consulConfig.Address = *consulAddr
		}
		consulClient, err := api.NewClient(consulConfig)
		if err != nil {
			log.Fatal(err)
		}

		client = consulsd.NewClient(consulClient)
	}

	var (
		tags        = []string{}
		passingOnly = true
		endpoints   = be.Set{}
		instancer   = consulsd.NewInstancer(client, logger, "bookingservice", tags, passingOnly)
	)

	{
		// book new cargo
		factory := bookingServiceFactory(be.MakeBookNewCargoEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		endpoints.BookNewCargoEndpoint = retry
	}
	{
		// load cargo
		factory := bookingServiceFactory(be.MakeLoadCargoEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		endpoints.LoadCargoEndpoint = retry
	}
	{
		// assign cargo to route
		factory := bookingServiceFactory(be.MakeAssignCargoToRouteEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		endpoints.AssignCargoToRouteEndpoint = retry
	}
	{
		// change destination
		factory := bookingServiceFactory(be.MakeChangeDestinationEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		endpoints.ChangeDestinationEndpoint = retry
	}
	{
		// list all cargos
		factory := bookingServiceFactory(be.MakeListCargosEndpoint)
		endpointer := sd.NewEndpointer(instancer, factory, logger)
		balancer := lb.NewRoundRobin(endpointer)
		retry := lb.Retry(*retryMax, *retryTimeout, balancer)
		endpoints.CargosEndpoint = retry
	}

	r := mux.NewRouter()
	r.PathPrefix("/booking").Handler(bt.NewHttpHandler(endpoints))

	err := r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Println("Route:", pathTemplate)
		}

		return nil
	})

	if err != nil {
		log.Fatal("route walk error", err)
	}

	// Interrupt handler.
	errc := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errc <- fmt.Errorf("%s", <-c)
	}()

	// HTTP transport.
	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errc <- http.ListenAndServe(*httpAddr, r)
	}()

	// Run!
	logger.Log("exit", <-errc)
}

func bookingServiceFactory(makeEndpoint func(bookingService bs.BookingServiceContract) endpoint.Endpoint) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}

		service := bt.NewGRPCClient(conn)
		endpoint := makeEndpoint(service)

		return endpoint, conn, nil
	}
}
