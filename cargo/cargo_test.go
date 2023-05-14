package cargo

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/stretchr/testify/require"
)

func createNewCargo(t *testing.T) *Cargo {
	rs := RouteSpecification{
		Origin:          location.UNLocode("IDJKT"),
		Destination:     location.UNLocode("IDSLO"),
		ArrivalDeadline: time.Now(),
	}

	trackingID := NextTrackingID()
	c := New(trackingID, rs)
	nc, err := cargoTest.Upsert(context.Background(), dbTest, c)

	require.NoError(t, err)
	require.Equal(t, trackingID, nc.TrackingID)
	require.Empty(t, c.Itinerary)
	require.Equal(t, NotHandled, nc.Delivery.LastEvent.Activity.Type)

	return nc
}

func TestInsertCargo(t *testing.T) {
	createNewCargo(t)
}

func TestUpdateCargo(t *testing.T) {
	c := createNewCargo(t)
	newItinerary := Itinerary{
		ID: c.Itinerary.ID,
		Legs: []Leg{
			{
				VoyageNumber:   "234525",
				LoadLocation:   "IDJKT",
				UnloadLocation: "IDBDG",
				LoadTime:       time.Date(2023, 5, 8, 0, 0, 0, 0, time.UTC),
				UnloadTime:     time.Date(2023, 5, 9, 0, 0, 0, 0, time.UTC),
			},
			{
				VoyageNumber:   "234525",
				LoadLocation:   "IDBDG",
				UnloadLocation: "IDSLO",
				LoadTime:       time.Date(2023, 5, 9, 0, 0, 0, 0, time.UTC),
				UnloadTime:     time.Date(2023, 10, 9, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	// event := HandlingEvent{
	// 	TrackingID: c.TrackingID,
	// 	Activity: HandlingActivity{
	// 		Type:     Receive,
	// 		Location: location.UNLocode("IDJKT"),
	// 	},
	// }

	c.Itinerary = newItinerary
	// TODO: create new delivery with new event
	// newDel := newDelivery(event, c.Itinerary, c.RouteSpecification)
	// newDel.ID = c.Delivery.ID
	// c.Delivery = newDel

	nc, err := cargoTest.Upsert(context.Background(), dbTest, c)
	require.NoError(t, err)
	require.Equal(t, c.TrackingID, nc.TrackingID)
	require.Equal(t, c.Itinerary.ID, nc.Itinerary.ID)
	require.Equal(t, c.Delivery.ID, nc.Delivery.ID)
	require.Equal(t, 2, len(nc.Itinerary.Legs))

	// TODO: check LastEvent and NextExpectedActivity but to reach this point delivery
	// requires event_id, so we need to create event before updating deliveries.last_event
	// require.Equal(t, Receive, nc.Delivery.LastEvent.Activity.Type)
	// require.Equal(t, Load, nc.Delivery.NextExpectedActivity.Type)
}

func TestFindCargo(t *testing.T) {
	c := createNewCargo(t)
	nc, err := cargoTest.Find(context.Background(), dbTest, c.TrackingID)
	require.NoError(t, err)
	require.NotNil(t, nc)
	require.Equal(t, c.TrackingID, nc.TrackingID)
}

func TestFindCargoNotFound(t *testing.T) {
	c, err := cargoTest.Find(context.Background(), dbTest, "hfjhskjghshkj")
	require.Error(t, err)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.Nil(t, c)
}

func TestFindAllCargo(t *testing.T) {
	cs, err := cargoTest.FindAll(context.Background(), dbTest)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(cs), 1)
}
