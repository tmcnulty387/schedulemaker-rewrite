-- +goose Up
CREATE TABLE sections (
  `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `course`      INT UNSIGNED NOT NULL,
  `section`     VARCHAR(4) NOT NULL,
  `title`       VARCHAR(30) NOT NULL,
  `type`        ENUM('R','N','OL','H', 'BL') NOT NULL DEFAULT 'R',
  `status`      ENUM('O','C','X') NOT NULL,
  `instructor`  VARCHAR(64) NOT NULL DEFAULT 'TBA',
  `maxenroll`   SMALLINT(3) UNSIGNED NOT NULL,
  `curenroll`   SMALLINT(3) UNSIGNED NOT NULL
) ENGINE=InnoDB;

ALTER TABLE sections
    ADD CONSTRAINT UQ_sections_course_section
    UNIQUE (`course`, `section`);

ALTER TABLE sections
    ADD CONSTRAINT FK_sections_course
    FOREIGN KEY (`course`)
    REFERENCES courses(`id`)
    ON UPDATE CASCADE
    ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS sections;
