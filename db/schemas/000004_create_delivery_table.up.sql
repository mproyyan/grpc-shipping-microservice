CREATE TABLE IF NOT EXISTS deliveries (
    id BIGSERIAL PRIMARY KEY,
    itinerary_id BIGINT REFERENCES itineraries (id) NOT NULL,
    origin VARCHAR(5) NOT NULL,
    destination VARCHAR(5) NOT NULL,
    arrival_deadline TIMESTAMPTZ,
    routing_status INT NOT NULL DEFAULT 0,
    transport_status INT NOT NULL DEFAULT 0,
    last_event BIGINT REFERENCES events (id),
    current_voyage VARCHAR(10),
    eta TIMESTAMPTZ,
    is_misdirected BOOLEAN NOT NULL DEFAULT false,
    is_unloaded_at_destination BOOLEAN NOT NULL DEFAULT false
)