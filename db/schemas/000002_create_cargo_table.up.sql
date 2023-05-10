CREATE TABLE IF NOT EXISTS cargos (
    tracking_id VARCHAR(10) PRIMARY KEY,
    origin VARCHAR(5) NOT NULL, 
    destination VARCHAR(5) NOT NULL,
    arrival_deadline TIMESTAMPTZ,
    itinerary_id BIGINT,
    delivery_id BIGINT,
    FOREIGN KEY (itinerary_id) REFERENCES itineraries (id)
);