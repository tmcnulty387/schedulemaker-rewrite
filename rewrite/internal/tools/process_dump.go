package tools

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"rewrite/internal/config"
	"strconv"
	"time"
)

const createClassesTable = `
CREATE TABLE IF NOT EXISTS classes (
  crse_id         INT UNSIGNED NOT NULL,
  crse_offer_nbr  INT UNSIGNED NOT NULL,
  strm            INT UNSIGNED NOT NULL,
  session_code    VARCHAR(4)   NOT NULL,
  class_section   VARCHAR(4)   NOT NULL,
  subject         VARCHAR(4)   NOT NULL,
  catalog_nbr     VARCHAR(4)   NOT NULL,
  descr           TEXT         NOT NULL,
  topic           TEXT         NOT NULL,
  class_nbr       INT UNSIGNED NOT NULL,
  ssr_component   VARCHAR(3)   NOT NULL,
  units           INT UNSIGNED NOT NULL,
  enrl_stat       VARCHAR(1)   NOT NULL,
  class_stat      VARCHAR(1)   NOT NULL,
  class_type      VARCHAR(1)   NOT NULL,
  schedule_print  VARCHAR(1)   NOT NULL,
  enrl_cap        INT UNSIGNED NOT NULL,
  enrl_tot        INT UNSIGNED NOT NULL,
  institution     VARCHAR(5)   NOT NULL,
  acad_org        VARCHAR(10)  NOT NULL,
  acad_group      VARCHAR(5)   NOT NULL,
  acad_career     VARCHAR(4)   NOT NULL,
  instruction_mode VARCHAR(2)  NOT NULL,
  course_descrlong TEXT        NOT NULL,
  PRIMARY KEY (crse_id, crse_offer_nbr, strm, session_code, class_section)
) ENGINE=MyISAM DEFAULT CHARSET=latin1`

const createMeetingTable = `
CREATE TABLE IF NOT EXISTS meeting (
  crse_id             INT  NOT NULL,
  crse_offer_nbr      INT  NOT NULL,
  strm                INT  NOT NULL,
  session_code        VARCHAR(4)  NOT NULL,
  class_section       VARCHAR(4)  NOT NULL,
  class_mtg_nbr       INT  NOT NULL,
  start_dt            DATE NOT NULL,
  end_dt              DATE NOT NULL,
  bldg                VARCHAR(10) NOT NULL,
  room_nbr            VARCHAR(10) NOT NULL,
  meeting_time_start  TIME NOT NULL,
  meeting_time_end    TIME NOT NULL,
  mon                 VARCHAR(1)  NOT NULL,
  tues                VARCHAR(1)  NOT NULL,
  wed                 VARCHAR(1)  NOT NULL,
  thurs               VARCHAR(1)  NOT NULL,
  fri                 VARCHAR(1)  NOT NULL,
  sat                 VARCHAR(1)  NOT NULL,
  sun                 VARCHAR(1)  NOT NULL,
  PRIMARY KEY (crse_id, crse_offer_nbr, strm, session_code, class_section, class_mtg_nbr),
  INDEX idx_meeting (crse_id, crse_offer_nbr, strm, session_code, class_section)
) ENGINE=MyISAM DEFAULT CHARSET=latin1`

const createInstructorsTable = `
CREATE TABLE IF NOT EXISTS instructors (
  crse_id        INT         NOT NULL,
  crse_offer_nbr INT         NOT NULL,
  strm           INT         NOT NULL,
  session_code   VARCHAR(4)  NOT NULL,
  class_section  VARCHAR(4)  NOT NULL,
  class_mtg_nbr  INT         NOT NULL,
  last_name      VARCHAR(30) NOT NULL,
  first_name     VARCHAR(30) NOT NULL,
  INDEX idx_instructors (crse_id, crse_offer_nbr, strm, session_code, class_section)
) ENGINE=MyISAM DEFAULT CHARSET=latin1`

