package tools

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"slices"
)

var coursesUpdated = 0
var coursesAdded = 0

type Parser struct {
	dbConn    *sql.DB
	debugMode bool
	quietMode bool
}

func NewParser(ctx context.Context, dbConn *sql.DB, arguments []string) *Parser {
	debugMode := slices.Contains(arguments, "-d")
	quietMode := slices.Contains(arguments, "-q")
	p := &Parser{
		dbConn:    dbConn,
		debugMode: debugMode,
		quietMode: quietMode,
	}

	if slices.Contains(arguments, "-c") {
		p.cleanup(ctx)
		os.Exit(0)
	}
	return p
}

func (p *Parser) cleanup(ctx context.Context) {
	p.debug("... Cleaning up temporary tables")

	if _, err := p.dbConn.ExecContext(ctx, "DROP TABLE classes"); err != nil {
		fmt.Fprintln(os.Stderr, "*** Failed to drop table classes (ignored)")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
	}
	if _, err := p.dbConn.ExecContext(ctx, "DROP TABLE meeting"); err != nil {
		fmt.Fprintln(os.Stderr, "*** Failed to drop table meeting (ignored)")
		fmt.Fprintf(os.Stderr, "    %v\n", err)
	}
	if _, err := p.dbConn.ExecContext(ctx, "DROP TABLE instructors"); err != nil {
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
	rows, err := p.dbConn.QueryContext(ctx, query, prm.quarter, prm.classCode, prm.course, prm.credits, prm.title, prm.description)
	if err != nil {
		rows, err = p.dbConn.QueryContext(ctx, query, prm.quarter, prm.departCode, prm.course, prm.credits, prm.title, prm.description)
		if err != nil {
			return 0, err
		}
	}

	// First row: action (updated/inserted)
	if (!rows.Next()) {
		return 0, fmt.Errorf("unexpected result from InsertOrUpdateCourse: no rows returned")
	}
	var action string
	if err := rows.Scan(&action); err != nil {
		return 0, fmt.Errorf("failed to scan result from InsertOrUpdateCourse: %w", err)
	}
	switch action {
	case "updated":
		coursesUpdated++
	case "inserted":
		coursesAdded++
	default:
		return 0, fmt.Errorf("unexpected action from InsertOrUpdateCourse: %s", action)
	}

	// Second row: course id
	if (!rows.Next()) {
		return 0, fmt.Errorf("unexpected result from InsertOrUpdateCourse: only one row returned")
	}
	var courseID int
	if err := rows.Scan(&courseID); err != nil {
		return 0, fmt.Errorf("failed to scan course id from InsertOrUpdateCourse: %w", err)
	}

	return courseID, nil
}
