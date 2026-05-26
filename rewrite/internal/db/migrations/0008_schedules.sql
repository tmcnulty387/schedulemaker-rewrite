-- +goose Up
CREATE TABLE schedules (
  `id`                INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `oldid`             VARCHAR(7) NULL DEFAULT NULL COLLATE latin1_general_cs,
  `datelastaccessed`  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `startday`          TINYINT(1) UNSIGNED NOT NULL DEFAULT 1,
  `endday`            TINYINT(1) UNSIGNED NOT NULL DEFAULT 6,
  `starttime`         SMALLINT(4) UNSIGNED ZEROFILL NOT NULL DEFAULT 0480,
  `endtime`           SMALLINT(4) UNSIGNED ZEROFILL NOT NULL DEFAULT 1320,
  `building`          SET('code', 'number') NOT NULL DEFAULT 'number',
  `quarter`           SMALLINT(5) UNSIGNED NULL DEFAULT NULL,
  `image`             BOOL NOT NULL DEFAULT FALSE
) ENGINE=InnoDb;

ALTER TABLE schedules ADD INDEX (`oldid`);

ALTER TABLE `schedules`
  ADD FOREIGN KEY FK_schedules_quarter(`quarter`)
  REFERENCES `quarters`(`quarter`)
  ON UPDATE CASCADE
  ON DELETE SET NULL;

-- +goose Down
DROP TABLE IF EXISTS schedules;