func (p *Parser) ParseDumps(ctx context.Context, cfg *config.Config) {
	classFile, classSize := p.openDumpFile(ctx, cfg.DumpClasses)
	attrFile, _ := p.openDumpFile(ctx, cfg.DumpClassAttr)
	instrFile, instrSize := p.openDumpFile(ctx, cfg.DumpInstruct)
	meetFile, meetSize := p.openDumpFile(ctx, cfg.DumpMeeting)
	notesFile, _ := p.openDumpFile(ctx, cfg.DumpNotes)
	defer classFile.Close()
	defer attrFile.Close()
	defer instrFile.Close()
	defer meetFile.Close()
	defer notesFile.Close()

	// Build the temporary tables
	_, err := p.db.ExecContext(ctx, createClassesTable)
	if err != nil {
		p.Halt(ctx, "Error: Failed to create temporary class table ", err)
	}
	p.debug("... Temporary class table created successfully")

	p.fileToTempTable(ctx, "classes", classFile, 24, classSize, procClassArray)

	// Build a temporary table for the meeting patterns
	_, err = p.db.ExecContext(ctx, createMeetingTable)
	if err != nil {
		p.Halt(ctx, "Error: Failed to create temporary meeting pattern table", err)
	}
	p.debug("... Temporary meeting pattern table created successfully")

	p.fileToTempTable(ctx, "meeting", meetFile, 19, meetSize, procMeetArray)

	// Process the instructor file
	_, err = p.db.ExecContext(ctx, createInstructorsTable)
	if err != nil {
		p.Halt(ctx, "Error: Failed to create temporary instructor table", err)
	}
	p.debug("... Temporary instructor table created successfully")

	p.fileToTempTable(ctx, "instructors", instrFile, 8, instrSize, procInstrArray)
}

func (p *Parser) openDumpFile(ctx context.Context, path string) (*os.File, int64) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			p.Halt(ctx, fmt.Sprintf("Fatal Error: Dump file %s does not exist!", path))
		}
		p.Halt(ctx, fmt.Sprintf("Fatal Error: Failed to open dump file %s", path), err)
	}

	info, err := file.Stat()
	if err != nil {
		p.Halt(ctx, fmt.Sprintf("Fatal Error: Failed to get file info for dump file %s", path), err)
	}
	return file, info.Size()
}

