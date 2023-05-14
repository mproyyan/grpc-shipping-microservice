package cargo

import (
	"database/sql"
	"os"
	"testing"

	"github.com/mproyyan/grpc-shipping-microservice/config"
	"github.com/mproyyan/grpc-shipping-microservice/db"
)

var (
	dbTest        *sql.DB
	itineraryTest ItineraryRepositoryContract
	deliveryTest  DeliveryRepositoryContract
	cargoTest     CargoRepositoryContract
)

func TestMain(m *testing.M) {
	env := config.Environment{
		DBUsername: "postgres",
		DBPassword: "ligmaballs",
		DBHost:     "localhost",
		DBPort:     "5432",
		DBName:     "grpc_shipping",
	}

	dbTest, _ = db.NewPostgreSQL(env).Connect()

	itineraryTest = ItineraryRepository{}
	deliveryTest = DeliveryRepository{}
	cargoTest = CargoRepository{ItineraryRepository: itineraryTest, DeliveryRepository: deliveryTest}

	os.Exit(m.Run())
}
