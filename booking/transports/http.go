package transports

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	ht "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/mproyyan/grpc-shipping-microservice/booking/endpoints"
	"github.com/mproyyan/grpc-shipping-microservice/booking/services"
	"github.com/mproyyan/grpc-shipping-microservice/cargo"
)

var errBadRoute = errors.New("bad route")

func NewHttpHandler(ep endpoints.Set) http.Handler {
	bookNewCargoHandler := ht.NewServer(
		ep.BookNewCargoEndpoint,
		decodeHttpBookNewCargoRequest,
		encodeGenericResponse,
	)

	loadCargoHandler := ht.NewServer(
		ep.LoadCargoEndpoint,
		decodeLoadCargoRequest,
		encodeGenericResponse,
	)

	assignCargoToRouteHandler := ht.NewServer(
		ep.AssignCargoToRouteEndpoint,
		decodeAssignCargoToRouteRequest,
		encodeGenericResponse,
	)

	changeDestinationHandler := ht.NewServer(
		ep.ChangeDestinationEndpoint,
		decodeChangeDestinationRequest,
		encodeGenericResponse,
	)

	listCargoHandler := ht.NewServer(
		ep.CargosEndpoint,
		decodeListCargoRequest,
		encodeGenericResponse,
	)

	r := mux.NewRouter()
	r.Handle("/booking/cargos", bookNewCargoHandler).Methods("POST")
	r.Handle("/booking/cargos", listCargoHandler).Methods("GET")
	r.Handle("/booking/cargos/{id}", loadCargoHandler).Methods("GET")
	r.Handle("/booking/cargos/{id}/assign_route", assignCargoToRouteHandler).Methods("POST")
	r.Handle("/booking/cargos/{id}/change_destination", changeDestinationHandler).Methods("POST")

	return r
}

// server
// book new cargo
func decodeHttpBookNewCargoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req endpoints.BookNewCargoRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// Load cargo
func decodeLoadCargoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}

	return endpoints.LoadCargoRequest{TrackingID: cargo.TrackingID(id)}, nil
}

// assign cargo to route
func decodeAssignCargoToRouteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}

	var itinerary cargo.Itinerary
	if err := json.NewDecoder(r.Body).Decode(&itinerary); err != nil {
		return nil, err
	}

	return endpoints.AssignCargoToRouteRequest{
		TrackingID: cargo.TrackingID(id),
		Itinerary:  itinerary,
	}, nil
}

// change destination
func decodeChangeDestinationRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		return nil, errBadRoute
	}

	var req endpoints.ChangeDestinationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	req.TrackingID = cargo.TrackingID(id)
	return req, nil
}

// list cargo
func decodeListCargoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return endpoints.ListCargosRequest{}, nil
}

type errorer interface {
	error() error
}

func encodeGenericResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case cargo.ErrUnknown:
		w.WriteHeader(http.StatusNotFound)
	case services.ErrInvalidArgument:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
