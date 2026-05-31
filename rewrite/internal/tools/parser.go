package tools

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Parser struct {
	db              *sql.DB
	debugMode       bool
	quietMode       bool
	timeStarted     int64
	quartersProc    int
	departmentsProc int
	coursesAdded    int
	coursesUpdated  int
	sectAdded       int
	sectUpdated     int
	failures        int
}

func NewParser(ctx context.Context, dbConn *sql.DB, arguments []string) *Parser {
	debugMode := slices.Contains(arguments, "-d")
	quietMode := slices.Contains(arguments, "-q")
	p := &Parser{
		db:              dbConn,
		debugMode:       debugMode,
		quietMode:       quietMode,
		timeStarted:     time.Now().Unix(),
		quartersProc:    0,
		departmentsProc: 0,
		coursesAdded:    0,
		coursesUpdated:  0,
		sectAdded:       0,
		sectUpdated:     0,
		failures:        0,
	}

	if slices.Contains(arguments, "-c") {
		p.cleanup(ctx)
		os.Exit(0)
	}
	return p
}

func (p *Parser) cleanup(ctx context.Context) {
	p.debug("... Cleaning up temporary tables")

	if _, err := p.db.ExecContext(ctx, "DROP TABLE classes"); err != nil {
		fmt.Fprintln(os.Stderr, "*** Failed to drop table classes (ignored)")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
	}
	if _, err := p.db.ExecContext(ctx, "DROP TABLE meeting"); err != nil {
		fmt.Fprintln(os.Stderr, "*** Failed to drop table meeting (ignored)")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
	}
	if _, err := p.db.ExecContext(ctx, "DROP TABLE instructors"); err != nil {
		fmt.Fprintln(os.Stderr, "*** Failed to drop table instructor (ignored)")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
	}
}

func (p *Parser) debug(str string, nl ...bool) {
	if !p.debugMode {
		return
	}
	newline := true
	if len(nl) > 0 {
		newline = nl[0]
	}
	if newline {
		fmt.Println(str)
	} else {
		fmt.Print(str)
	}
}

// cleanupExtraResults not needed due to how mysql
// works in Go

func (p *Parser) Halt(ctx context.Context, msgs ...any) {
	for _, msg := range msgs {
		fmt.Printf("*** %v\n", msg)
	}
	p.cleanup(ctx)
	os.Exit(0)
}

type insertOrUpdateCourseParams struct {
	quarter     int
	departCode  string
	classCode   string
	course      string
	credits     int
	title       string
	description string
}

func (p *Parser) insertOrUpdateCourse(ctx context.Context, prm insertOrUpdateCourseParams) (int, error) {
	query := "CALL InsertOrUpdateCourse(?, 0000, ?, ?, ?, ?, ?)"
	rows, err := p.db.QueryContext(ctx, query, prm.quarter, prm.classCode, prm.course, prm.credits, prm.title, prm.description)
	if err != nil {
		rows, err = p.db.QueryContext(ctx, query, prm.quarter, prm.departCode, prm.course, prm.credits, prm.title, prm.description)
		if err != nil {
			return 0, err
		}
	}
	defer rows.Close()

	// First row: action (updated/inserted)
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, fmt.Errorf("failed to advance to action row in InsertOrUpdateCourse: %w", err)
		}
		return 0, fmt.Errorf("unexpected result from InsertOrUpdateCourse: no rows returned")
	}
	var action string
	if err := rows.Scan(&action); err != nil {
		return 0, fmt.Errorf("failed to scan result from InsertOrUpdateCourse: %w", err)
	}
	switch action {
	case "updated":
		p.coursesUpdated++
	case "inserted":
		p.coursesAdded++
	default:
		return 0, fmt.Errorf("unexpected action from InsertOrUpdateCourse: %s", action)
	}

	// Second row: course id
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, fmt.Errorf("failed to advance to id row in InsertOrUpdateCourse: %w", err)
		}
		return 0, fmt.Errorf("unexpected result from InsertOrUpdateCourse: only one row returned")
	}
	var courseID int
	if err := rows.Scan(&courseID); err != nil {
		return 0, fmt.Errorf("failed to scan course id from InsertOrUpdateCourse: %w", err)
	}

	return courseID, nil
}

