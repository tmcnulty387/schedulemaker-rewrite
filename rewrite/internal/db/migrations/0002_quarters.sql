-- +goose Up
CREATE TABLE quarters (
  `quarter`     SMALLINT(5) UNSIGNED NOT NULL PRIMARY KEY,
  `start`       DATE NOT NULL,
  `end`         DATE NOT NULL
) ENGINE=InnoDb;

-- +goose Down
DROP TABLE IF EXISTS quarters;
