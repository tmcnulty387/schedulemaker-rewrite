-- +goose Up
CREATE TABLE schedulenoncourses (
  `id`        INT UNSIGNED NOT NULL PRIMARY KEY,
  `schedule`  INT UNSIGNED NOT NULL,
  `title`     VARCHAR(30) NOT NULL,
  `day`       TINYINT(1) UNSIGNED NOT NULL,
  `start`     SMALLINT(4) UNSIGNED NOT NULL,
  `end`       SMALLINT(4) UNSIGNED NOT NULL
) ENGINE=InnoDb;

ALTER TABLE `schedulenoncourses`
  ADD FOREIGN KEY FK_schednoncourses_schedule(`schedule`)
  REFERENCES `schedules`(`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS schedulenoncourses;
