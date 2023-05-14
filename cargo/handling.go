package cargo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/mproyyan/grpc-shipping-microservice/db"
	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/mproyyan/grpc-shipping-microservice/voyage"
)

// HandlingActivity represents how and where a cargo can be handled, and can
// be used to express predictions about what is expected to happen to a cargo
// in the future.
type HandlingActivity struct {
	Type         HandlingEventType
	Location     location.UNLocode
	VoyageNumber voyage.Number
}

// HandlingEvent is used to register the event when, for instance, a cargo is
// unloaded from a carrier at a some location at a given time.
type HandlingEvent struct {
	ID         int64
	TrackingID TrackingID
	Activity   HandlingActivity
}

// HandlingEventType describes type of a handling event.
type HandlingEventType int

// Valid handling event types.
const (
	NotHandled HandlingEventType = iota
	Load
	Unload
	Receive
	Claim
	Customs
)

func (t HandlingEventType) String() string {
	switch t {
	case NotHandled:
		return "Not Handled"
	case Load:
		return "Load"
	case Unload:
		return "Unload"
	case Receive:
		return "Receive"
	case Claim:
		return "Claim"
	case Customs:
		return "Customs"
	}

	return ""
}

// HandlingHistory is the handling history of a cargo.
type HandlingHistory struct {
	HandlingEvents []HandlingEvent
}

// MostRecentlyCompletedEvent returns most recently completed handling event.
func (h HandlingHistory) MostRecentlyCompletedEvent() (HandlingEvent, error) {
	if len(h.HandlingEvents) == 0 {
		return HandlingEvent{}, errors.New("delivery history is empty")
	}

	return h.HandlingEvents[len(h.HandlingEvents)-1], nil
}

type EventRepositoryContract interface {
	Store(ctx context.Context, dbtx db.DBTX, e HandlingEvent) (HandlingEvent, error)
	QueryHandlingHistory(ctx context.Context, dbtx db.DBTX, id TrackingID) (HandlingHistory, error)
}

type EventRepository struct {
}

type eventResult struct {
	id           sql.NullInt64
	trackingId   sql.NullString
	eventType    sql.NullInt32
	location     sql.NullString
	voyageNumber sql.NullString
}

func (er eventResult) build() HandlingEvent {
	if !er.id.Valid {
		return HandlingEvent{}
	}

	return HandlingEvent{
		TrackingID: TrackingID(er.trackingId.String),
		Activity: HandlingActivity{
			Type:         HandlingEventType(er.eventType.Int32),
			Location:     location.UNLocode(er.location.String),
			VoyageNumber: voyage.Number(er.voyageNumber.String),
		},
	}
}

func (er EventRepository) Store(ctx context.Context, dbtx db.DBTX, e HandlingEvent) (HandlingEvent, error) {
	query := `
		INSERT INTO events (tracking_id, event_type, location, voyage_number)
		VALUES ($1, $2, $3, $4)
		RETURNING id, tracking_id, event_type, location, voyage_number
	`

	var result eventResult
	row := dbtx.QueryRowContext(ctx, query, e.TrackingID, e.Activity.Type, e.Activity.Location, e.Activity.VoyageNumber)
	err := row.Scan(&result.id, &result.trackingId, &result.eventType, &result.location, &result.voyageNumber)
	if err != nil {
		return HandlingEvent{}, err
	}

	return result.build(), nil
}

func (er EventRepository) QueryHandlingHistory(ctx context.Context, dbtx db.DBTX, id TrackingID) (HandlingHistory, error) {
	query := `
		SELECT id, tracking_id, event_type, location, voyage_number FROM events
		WHERE tracking_id = $1
	`

	var result eventResult
	var handlinghistory HandlingHistory
	row, err := dbtx.QueryContext(ctx, query, id)
	if err != nil {
		return handlinghistory, nil
	}

	for row.Next() {
		err := row.Scan(
			&result.id,
			&result.trackingId,
			&result.eventType,
			&result.location,
			&result.voyageNumber,
		)

		if err != nil {
			return handlinghistory, err
		}

		event := result.build()
		handlinghistory.HandlingEvents = append(handlinghistory.HandlingEvents, event)
	}

	return handlinghistory, nil
}
