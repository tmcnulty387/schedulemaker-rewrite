-- +goose Up
CREATE TABLE scrapelog (
  `id`              INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `timeStarted`     INT(11) NOT NULL,
  `timeEnded`       INT(11) NOT NULL,
  `quartersAdded`   TINYINT(3) UNSIGNED NOT NULL,
  `coursesAdded`    INT UNSIGNED NOT NULL,
  `coursesUpdated`  INT UNSIGNED NOT NULL,
  `sectionsAdded`   INT UNSIGNED NOT NULL,
  `sectionsUpdated` INT UNSIGNED NOT NULL,
  `failures`        INT UNSIGNED NOT NULL
) ENGINE=InnoDB;

-- +goose Down
DROP TABLE IF EXISTS scrapelog;
