package api

import (
	"reflect"
	"testing"
)

// newGenerateHarness can be registered by the Go implementation to run this
// parity suite against real code. Tests fail until this harness is registered.
var newGenerateHarness func(t *testing.T) generateHarness

type generateHarness interface {
	OverlapBase(itemTimes []TimeSlot, courseTimes []TimeSlot) bool
	GenerateSchedules(courses [][]Course, nonCourses []NonCourse, noCourses []NonCourse) ([][]Course, error)
	GetCleanCourseNum(course Course) string
	PruneSpecialCourses(schedules [][]Course, courseGroups map[string]map[string]struct{}) [][]Course
}

func requireGenerateHarness(t *testing.T) generateHarness {
	t.Helper()
	if newGenerateHarness == nil {
		t.Fatalf("generate parity harness is not registered: register newGenerateHarness in implementation tests")
	}
	return newGenerateHarness(t)
}

func TestGenerate_OverlapBase_ParityCases(t *testing.T) {
	h := requireGenerateHarness(t)

	tests := []struct {
		name       string
		itemTimes  []TimeSlot
		courseTime []TimeSlot
		want       bool
	}{
		{name: "empty item times", itemTimes: nil, courseTime: []TimeSlot{{Day: 1, Start: 600, End: 650}}, want: false},
		{name: "empty course times", itemTimes: []TimeSlot{{Day: 1, Start: 600, End: 650}}, courseTime: nil, want: false},
		{name: "same day start inside item", itemTimes: []TimeSlot{{Day: 2, Start: 600, End: 700}}, courseTime: []TimeSlot{{Day: 2, Start: 650, End: 750}}, want: true},
		{name: "same day end inside item", itemTimes: []TimeSlot{{Day: 2, Start: 600, End: 700}}, courseTime: []TimeSlot{{Day: 2, Start: 500, End: 650}}, want: true},
		{name: "course engulfs item", itemTimes: []TimeSlot{{Day: 2, Start: 600, End: 700}}, courseTime: []TimeSlot{{Day: 2, Start: 550, End: 725}}, want: true},
		{name: "adjacent end start is non-overlap", itemTimes: []TimeSlot{{Day: 3, Start: 600, End: 700}}, courseTime: []TimeSlot{{Day: 3, Start: 700, End: 800}}, want: false},
		{name: "different day is non-overlap", itemTimes: []TimeSlot{{Day: 4, Start: 600, End: 700}}, courseTime: []TimeSlot{{Day: 5, Start: 600, End: 700}}, want: false},
		{name: "multi-slot any overlap wins", itemTimes: []TimeSlot{{Day: 1, Start: 480, End: 540}, {Day: 3, Start: 600, End: 660}}, courseTime: []TimeSlot{{Day: 3, Start: 620, End: 700}}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := h.OverlapBase(tc.itemTimes, tc.courseTime)
			if got != tc.want {
				t.Fatalf("OverlapBase() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestGenerate_GetCleanCourseNum_ParityCases(t *testing.T) {
	h := requireGenerateHarness(t)

	tests := []struct {
		in   string
		want string
	}{
		{in: "CSCI-141-01", want: "CSCI-141-01"},
		{in: "CSCI-141-A", want: "CSCI-141"},
		{in: "MATH-190-A1", want: "MATH-190"},
		{in: "BIOL-101-Z99", want: "BIOL-101"},
		{in: "SOIS-001", want: "SOIS-001"},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := h.GetCleanCourseNum(Course{CourseNum: tc.in})
			if got != tc.want {
				t.Fatalf("GetCleanCourseNum(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestGenerate_GenerateSchedules_CartesianNoConflicts(t *testing.T) {
	h := requireGenerateHarness(t)

	courses := [][]Course{
		{
			{CourseNum: "A-1", Times: []TimeSlot{{Day: 1, Start: 480, End: 540}}},
			{CourseNum: "A-2", Times: []TimeSlot{{Day: 2, Start: 480, End: 540}}},
		},
		{
			{CourseNum: "B-1", Times: []TimeSlot{{Day: 3, Start: 540, End: 600}}},
			{CourseNum: "B-2", Times: []TimeSlot{{Day: 4, Start: 540, End: 600}}},
		},
	}

	got, err := h.GenerateSchedules(courses, nil, nil)
	if err != nil {
		t.Fatalf("expected no error for conflict-free cartesian generation, got %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("expected 4 schedules (cartesian product), got %d", len(got))
	}
}

func TestGenerate_GenerateSchedules_WithConflictsAndConstraints(t *testing.T) {
	h := requireGenerateHarness(t)

	courses := [][]Course{
		{{CourseNum: "A-1", Times: []TimeSlot{{Day: 1, Start: 480, End: 540}}}},
		{
			{CourseNum: "B-1", Times: []TimeSlot{{Day: 1, Start: 500, End: 560}}}, // conflicts with A-1
			{CourseNum: "B-2", Times: []TimeSlot{{Day: 2, Start: 500, End: 560}}},
		},
	}

	nonCourses := []NonCourse{{Title: "Work", Times: []TimeSlot{{Day: 2, Start: 490, End: 510}}}}  // conflicts with B-2
	noCourses := []NonCourse{{Title: "No 8am", Times: []TimeSlot{{Day: 1, Start: 470, End: 550}}}} // conflicts with A-1

	got, err := h.GenerateSchedules(courses, nonCourses, noCourses)
	if len(got) != 0 {
		t.Fatalf("expected zero valid schedules, got %d", len(got))
	}
	if err == nil {
		t.Fatalf("expected non-nil error when all branches are invalid")
	}
}

func TestGenerate_PruneSpecialCourses_ParityCases(t *testing.T) {
	h := requireGenerateHarness(t)

	schedules := [][]Course{
		{
			{CourseNum: "CHEM-101", CourseParentNum: "CHEM-101"},
			{CourseNum: "CHEM-101L", CourseParentNum: "CHEM-101"},
		},
		{
			{CourseNum: "CHEM-101", CourseParentNum: "CHEM-101"},
		},
		{
			{CourseNum: "MATH-181", CourseParentNum: "MATH-181"},
		},
	}

	courseGroups := map[string]map[string]struct{}{
		"CHEM-101": {"CHEM-101L": {}},
	}

	got := h.PruneSpecialCourses(schedules, courseGroups)
	want := [][]Course{schedules[0], schedules[2]}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("PruneSpecialCourses() mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
