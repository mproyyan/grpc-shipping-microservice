CREATE TABLE IF NOT EXISTS events (
    id BIGSERIAL PRIMARY KEY,
    tracking_id VARCHAR(10) NOT NULL REFERENCES cargos (tracking_id),
    event_type INT NOT NULL,
    location VARCHAR(5) NOT NULL,
    voyage_number VARCHAR(10)
);