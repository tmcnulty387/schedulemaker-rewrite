-- +goose Up
CREATE TABLE buildings (
    `number`      VARCHAR(5) PRIMARY KEY,
    `code`        VARCHAR(5) UNIQUE,
    `name`        VARCHAR(100),
    `off_campus`  BOOLEAN DEFAULT TRUE
) Engine=InnoDb;

-- +goose Down
DROP TABLE IF EXISTS buildings;
