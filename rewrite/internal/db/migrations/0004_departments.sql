-- +goose Up
CREATE TABLE departments (
  `id`      INT UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `school`  INT UNSIGNED NOT NULL,
  `number`  SMALLINT(4) UNSIGNED ZEROFILL NULL DEFAULT NULL,
  `code`    VARCHAR(4) NULL DEFAULT NULL,
  `title`   VARCHAR(100) NOT NULL,
  `qtrnums` VARCHAR(20) NULL DEFAULT NULL
) Engine=InnoDb;

ALTER TABLE `departments`
  ADD CONSTRAINT UQ_departments_number
  UNIQUE (`number`);

ALTER TABLE `departments`
  ADD CONSTRAINT UQ_departments_code
  UNIQUE (`code`);

ALTER TABLE departments ADD INDEX departments(school);
ALTER TABLE departments ADD CONSTRAINT fk_school FOREIGN KEY departments(school)
  REFERENCES schools(id)
  ON UPDATE CASCADE
  ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS departments;