// Select all the 'quarters' from the meeting pattern to get the start/end
// times for the quarter. Then insert into the quarters table
func (p *Parser) ParseDB(ctx context.Context) {
	quarterQuery := "SELECT strm, start_dt, end_dt FROM meeting GROUP BY strm"
	p.debug("... Creating quarters\n0%", false)

	quarterResult, err := p.db.QueryContext(ctx, quarterQuery)
	if err != nil {
		p.Halt(ctx, "Error: Failed to query quarters", err)
	}
	defer quarterResult.Close()

	// Get total count of quarters for progress tracking
	var totQuart int
	err = p.db.QueryRowContext(ctx, "SELECT COUNT(DISTINCT strm) FROM meeting").Scan(&totQuart)
	if err != nil {
		p.Halt(ctx, "Error: Failed to count quarters", err)
	}

	procQuart := 0   // Count of quarters processed so far for progress tracking
	lastPercent := 0 // last percentage printed in debug mode
	quarters := []string{}
	for quarterResult.Next() {
		// Progress bar
		if p.debugMode {
			percent := int(float64(procQuart) / float64(totQuart) * 100)
			if lastPercent+10 <= percent {
				lastPercent += 10
				fmt.Printf("...%d%%", lastPercent)
			}
		}

		// Get the quarter info
		var strm int
		var startDt, endDt time.Time
		if err := quarterResult.Scan(&strm, &startDt, &endDt); err != nil {
			p.Halt(ctx, "Error: Failed to scan quarter info", err)
		}

		// Convert 4 digits to 5 (2124 -> 20124)
		strmStr := strconv.Itoa(strm)
		if len(strmStr) != 4 {
			p.Halt(ctx, fmt.Sprintf("Error: Invalid strm value %d", strm))
		}
		term := strmStr[0:3] + "0" + strmStr[3:4]

		// Insert the quarter
		// TODO: Change schema from quarters to semesters
		q := `INSERT INTO quarters (quarter, start, end) 
			VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE 
			start=VALUES(start), end=VALUES(end)`
		_, err = p.db.ExecContext(ctx, q, term, startDt, endDt)
		if err != nil {
			fmt.Printf("    *** Error: Failed to insert/update quarter %s\n", term)
			fmt.Printf("        %v\n%v\n", err, q)
			p.failures++
		} else {
			p.quartersProc++
			quarters = append(quarters, term)
		}
	}
	p.debug("...100%")

	// Mark all existing sections as cancelled. If they truly exist, they will be
	// reinstated later in the run
	q := fmt.Sprintf(`UPDATE sections AS s
		JOIN courses AS c ON c.id = course
		SET status = 'X'
		WHERE c.quarter IN (%s)`, getPlaceholders(len(quarters)))
	p.debug("... Marking all sections as canceled")
	_, err = p.db.ExecContext(ctx, q, sliceToAny(quarters)...)
	if err != nil {
		fmt.Printf("*** Error: Failed to mark sections as canceled.\n")
		fmt.Printf("    %v\n", err)
		fmt.Printf("    %s\n", replacePlaceholders(q, quarters))
		p.failures++
		os.Exit(0)
	}

	// Update all the school
	// NOTE: After semesters start, we can no longer use the subject as a lookup
	// for the schools. Subjects are not provided with semester data, and the schools
	// for quarters are well defined. We shall no longer update numeric schools.
	q = `INSERT INTO schools (code)
		SELECT acad_group FROM classes
		WHERE acad_group NOT IN(SELECT code FROM schools 
		WHERE code IS NOT NULL) ON DUPLICATE KEY UPDATE code = code`
	p.debug("... Updating schools")

	if _, err = p.db.ExecContext(ctx, q); err != nil {
		fmt.Printf("*** Error: Failed to update school listings\n")
		fmt.Printf("    %v\n", err)
		fmt.Printf("    %s\n", q)
		p.failures++
	}

	// Select all the departments to add/update
	// NOTE: Again, we're not going to pay attention to numeric schools any longer.
	q = `INSERT INTO departments("code", "school")
		SELECT c."acad_org", s."id"
		FROM classes AS c
			JOIN schools AS s ON s."code" = c."acad_group"
		WHERE strm > 2130
		GROUP BY c."acad_org"
		ON DUPLICATE KEY UPDATE school=VALUES(school)`
	p.debug("... Updating departments")
	_, err = p.db.ExecContext(ctx, q)
	if err != nil {
		fmt.Printf("*** Error: Failed to update department listings\n")
		fmt.Printf("    %v\n", err)
		p.failures++
	}
	// departmentsProc unused in PHP
	// departmentsProc, err := res.RowsAffected()
	// if err != nil {
	// 	fmt.Printf("*** Error: Failed to get department rows affected\n")
	// 	fmt.Printf("    %v\n", err)
	// }

	// Grab each COURSE from the classes table
	q = `SELECT strm, subject, units, acad_org, catalog_nbr, 
		descr, course_descrlong, crse_id, crse_offer_nbr, session_code 
		FROM classes WHERE strm < 20130 GROUP BY crse_id, strm, session_code`
	p.debug("... Updating courses\n0%", false)
	courseResult, err := p.db.QueryContext(ctx, q)
	if err != nil {
		fmt.Printf("*** Error: Failed to get courses\n")
		fmt.Printf("    %v\n", err)
		p.failures++
	}
	procCourses := 0
	var totCourses int
	err = p.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM classes WHERE strm > 20130").Scan(&totCourses)
	if err != nil {
		p.Halt(ctx, "Error: Failed to count courses", err)
	}
	lastPercent = 0
	for courseResult.Next() {
		// Progress bar
		if p.debugMode {
			percent := int(float64(procCourses) / float64(totCourses) * 100)
			if lastPercent+10 <= percent {
				lastPercent += 10
				fmt.Printf("...%d%%", lastPercent)
			}
		}

		var strm, units, crseID, crseOfferNbr int
		var subject, acadOrg, catalogNbr, descr, courseDescrlong, sessionCode string
		if err := courseResult.Scan(&strm, &subject, &units, &acadOrg, &catalogNbr, &descr, &courseDescrlong, &crseID, &crseOfferNbr, &sessionCode); err != nil {
			p.Halt(ctx, "Error: Failed to scan course row", err)
		}

		// Make the term number correct
		strmStr := strconv.Itoa(strm)
		if len(strmStr) != 4 {
			p.Halt(ctx, fmt.Sprintf("Error: Invalid strm value %d", strm))
		}
		strm, _ = strconv.Atoi(strmStr[0:3] + "0" + strmStr[3:4])

		// Insert or update the course
		prm := insertOrUpdateCourseParams{
			quarter:     strm,
			departCode:  acadOrg,
			classCode:   subject,
			course:      catalogNbr,
			credits:     units,
			title:       descr,
			description: courseDescrlong,
		}
		courseId, err := p.insertOrUpdateCourse(ctx, prm)
		if err != nil {
			fmt.Printf("    *** Error: Failed to update %d %s-%s", strm, subject, catalogNbr)
			fmt.Printf("    courseID: %v", courseId)
			fmt.Printf("    %v", err)
			p.failures++
			procCourses++
			continue
		}
		// Process the sections that this course has
		// Step 2) Grab the sections that this course has from temp tables
		sections, err := p.getTempSections(ctx, crseID, crseOfferNbr, strm, sessionCode)
		if err != nil || len(sections) == 0 {
			fmt.Printf("*** Failed to lookup sections for course")
			if err != nil {
				fmt.Printf("    %v", err)
			}
			continue
		}

		// Iterate over the sections of the course
		for _, sect := range sections {
			// Fetch the first instructor for the section
			q = `SELECT CONCAT(first_name,' ',last_name)
				AS i FROM instructors WHERE crse_id=? 
				AND crse_offer_nbr=? AND strm=? AND 
				session_code=? AND class_section=? 
				LIMIT 1`
			var instructor string
			err = p.db.QueryRowContext(ctx, q, crseID, crseOfferNbr, strm, sessionCode, sect.ClassSection).Scan(&instructor)
			if err != nil && err != sql.ErrNoRows {
				p.Halt(ctx, "Failed to find an instructor for course", err)
			}
			if err == sql.ErrNoRows || instructor == "" {
				instructor = "TBA"
			}

			// Process the information about the sesction
			var status string
			if sect.ClassStatus == "X" || sect.SchedulePrint == "N" {
				// Cancelled class (Cancelled, Nonenrollment, Non-printing)
				status = "X"
			} else {
				status = sect.EnrollmentStatus
			}

			var sectionType string
			if sect.InstructionMode == "P" {
				// Regular Mode
				sectionType = "R"
			} else {
				// Just listen to the mode
				sectionType = sect.InstructionMode
			}

			title := sect.Topic

			prm := insertOrUpdateSectionParams{
				courseID:    courseId,
				section:     sect.ClassSection,
				title:       title,
				instructor:  instructor,
				sectionType: sectionType,
				status:      status,
				maxEnroll:   sect.EnrollmentCap,
				curEnroll:   sect.EnrollmentTotal,
			}
			sectId, err := p.insertOrUpdateSection(ctx, prm)
			if err != nil {
				fmt.Printf("*** Failed to insert/update section!")
				fmt.Printf("    %v", err)
				p.failures++
				continue
			}

			/// PROCESS MEETING TIMES ///
			// Remove the meeting times for the section
			q = `DELETE FROM times WHERE section = ?`
			if _, err = p.db.ExecContext(ctx, q, sectId); err != nil {
				fmt.Printf("*** Failed to remove section times\n")
				fmt.Printf("    %v\n", err)
				p.failures++
				continue
			}

			// Select all the meeting times of the section
			q = `SELECT bldg, room_nbr, meeting_time_start
				meeting_time_end, mon, tues, wed, thurs, fri
				sat, sun FROM meeting WHERE crse_id=? AND 
				crse_offer_nbr=? AND strm=? AND session_code=? 
				AND class_section=?`
			timeResult, err := p.db.QueryContext(ctx, q, crseID, crseOfferNbr, strm, sessionCode, sect.ClassSection)
			if err != nil {
				fmt.Printf("*** Failed to query for meeting times\n")
				fmt.Printf("    %v\n", err)
				p.failures++
				continue
			}

			for timeResult.Next() {
				var bldg, roomNbr, meetingTimeStart, meetingTimeEnd string
				var mon, tues, wed, thurs, fri, sat, sun string
				if err := timeResult.Scan(&bldg, &roomNbr, &meetingTimeStart, &meetingTimeEnd, &mon, &tues, &wed, &thurs, &fri, &sat, &sun); err != nil {
					p.Halt(ctx, "Error: Failed to scan meeting time", err)
				}
				origBldg := bldg

				// Parse start and end times into minutes since midnight
				timeRe := regexp.MustCompile(`(\d\d):(\d\d):\d\d`)

				startMatch := timeRe.FindStringSubmatch(meetingTimeStart)
				startHour, _ := strconv.Atoi(startMatch[1])
				startMin, _ := strconv.Atoi(startMatch[2])
				startTime := startHour*60 + startMin

				endMatch := timeRe.FindStringSubmatch(meetingTimeEnd)
				endHour, _ := strconv.Atoi(endMatch[1])
				endMin, _ := strconv.Atoi(endMatch[2])
				endTime := endHour*60 + endMin

				// Special buildings
				switch bldg {
				case "UNKNOWN", "TBD":
					bldg = "TBA"
					roomNbr = "TBA"
				case "OFFC":
					bldg = "OFF"
					roomNbr = "SITE"
				case "ONLINE":
					bldg = "ON"
					roomNbr = "LINE"
				}

				// Lop off a leading 0 (if < 100)
				if n, err := strconv.Atoi(bldg); err == nil && len(bldg) >= 3 && n < 100 {
					bldg = bldg[len(bldg)-2:]
				}

				// Building 7/Institute Hall Situations
				if matched, _ := regexp.MatchString(`[0-9]{3}[A-Za-z]`, bldg); matched {
					bldg = bldg[len(bldg)-3:]
				}

				// Iterate over the days and execute a query
				days := []string{sun, mon, tues, wed, thurs, fri, sat}
				for i, dayTruth := range days {
					if dayTruth == "Y" {
						// TODO: Fix schema to allow `room` to be larger than varchar(4)
						q = `INSERT INTO times (section, day, start, end, building, room) VALUES (?, ?, ?, ?, ?, ?)`
						if _, err = p.db.ExecContext(ctx, q, sectId, i, startTime, endTime, bldg, roomNbr); err != nil {
							fmt.Printf("*** Failed to insert meeting time")
							fmt.Printf("    %v", err)
							fmt.Printf("    %s=>%s", origBldg, bldg)
							p.failures++
						}
					}
				}
			}
		}
		procCourses++
	}
	p.debug("...100%")

	// I guess we're done!
	// Cleanup time
	p.cleanup(ctx)

	// Insert processing statistics
	q = `INSERT INTO scrapelog (timeStarted, timeEnded, quartersAdded, coursesAdded, coursesUpdated, sectionsAdded, sectionsUpdated, failures) VALUES(?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = p.db.ExecContext(ctx, q, p.timeStarted, time.Now().Unix(), p.quartersProc, p.coursesAdded, p.coursesUpdated, p.sectAdded, p.sectUpdated, p.failures)
	if err != nil {
		fmt.Printf("*** Failed to update scrape log")
		fmt.Printf("    %v", err)
	}
}
