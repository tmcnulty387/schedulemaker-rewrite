package utils

import "fmt"

// Day codes mapping (numeric -> string)
var dayCodes = map[int]string{
	0: "Sun",
	1: "Mon",
	2: "Tue",
	3: "Wed",
	4: "Thur",
	5: "Fri",
	6: "Sat",
}

// String to numeric day mapping
var stringToDay = map[string]int{
	"Sun":  0,
	"Mon":  1,
	"Tue":  2,
	"Wed":  3,
	"Thu":  4, // Accept both Thu and Thur
	"Thur": 4,
	"Fri":  5,
	"Sat":  6,
}

// TranslateDay converts between numeric and string day representations.
// If input is numeric (0-6), returns 3-letter string (e.g., "Mon").
// If input is string (e.g., "Mon"), returns numeric (0-6).
// Defaults to Sunday (0) if input is unrecognized.
func TranslateDay(day interface{}) interface{} {
	switch v := day.(type) {
	case int:
		if code, ok := dayCodes[v]; ok {
			return code
		}
		return dayCodes[0] // Default to Sunday
	case string:
		if num, ok := stringToDay[v]; ok {
			return num
		}
		return 0 // Default to Sunday
	default:
		return 0 // Default to Sunday
	}
}

// TranslateTime converts minutes from midnight to human-readable time format.
// time: minutes from midnight (0-1439)
// twelve: if true, use 12-hour format with AM/PM; if false, use 24-hour format
// Returns formatted time string (e.g., "8:00 am" or "08:00")
func TranslateTime(minutes int, twelve bool) string {
	var period string
	hour := minutes / 60
	min := minutes % 60

	if twelve {
		// Determine AM/PM
		if minutes >= 720 { // 12:00 PM or later
			period = " pm"
		} else {
			period = " am"
		}

		// Convert to 12-hour format
		if minutes >= 780 { // 1:00 PM or later
			hour = (minutes - 720) / 60
		} else if minutes < 60 { // Before 1:00 AM
			hour = 12
		} else if hour == 0 {
			hour = 12
		}
	}

	if twelve {
		return fmt.Sprintf("%d:%02d%s", hour, min, period)
	}
	return fmt.Sprintf("%02d:%02d", hour, min)
}

// TranslateTimeDump converts time from RITS dump format (HHMM) to minutes from midnight.
// Input: integer like 1430 (2:30 PM)
// Output: integer minutes (870)
func TranslateTimeDump(time int) int {
	hour := time / 100
	min := time % 100
	return (hour * 60) + min
}
