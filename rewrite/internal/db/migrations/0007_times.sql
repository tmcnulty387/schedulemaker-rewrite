-- +goose Up
CREATE TABLE times (
  `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `section`     INT UNSIGNED NOT NULL,
  `day`         TINYINT(1) UNSIGNED NOT NULL,
  `start`       SMALLINT(4) UNSIGNED NOT NULL,
  `end`         SMALLINT(4) UNSIGNED NOT NULL,
  `building`    VARCHAR(5) NULL DEFAULT NULL,
  `room`        VARCHAR(10) NULL DEFAULT NULL
) ENGINE=InnoDB;

ALTER TABLE times
    ADD CONSTRAINT FK_times_section
    FOREIGN KEY (`section`)
    REFERENCES sections(`id`)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

ALTER TABLE times
    ADD CONSTRAINT FK_times_building
    FOREIGN KEY (`building`)
    REFERENCES buildings(`number`)
    ON DELETE SET NULL
    ON UPDATE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS times;
