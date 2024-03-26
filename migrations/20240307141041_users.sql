-- +goose Up
create table users(
  id serial primary key ,
  name varchar not null,
  phone varchar not null unique,
  login varchar not null unique ,
  password varchar,
  chatID integer
);
-- +goose Down
drop table users;