type insertOrUpdateSectionParams struct {
	courseID    int
	section     string
	title       string
	instructor  string
	sectionType string
	status      string
	maxEnroll   int
	curEnroll   int
}

func (p *Parser) insertOrUpdateSection(ctx context.Context, prm insertOrUpdateSectionParams) (int, error) {
	query := "CALL InsertOrUpdateSection(?, ?, ?, ?, ?, ?, ?, ?)"
	rows, err := p.db.QueryContext(ctx, query, prm.courseID, prm.section, prm.title, prm.instructor, prm.sectionType, prm.status, prm.maxEnroll, prm.curEnroll)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	// First row: action (updated/inserted)
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, fmt.Errorf("failed to advance to action row in InsertOrUpdateSection: %w", err)
		}
		return 0, fmt.Errorf("unexpected result from InsertOrUpdateSection: no rows returned")
	}
	var action string
	if err := rows.Scan(&action); err != nil {
		return 0, fmt.Errorf("failed to scan result from InsertOrUpdateSection: %w", err)
	}
	switch action {
	case "updated":
		p.sectUpdated++
	case "inserted":
		p.sectAdded++
	default:
		return 0, fmt.Errorf("unexpected action from InsertOrUpdateSection: %s", action)
	}

	// Second row: section id
	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return 0, fmt.Errorf("failed to advance to id row in InsertOrUpdateSection: %w", err)
		}
		return 0, fmt.Errorf("unexpected result from InsertOrUpdateSection: only one row returned")
	}
	var sectionID int
	if err := rows.Scan(&sectionID); err != nil {
		return 0, fmt.Errorf("failed to scan section id from InsertOrUpdateSection: %w", err)
	}

	return sectionID, nil
}

type TempSection struct {
	ClassSection     string
	Description      string
	Topic            string
	EnrollmentStatus string
	ClassStatus      string
	ClassType        string
	EnrollmentCap    int
	EnrollmentTotal  int
	InstructionMode  string
	SchedulePrint    string
}

