package api

import (
	"strings"
	"testing"
	"time"
)

// newScheduleHarness can be registered by the Go implementation to run this
// parity suite against real code. Tests fail until this harness is registered.
var newScheduleHarness func(t *testing.T) scheduleHarness

type scheduleHarness interface {
	ICalFormatTime(minutes int) string
	HashTime(slot TimeSlot, bldgStyle string) string
	GenerateICal(schedule Schedule, courses []Course, termStart time.Time, termEnd time.Time, now time.Time, httpRoot string) (string, error)
}

func requireScheduleHarness(t *testing.T) scheduleHarness {
	t.Helper()
	if newScheduleHarness == nil {
		t.Fatalf("schedule parity harness is not registered: register newScheduleHarness in implementation tests")
	}
	return newScheduleHarness(t)
}

func TestSchedule_ICalFormatTime_ParityCases(t *testing.T) {
	h := requireScheduleHarness(t)

	tests := []struct {
		minutes int
		want    string
	}{
		{0, "000000"},
		{60, "010000"},
		{750, "123000"},
		{1439, "235900"},
		{1440, "000000"}, // wraps midnight like PHP modulo behavior
	}

	for _, tc := range tests {
		got := h.ICalFormatTime(tc.minutes)
		if got != tc.want {
			t.Fatalf("ICalFormatTime(%d) = %q, want %q", tc.minutes, got, tc.want)
		}
	}
}

func TestSchedule_HashTime_UsesBldgStyle(t *testing.T) {
	h := requireScheduleHarness(t)

	slot := TimeSlot{
		Day:   2,
		Start: 540,
		End:   600,
		Bldg: map[string]string{
			"code":   "GOL",
			"number": "070",
		},
		Room: "1410",
	}

	if got := h.HashTime(slot, "code"); got != "540-600-GOL-1410" {
		t.Fatalf("HashTime(code) = %q, want %q", got, "540-600-GOL-1410")
	}
	if got := h.HashTime(slot, "number"); got != "540-600-070-1410" {
		t.Fatalf("HashTime(number) = %q, want %q", got, "540-600-070-1410")
	}
}

func TestSchedule_GenerateICal_CoreShape(t *testing.T) {
	h := requireScheduleHarness(t)

	s := Schedule{BldgStyle: "code", Term: "20241"}
	courses := []Course{
		{
			CourseNum: "CSCI-141-01",
			Title:     "Computer Science I",
			Times: []TimeSlot{
				{Day: 1, Start: 540, End: 590, Bldg: map[string]string{"code": "GOL"}, Room: "1400"},
				{Day: 3, Start: 540, End: 590, Bldg: map[string]string{"code": "GOL"}, Room: "1400"},
			},
		},
		{CourseNum: "non", Title: "Gym", Times: []TimeSlot{{Day: 2, Start: 600, End: 650}}},
	}

	termStart := time.Date(2026, time.January, 12, 0, 0, 0, 0, time.UTC)
	termEnd := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)

	ical, err := h.GenerateICal(s, courses, termStart, termEnd, now, "example.edu")
	if err != nil {
		t.Fatalf("GenerateICal() error = %v", err)
	}

	required := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"METHOD:PUBLISH",
		"BEGIN:VEVENT",
		"SUMMARY:Computer Science I (CSCI-141-01)",
		"LOCATION:GOL-1400",
		"RRULE:FREQ=WEEKLY;INTERVAL=1;WKST=SU;BYDAY=MO,WE;UNTIL=20260501",
		"SUMMARY:Gym",
		"END:VCALENDAR",
	}
	for _, token := range required {
		if !strings.Contains(ical, token) {
			t.Fatalf("GenerateICal() output missing token %q", token)
		}
	}
}
