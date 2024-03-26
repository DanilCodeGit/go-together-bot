-- +goose Up
create table events(
                       id serial primary key ,
                       date date not null default current_date,
                       driverName varchar not null,
                       departureTime date not null  default current_date,
                       freeSpace integer not null,
                       price real not null
);
-- +goose Down
drop table events;