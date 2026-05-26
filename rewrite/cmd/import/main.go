package main

/// processDump.php

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"rewrite/internal/config"
	"rewrite/internal/tools"
)

func main() {
	ctx := context.Background()

	c := config.Load()
	dbConn, err := sql.Open("mysql", c.GetDataSourceName())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	parser := tools.NewParser(ctx, dbConn, os.Args)

	timeStarted := time.Now().Unix()
	quartersProc := 0
	departmentsProc := 0
	coursesAdded := 0
	coursesUpdated := 0
	sectAdded := 0
	sectUpdated := 0
	failures := 0

	// Dump... vars will never be empty b/c they have defaults
	if (!fileExists(c.DumpClasses)) {
		parser.Halt(ctx, "Fatal Error: Class dump file does not exist!")
	}
	if (!fileExists(c.DumpClassAttr)) {
		parser.Halt(ctx, "Fatal Error: Class attribute dump file does not exist!")
	}
	if (!fileExists(c.DumpInstruct)) {
		parser.Halt(ctx, "Fatal Error: Instructor dump file does not exist!")
	}
	if (!fileExists(c.DumpMeeting)) {
		parser.Halt(ctx, "Fatal Error: Class meeting pattern dump file does not exist!")
	}
	if (!fileExists(c.DumpNotes)) {
		parser.Halt(ctx, "Fatal Error: Class notes dump file does not exist!")
	}

	// FILE PARSING ////////////////////////////////////////////////////////////
	// Open handles to the files that were given to us from ITS
	// $classFile = fopen($DUMPCLASSES, 'r');
	// $attrFile = fopen($DUMPCLASSATTR, 'r');
	// $instrFile = fopen($DUMPINSTRUCT, 'r');
	// $meetFile = fopen($DUMPMEETING, 'r');
	// $notesFile = fopen($DUMPNOTES, 'r');

	//  Store how many bytes we have
	// $classSize = filesize($DUMPCLASSES);
	// $attrSize = filesize($DUMPCLASSATTR);
	// $instrSize = filesize($DUMPINSTRUCT);
	// $meetSize = filesize($DUMPMEETING);
	// $notesSize = filesize($DUMPNOTES);

	
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}