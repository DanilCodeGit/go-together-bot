-- +goose Up
create table if not exists passengers(
    name VARCHAR(50) not null,
    event_id bigint unsigned NOT NULL,
    user_chatId bigint not null,
    FOREIGN KEY (event_id) REFERENCES events (id_events)
);

-- +goose Down
DROP table passengers;
