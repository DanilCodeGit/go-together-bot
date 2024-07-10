-- +goose Up
CREATE TABLE IF NOT EXISTS events (
                                      id_events serial PRIMARY KEY AUTO_INCREMENT ,
                                      date_of_trip DATE,
                                      available_seats INT,
                                      trip_cost DECIMAL(10, 2),
                                      cost_per_person DECIMAL(10, 2),
                                      departure_address VARCHAR(255),
                                      arrival_address VARCHAR(255),
                                      car_number VARCHAR(20),
                                      user_id bigint unsigned,
                                      driver_name VARCHAR(50),
                                      status bool default true,
                                      FOREIGN KEY (user_id) REFERENCES users(id)
);
-- +goose Down
drop table if exists events;