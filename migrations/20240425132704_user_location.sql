-- +goose Up
CREATE TABLE IF NOT EXISTS user_location (
                           id bigint AUTO_INCREMENT PRIMARY KEY,
                           user_id bigint unsigned NOT NULL,
                           latitude DECIMAL(10, 8) NOT NULL,
                           longitude DECIMAL(11, 8) NOT NULL,
                           created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                           FOREIGN KEY (user_id) REFERENCES users (id)
) DEFAULT CHARSET=utf8;
-- +goose Down
drop table user_location;