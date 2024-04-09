-- +goose Up
CREATE TABLE users (
    id bigint unsigned auto_increment PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    phone VARCHAR(20) NOT NULL UNIQUE,
    login VARCHAR(50) NOT NULL UNIQUE,
    chatID INT
);

-- +goose Down
drop table users;
