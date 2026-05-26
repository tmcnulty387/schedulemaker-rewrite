-- +goose Up
CREATE TABLE schedulecourses (
  `id`        INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `schedule`  INT UNSIGNED NOT NULL,
  `section`   INT UNSIGNED NOT NULL
) ENGINE=InnoDb;

ALTER TABLE `schedulecourses`
  ADD FOREIGN KEY FK_schedcourses_schedule(`schedule`)
  REFERENCES `schedules`(`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE;

ALTER TABLE `schedulecourses`
  ADD FOREIGN KEY FK_schedcourses_section(`section`)
  REFERENCES `sections`(`id`)
  ON DELETE CASCADE
  ON UPDATE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS schedulecourses;
