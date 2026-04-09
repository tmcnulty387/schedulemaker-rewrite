package api

// TimeSlot represents a single meeting time for a course or non-course item.
// Days are numeric: 0=Sunday, 1=Monday, ..., 6=Saturday
// Times are in minutes from midnight (e.g., 480 = 8:00 AM, 780 = 1:00 PM)
type TimeSlot struct {
	Day   int            `json:"day"`
	Start int            `json:"start"`
	End   int            `json:"end"`
	Bldg  map[string]string `json:"bldg"` // Building style -> building name (e.g., "code": "HAY")
	Room  string         `json:"room"`
}

// Course represents a course section with all its meeting times.
type Course struct {
	CourseNum       string     `json:"courseNum"`       // e.g., "CS-1234-001"
	CourseParentNum string     `json:"courseParentNum"` // e.g., "CS-1234" (base course without section)
	Title           string     `json:"title"`
	Instructor      string     `json:"instructor"`
	Times           []TimeSlot `json:"times"`
	Type            string     `json:"type"`            // e.g., "R" (Lecture), "L" (Lab)
	MaxEnroll       int        `json:"maxenroll"`
	CurrentEnroll   int        `json:"curenroll"`
	Credits         string     `json:"credits"`
	Online          bool       `json:"online"`
}

// NonCourse represents a fixed schedule item (appointment, work, etc.)
type NonCourse struct {
	Title string     `json:"title"`
	Times []TimeSlot `json:"times"`
}

// Schedule represents a complete saved schedule with metadata.
type Schedule struct {
	Courses   []Course   `json:"courses"`
	StartTime int        `json:"startTime"`
	EndTime   int        `json:"endTime"`
	StartDay  int        `json:"startDay"`
	EndDay    int        `json:"endDay"`
	BldgStyle string     `json:"bldgStyle"` // "code" or "number"
	Term      string     `json:"term"`      // e.g., "20241"
	Image     bool       `json:"image"`
}

// GenerateRequest represents the input for schedule generation.
type GenerateRequest struct {
	CourseCount    int     `json:"courseCount"`
	Courses1Opt    []int   `json:"courses1Opt"`
	Courses2Opt    []int   `json:"courses2Opt"`
	// ... additional course option slices as needed
	NonCourseCount int     `json:"nonCourseCount"`
	NoCourseCount  int     `json:"noCourseCount"`
	Verbose        bool    `json:"verbose"`
}

// GenerateResponse represents the output from schedule generation.
type GenerateResponse struct {
	Schedules []Schedule `json:"schedules"`
	Errors    []Error    `json:"errors"`
}

// Error represents an error message from the generation process.
type Error struct {
	Type string `json:"error"` // e.g., "conflict"
	Msg  string `json:"msg"`
}