func (p *Parser) getTempSections(ctx context.Context, courseNum, offerNum, term int, sessionCode string) ([]TempSection, error) {
	query := `SELECT class_section,descr,topic,enrl_stat,class_stat,class_type,enrl_cap,enrl_tot,instruction_mode,schedule_print
		FROM classes WHERE crse_id=? AND crse_offer_nbr=? AND strm=? AND session_code=?`

	rows, err := p.db.QueryContext(ctx, query, courseNum, offerNum, term, sessionCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []TempSection
	for rows.Next() {
		var s TempSection
		if err := rows.Scan(&s.ClassSection, &s.Description, &s.Topic, &s.EnrollmentStatus,
			&s.ClassStatus, &s.ClassType, &s.EnrollmentCap, &s.EnrollmentTotal,
			&s.InstructionMode, &s.SchedulePrint); err != nil {
			return nil, err
		}
		sections = append(sections, s)
	}
	return sections, rows.Err()
}

type LineProcessor func([]string) ([]string, bool)

func (p *Parser) fileToTempTable(ctx context.Context, tableName string, file *os.File, fields int, fileSize int64, procFunc LineProcessor) error {
	procChars := 0   // total number of characters read from file so far, only used in debug mode
	lastPercent := 0 // last percentage printed in debug mode
	p.debug(fmt.Sprintf("... Copying %s file to temporary table\n0%%", tableName), false)

	scanner := bufio.NewScanner(file)
	decoder := charmap.ISO8859_1.NewDecoder()

	for scanner.Scan() {
		str, err := readLine(decoder, scanner)
		if err != nil {
			return err
		}

		// Progress bar
		if p.debugMode {
			procChars += len(str) + 1 // +1 for newline character
			percent := int((float64(procChars) / float64(fileSize)) * 100)
			if lastPercent+10 <= percent {
				lastPercent += 10
				fmt.Printf("...%d%%", lastPercent)
			}
		}

		// If we don't have 23 pipes, then we need to read another line
		lineSplit := strings.Split(str, "|")
		for len(lineSplit) < fields+1 {
			if !scanner.Scan() {
				return fmt.Errorf("unexpected end of file while reading multi-line record for %s", tableName)
			}
			nextStr, err := readLine(decoder, scanner)
			if err != nil {
				return err
			}
			str += nextStr
			lineSplit = strings.Split(str, "|")
		}

		// If we don't have $fields+1 fields, shit's borked
		if len(lineSplit) != fields+1 {
			fmt.Printf("*** Malformed line %d, %d\n", fields, len(lineSplit))
			fmt.Printf("%s\n", str)
			continue
		}

		// We only need the first $fields, otherwise imploding will break
		lineSplit = lineSplit[:fields]

		// Call the special attribute processing function
		lineSplit, readLine := procFunc(lineSplit)
		if !readLine {
			// procFunc doesn't want us to read this line, skip it
			continue
		}

		// Build a query
		query := fmt.Sprintf("INSERT INTO %s VALUES(%s)", tableName, getPlaceholders(fields))

		_, err = p.db.ExecContext(ctx, query, sliceToAny(lineSplit)...)
		if err != nil {
			fmt.Printf("*** Failed to insert %s\n", tableName)
			fmt.Printf("    %s\n", err.Error())
		}
	}
	p.debug("...100%")
	return nil
}

func readLine(d *encoding.Decoder, s *bufio.Scanner) (string, error) {
	bytes, _, err := transform.Bytes(d, s.Bytes())
	if err != nil {
		return "", fmt.Errorf("failed to convert line to UTF-8: %w", err)
	}
	str := strings.TrimSpace(string(bytes))
	return str, nil
}

func sliceToAny[T any](s []T) []any {
	result := make([]any, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

// Process the class file
func procClassArray(lineSplit []string) ([]string, bool) {
	// Trim course number and topic
	lineSplit[6] = strings.TrimSpace(lineSplit[6])
	lineSplit[8] = strings.TrimSpace(lineSplit[8])

	// Grab the integer credit count (they give it to us as a decimal)
	re := regexp.MustCompile(`(\d)+\.\d\d`)
	match := re.FindStringSubmatch(lineSplit[11])
	if len(match) > 1 {
		lineSplit[11] = match[1]
	}

	// Make the section number at least 2 digits
	n, _ := strconv.Atoi(lineSplit[4])
	lineSplit[4] = fmt.Sprintf("%02d", n)
	return lineSplit, true
}

func procMeetArray(lineSplit []string) ([]string, bool) {
	timeRe := regexp.MustCompile(`(\d\d):(\d\d) ([A-Z]{2})`)

	start := timeRe.FindStringSubmatch(lineSplit[10])
	end := timeRe.FindStringSubmatch(lineSplit[11])
	if start == nil || end == nil {
		return nil, false
	}

	lineSplit[10] = formatTime(start)
	lineSplit[11] = formatTime(end)

	n, _ := strconv.Atoi(lineSplit[4])
	lineSplit[4] = fmt.Sprintf("%02d", n)

	return lineSplit, true
}

func formatTime(match []string) string {
	hours, _ := strconv.Atoi(match[1])
	minutes := match[2]
	if match[3] == "PM" {
		hours = (hours % 12) + 12
	} else {
		hours = hours % 12
	}
	return fmt.Sprintf("%02d%s00", hours, minutes)
}

func procInstrArray(lineSplit []string) ([]string, bool) {
	n, _ := strconv.Atoi(lineSplit[4])
	lineSplit[4] = fmt.Sprintf("%02d", n)
	return lineSplit, true
}

func getPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat("?,", n-1) + "?"
}

func replacePlaceholders[T any](q string, args []T) string {
    for _, arg := range args {
        q = strings.Replace(q, "?", fmt.Sprintf("%v", arg), 1)
    }
    return q
}
