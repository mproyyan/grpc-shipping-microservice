package cargo

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/db"
	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/mproyyan/grpc-shipping-microservice/voyage"
)

// Leg describes the transportation between two locations on a voyage.
type Leg struct {
	VoyageNumber   voyage.Number     `json:"voyage_number"`
	LoadLocation   location.UNLocode `json:"from"`
	UnloadLocation location.UNLocode `json:"to"`
	LoadTime       time.Time         `json:"load_time"`
	UnloadTime     time.Time         `json:"unload_time"`
}

// NewLeg creates a new itinerary leg.
func NewLeg(voyageNumber voyage.Number, loadLocation, unloadLocation location.UNLocode, loadTime, unloadTime time.Time) Leg {
	return Leg{
		VoyageNumber:   voyageNumber,
		LoadLocation:   loadLocation,
		UnloadLocation: unloadLocation,
		LoadTime:       loadTime,
		UnloadTime:     unloadTime,
	}
}

// Itinerary specifies steps required to transport a cargo from its origin to
// destination.
type Itinerary struct {
	ID   int64 `json:"id"`
	Legs []Leg `json:"legs"`
}

type ItineraryRepositoryContract interface {
	Upsert(ctx context.Context, dbtx db.DBTX, itinerary Itinerary) (Itinerary, error)
	Find(ctx context.Context, dbtx db.DBTX, id int64) (Itinerary, error)
}

type ItineraryRepository struct {
}

func (ir ItineraryRepository) Upsert(ctx context.Context, dbtx db.DBTX, itinerary Itinerary) (Itinerary, error) {
	b, err := json.Marshal(itinerary.Legs)
	if err != nil {
		return Itinerary{}, err
	}

	var result struct {
		id   int64
		legs string
	}

	var row *sql.Row
	if itinerary.ID == 0 {
		query := "INSERT INTO itineraries (legs) VALUES($1) RETURNING id, legs"
		row = dbtx.QueryRowContext(ctx, query, string(b))
	} else {
		query := "UPDATE itineraries SET legs = $2 WHERE id = $1 RETURNING id, legs"
		row = dbtx.QueryRowContext(ctx, query, itinerary.ID, string(b))
	}

	err = row.Scan(
		&result.id,
		&result.legs,
	)

	if err != nil {
		return Itinerary{}, err
	}

	legs, err := buildLegs([]byte(result.legs))
	if err != nil {
		return Itinerary{}, err
	}

	return Itinerary{
		ID:   result.id,
		Legs: legs,
	}, nil
}

func (ir ItineraryRepository) Find(ctx context.Context, dbtx db.DBTX, id int64) (Itinerary, error) {
	query := "SELECT id, legs FROM itineraries WHERE id = $1 LIMIT 1"
	row := dbtx.QueryRowContext(ctx, query, id)

	var result struct {
		id   int64
		legs string
	}

	err := row.Scan(
		&result.id,
		&result.legs,
	)

	if err != nil {
		return Itinerary{}, err
	}

	legs, err := buildLegs([]byte(result.legs))
	if err != nil {
		return Itinerary{}, err
	}

	return Itinerary{
		ID:   result.id,
		Legs: legs,
	}, nil
}

func buildLegs(legsByte []byte) ([]Leg, error) {
	legs := []Leg{{}}
	err := json.Unmarshal(legsByte, &legs)
	if err != nil {
		return []Leg{}, err
	}

	return legs, nil
}