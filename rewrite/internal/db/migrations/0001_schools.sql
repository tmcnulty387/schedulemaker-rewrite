-- +goose Up
CREATE TABLE IF NOT EXISTS `schools` (
  `id`      INT UNSIGNED NOT NULL PRIMARY KEY,
  `number`  tinyint(2) UNSIGNED ZEROFILL NULL DEFAULT NULL,
  `code`    VARCHAR(5) NULL DEFAULT NULL,
  `title` varchar(30) NOT NULL
) ENGINE=InnoDb;

ALTER TABLE `schools`
  ADD CONSTRAINT UQ_schools_number_code
  UNIQUE (`number`, `code`);

-- +goose Down
DROP TABLE IF EXISTS `schools`;
