package cargo

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInsertItinerary(t *testing.T) {
	createNewItinerary(t)
}

func TestUpdateItinerary(t *testing.T) {
	it := createNewItinerary(t)
	newLegs := []Leg{
		{
			VoyageNumber:   "234525",
			LoadLocation:   "IDBGR",
			UnloadLocation: "IDSMG",
			LoadTime:       time.Date(2023, 5, 8, 0, 0, 0, 0, time.UTC),
			UnloadTime:     time.Date(2023, 5, 9, 0, 0, 0, 0, time.UTC),
		},
		{
			VoyageNumber:   "234525",
			LoadLocation:   "IDSMG",
			UnloadLocation: "IDYKT",
			LoadTime:       time.Date(2023, 5, 9, 0, 0, 0, 0, time.UTC),
			UnloadTime:     time.Date(2023, 10, 9, 0, 0, 0, 0, time.UTC),
		},
	}
	it.Legs = newLegs

	i, err := itineraryTest.Upsert(context.Background(), dbTest, it)
	require.NoError(t, err)
	require.NotEmpty(t, i)

	for idx, leg := range it.Legs {
		expectedLeg := newLegs[idx]
		require.Equal(t, expectedLeg.LoadLocation, leg.LoadLocation)
		require.Equal(t, expectedLeg.UnloadLocation, leg.UnloadLocation)
	}
}

func createNewItinerary(t *testing.T) Itinerary {
	legs := []Leg{
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
	}

	it := Itinerary{Legs: legs}
	i, err := itineraryTest.Upsert(context.Background(), dbTest, it)
	require.NoError(t, err)
	require.NotEmpty(t, i)

	return i
}

func TestFindItinerary(t *testing.T) {
	it := createNewItinerary(t)
	newIt, err := itineraryTest.Find(context.Background(), dbTest, it.ID)
	require.NoError(t, err)
	require.NotEmpty(t, newIt)
}
