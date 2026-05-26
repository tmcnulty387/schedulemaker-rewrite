-- +goose Up
CREATE TABLE courses (
  `id`          INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
  `quarter`     SMALLINT UNSIGNED NOT NULL,
  `department`  INT UNSIGNED NOT NULL,
  `course`      VARCHAR(4) NOT NULL,
  `credits`     TINYINT(2) UNSIGNED NOT NULL DEFAULT 0,
  `title`       VARCHAR(50) NOT NULL,
  `description` TEXT NOT NULL
) ENGINE=InnoDb;

ALTER TABLE `courses`
    ADD CONSTRAINT UQ_courses_quarter_department_course
    UNIQUE (`quarter`, `department`, `course`);

ALTER TABLE `courses`
    ADD FOREIGN KEY FK_courses_quarter(`quarter`)
    REFERENCES `quarters`(`quarter`)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

ALTER TABLE `courses`
    ADD FOREIGN KEY FK_courses_dept(`department`)
    REFERENCES `departments`(`id`)
    ON DELETE CASCADE
    ON UPDATE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS courses;
