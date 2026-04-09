package utils

import (
	"testing"
)

func TestTranslateDay_NumericToString(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"Sunday", 0, "Sun"},
		{"Monday", 1, "Mon"},
		{"Tuesday", 2, "Tue"},
		{"Wednesday", 3, "Wed"},
		{"Thursday", 4, "Thur"},
		{"Friday", 5, "Fri"},
		{"Saturday", 6, "Sat"},
		{"Invalid - defaults to Sunday", 99, "Sun"},
		{"Negative - defaults to Sunday", -1, "Sun"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateDay(tt.input)
			if result != tt.expected {
				t.Errorf("TranslateDay(%d) = %v; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTranslateDay_StringToNumeric(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Sunday", "Sun", 0},
		{"Monday", "Mon", 1},
		{"Tuesday", "Tue", 2},
		{"Wednesday", "Wed", 3},
		{"Thursday (Thu)", "Thu", 4},
		{"Thursday (Thur)", "Thur", 4},
		{"Friday", "Fri", 5},
		{"Saturday", "Sat", 6},
		{"Invalid - defaults to Sunday", "Invalid", 0},
		{"Empty string - defaults to Sunday", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateDay(tt.input)
			if result != tt.expected {
				t.Errorf("TranslateDay(%q) = %v; want %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTranslateTime_12HourFormat(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{"Midnight", 0, "12:00 am"},
		{"12:15 AM", 15, "12:15 am"},
		{"1:00 AM", 60, "1:00 am"},
		{"8:00 AM", 480, "8:00 am"},
		{"11:59 AM", 719, "11:59 am"},
		{"Noon", 720, "12:00 pm"},
		{"12:30 PM", 750, "12:30 pm"},
		{"1:00 PM", 780, "1:00 pm"},
		{"8:00 PM", 1200, "8:00 pm"},
		{"11:59 PM", 1439, "11:59 pm"},
		{"Edge case: 10:30 AM", 630, "10:30 am"},
		{"Edge case: 3:45 PM", 945, "3:45 pm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateTime(tt.minutes, true)
			if result != tt.expected {
				t.Errorf("TranslateTime(%d, true) = %q; want %q", tt.minutes, result, tt.expected)
			}
		})
	}
}

func TestTranslateTime_24HourFormat(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{"Midnight", 0, "00:00"},
		{"12:15 AM", 15, "00:15"},
		{"1:00 AM", 60, "01:00"},
		{"8:00 AM", 480, "08:00"},
		{"Noon", 720, "12:00"},
		{"1:00 PM", 780, "13:00"},
		{"8:00 PM", 1200, "20:00"},
		{"11:59 PM", 1439, "23:59"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateTime(tt.minutes, false)
			if result != tt.expected {
				t.Errorf("TranslateTime(%d, false) = %q; want %q", tt.minutes, result, tt.expected)
			}
		})
	}
}

func TestTranslateTimeDump(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Midnight", 0, 0},
		{"1:00 AM", 100, 60},
		{"8:00 AM", 800, 480},
		{"12:30 PM", 1230, 750},
		{"2:30 PM", 1430, 870},
		{"8:00 PM", 2000, 1200},
		{"11:59 PM", 2359, 1439},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateTimeDump(tt.input)
			if result != tt.expected {
				t.Errorf("TranslateTimeDump(%d) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
