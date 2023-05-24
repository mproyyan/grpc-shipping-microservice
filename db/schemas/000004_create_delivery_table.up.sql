CREATE TABLE IF NOT EXISTS deliveries (
    id BIGSERIAL PRIMARY KEY,
    itinerary_id BIGINT REFERENCES itineraries (id) NOT NULL,
    origin VARCHAR(5) NOT NULL,
    destination VARCHAR(5) NOT NULL,
    arrival_deadline TIMESTAMPTZ,
    last_event BIGINT REFERENCES events (id)
)