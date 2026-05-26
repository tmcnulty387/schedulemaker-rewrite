-- STATUS

-- name: GetScrapeLog :many
SELECT id, timeStarted, timeEnded, quartersAdded, coursesAdded, coursesUpdated,
       sectionsAdded, sectionsUpdated, failures
FROM scrapelog
ORDER BY timeStarted DESC
LIMIT 20;

-- TERMS / QUARTERS

-- name: GetTerms :many
SELECT quarter FROM quarters ORDER BY quarter DESC;

-- name: GetQuarterDates :one
SELECT start, end FROM quarters WHERE quarter = ?;

-- SCHOOLS

-- name: GetSchools :many
SELECT id, number, code, title FROM schools;

-- name: GetSchoolsBySemesterTerm :many
-- Used when term > 20130.
SELECT id, code, title FROM schools WHERE code IS NOT NULL ORDER BY code;

-- name: GetSchoolsByQuarterTerm :many
-- Used when term <= 20130.
SELECT id, number AS code, title FROM schools WHERE number IS NOT NULL ORDER BY number;

-- DEPARTMENTS

-- name: GetDepartmentsBySemesterTerm :many
-- Used when term > 20130: returns code, concatenates numbers.
SELECT id, title, code, GROUP_CONCAT(number, ', ') AS number
FROM departments AS d
WHERE school = ?
  AND (SELECT COUNT(*) FROM courses AS c WHERE c.department = d.id AND quarter = ?) >= 1
  AND code IS NOT NULL
GROUP BY code
ORDER BY code;

-- name: GetDepartmentsByQuarterTerm :many
-- Used when term <= 20130: returns number only.
SELECT id, title, number
FROM departments AS d
WHERE school = ?
  AND (SELECT COUNT(*) FROM courses AS c WHERE c.department = d.id AND quarter = ?) >= 1
  AND number IS NOT NULL
ORDER BY id;

-- COURSES

-- name: GetCoursesByDepartmentAndTerm :many
-- Non-cancelled courses for a department and term (getCourses endpoint).
SELECT c.title, c.course, c.description, c.id, d.number, d.code
FROM sections AS s
JOIN courses AS c ON s.course = c.id
JOIN departments AS d ON d.id = c.department
WHERE c.department = ?
  AND quarter = ?
  AND s.status != 'X'
GROUP BY c.id
ORDER BY course;

-- SECTIONS

-- name: GetSectionByID :one
-- Used by getCourseBySectionId (without description).
SELECT s.id,
       CASE WHEN s.title != '' THEN s.title ELSE c.title END AS title,
       c.id AS courseId,
       s.instructor, s.curenroll, s.maxenroll, s.type,
       c.quarter, c.credits, c.course, s.section,
       d.number, d.code
FROM sections AS s
JOIN courses AS c ON s.course = c.id
JOIN departments AS d ON d.id = c.department
WHERE s.id = ?;

-- name: GetSectionByIDWithDescription :one
-- Used by getCourseBySectionId with description=true (courseForSection endpoint).
SELECT s.id,
       CASE WHEN s.title != '' THEN s.title ELSE c.title END AS title,
       c.id AS courseId,
       s.instructor, s.curenroll, s.maxenroll, s.type,
       c.quarter, c.credits, c.course, c.description, s.section,
       d.number, d.code
FROM sections AS s
JOIN courses AS c ON s.course = c.id
JOIN departments AS d ON d.id = c.department
WHERE s.id = ?;

-- name: GetSectionsByCourse :many
-- Used by getSections: all non-cancelled sections with course and department info.
SELECT c.title AS coursetitle, c.course, d.number, d.code, s.section,
       s.instructor, s.id, s.type, s.maxenroll, s.curenroll, s.title AS sectiontitle
FROM sections AS s
JOIN courses AS c ON s.course = c.id
JOIN departments AS d ON d.id = c.department
WHERE s.course = ?
  AND s.status != 'X'
ORDER BY c.course, s.section;

-- name: GetSectionByCourseSemesterTerm :one
-- Used by getCourse when term > 20130: joins department by code.
SELECT s.id,
       CASE WHEN s.title != '' THEN s.title ELSE c.title END AS title,
       s.instructor, s.curenroll, s.maxenroll, s.type,
       d.code AS department, c.course, c.credits, s.section
FROM sections AS s
JOIN courses AS c ON c.id = s.course
JOIN departments AS d ON d.id = c.department
WHERE c.quarter = ?
  AND d.code = ?
  AND c.course = ?
  AND s.section = ?;

-- name: GetSectionByCourseQuarterTerm :one
-- Used by getCourse when term <= 20130: joins department by number.
SELECT s.id,
       CASE WHEN s.title != '' THEN s.title ELSE c.title END AS title,
       s.instructor, s.curenroll, s.maxenroll, s.type,
       d.number AS department, c.course, c.credits, s.section
FROM sections AS s
JOIN courses AS c ON c.id = s.course
JOIN departments AS d ON d.id = c.department
WHERE c.quarter = ?
  AND d.number = ?
  AND c.course = ?
  AND s.section = ?;

-- TIMES

-- name: GetTimesBySection :many
-- Used by getMeetingInfo and the getSections time lookup.
SELECT t.day, t.start, t.end, b.code, b.number, b.off_campus, t.room
FROM times AS t
JOIN buildings AS b ON b.number = t.building
WHERE t.section = ?
ORDER BY t.day, t.start;

-- SCHEDULES

-- name: GetSchedule :one
SELECT startday, endday, starttime, endtime, building, quarter,
       CAST(image AS unsigned) AS image
FROM schedules
WHERE id = ?;

-- name: GetScheduleByOldID :one
SELECT id FROM schedules WHERE oldid = ?;

-- name: TouchSchedule :exec
-- Updates datelastaccessed; called before every GET to confirm existence and record access.
UPDATE schedules SET datelastaccessed = NOW() WHERE id = ?;

-- name: GetScheduleCourses :many
SELECT section FROM schedulecourses WHERE schedule = ?;

-- name: GetScheduleNonCourses :many
SELECT id, title, day, start, end FROM schedulenoncourses WHERE schedule = ?;

-- name: InsertSchedule :execresult
INSERT INTO schedules (oldid, startday, endday, starttime, endtime, building, quarter)
VALUES ('', ?, ?, ?, ?, ?, ?);

-- name: SetScheduleImage :exec
UPDATE schedules SET image = 1 WHERE id = ?;

-- name: InsertScheduleCourse :exec
INSERT INTO schedulecourses (schedule, section) VALUES (?, ?);

-- name: InsertScheduleNonCourse :exec
INSERT INTO schedulenoncourses (title, day, start, end, schedule) VALUES (?, ?, ?, ?, ?);

