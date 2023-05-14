package cargo

import (
	"context"
	"testing"

	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/mproyyan/grpc-shipping-microservice/voyage"
	"github.com/stretchr/testify/require"
)

func createNewEvent(t *testing.T, c *Cargo, eventType HandlingEventType, loc, voyageNumber string) (HandlingEvent, *Cargo) {
	if c == nil {
		c = createNewCargo(t)
	}

	e := HandlingEvent{
		TrackingID: c.TrackingID,
		Activity: HandlingActivity{
			Type:         eventType,
			Location:     location.UNLocode(loc),
			VoyageNumber: voyage.Number(voyageNumber),
		},
	}

	e, err := eventTest.Store(context.Background(), dbTest, e)
	require.NoError(t, err)
	require.Equal(t, c.TrackingID, e.TrackingID)
	require.Equal(t, location.UNLocode(loc), e.Activity.Location)
	require.Equal(t, voyage.Number(voyageNumber), e.Activity.VoyageNumber)

	return e, c
}

func TestStoreEvent(t *testing.T) {
	createNewEvent(t, nil, Receive, "IDJKT", "")
}

func TestQueryEventHistory(t *testing.T) {
	_, c := createNewEvent(t, nil, Receive, "IDJKT", "")
	createNewEvent(t, c, Load, "IDJKT", "")

	h, err := eventTest.QueryHandlingHistory(context.Background(), dbTest, c.TrackingID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(h.HandlingEvents), 1)
}
