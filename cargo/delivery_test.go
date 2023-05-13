package cargo

import (
	"context"
	"testing"
	"time"

	"github.com/mproyyan/grpc-shipping-microservice/location"
	"github.com/stretchr/testify/require"
)

func TestInsertDeliveries(t *testing.T) {
	i := createNewItinerary(t)
	d := createNewDelivery(t, i, HandlingEvent{})

	require.Equal(t, NotHandled, d.LastEvent.Activity.Type)
}

func createNewDelivery(t *testing.T, i Itinerary, e HandlingEvent) Delivery {
	delivery := Delivery{
		Itinerary: i,
		LastEvent: e,
		RouteSpecification: RouteSpecification{
			Origin:          location.UNLocode("IDJKT"),
			Destination:     location.UNLocode("IDSLO"),
			ArrivalDeadline: time.Now(),
		},
	}

	newDelivery, err := deliveryTest.Upsert(context.Background(), dbTest, delivery)
	require.NoError(t, err)
	require.Equal(t, delivery.Itinerary.ID, newDelivery.Itinerary.ID)

	return newDelivery
}

func TestFindDelivery(t *testing.T) {
	d := createNewDelivery(t, createNewItinerary(t), HandlingEvent{})
	nd, err := deliveryTest.Find(context.Background(), dbTest, d.ID)

	require.NoError(t, err)
	require.Equal(t, d.ID, nd.ID)
}

func TestFindDeliveryNotFound(t *testing.T) {
	nd, err := deliveryTest.Find(context.Background(), dbTest, 9999999)
	require.NoError(t, err)
	require.Empty(t, nd)
}
