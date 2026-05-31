package api

import (
	"net/http"

	"rewrite/internal/config"
)

type handler struct {
	cfg *config.Config
}

func NewHandler(cfg *config.Config) http.Handler {
	h := &handler{cfg: cfg}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/status", h.handleStatus)
	mux.HandleFunc("GET /api/terms", h.handleTerms)

	mux.HandleFunc("POST /api/entity/getSchools", h.handleGetSchools)
	mux.HandleFunc("POST /api/entity/getSchoolsForTerm", h.handleGetSchoolsForTerm)
	mux.HandleFunc("POST /api/entity/getDepartments", h.handleGetDepartments)
	mux.HandleFunc("POST /api/entity/getCourses", h.handleGetCourses)
	mux.HandleFunc("POST /api/entity/getSections", h.handleGetSections)
	mux.HandleFunc("POST /api/entity/courseForSection", h.handleCourseForSection)

	mux.HandleFunc("POST /api/search/find", h.handleSearchFind)

	mux.HandleFunc("POST /api/generate/getCourseOpts", h.handleGetCourseOpts)
	mux.HandleFunc("POST /api/generate/getMatchingSchedules", h.handleGetMatchingSchedules)

	mux.HandleFunc("POST /api/schedule/new", h.handleScheduleNew)
	mux.HandleFunc("GET /api/schedule/{hex}/ical", h.handleScheduleICal)
	mux.HandleFunc("GET /api/schedule/{hex}/old", h.handleScheduleOld)
	mux.HandleFunc("GET /api/schedule/{hex}", h.handleScheduleGet)

	mux.HandleFunc("GET /img/schedules/", h.handleScheduleImage)

	mux.HandleFunc("POST /api/rmp", h.handleRMP)

	return mux
}

func (h *handler) handleStatus(w http.ResponseWriter, r *http.Request)               {
	
}
func (h *handler) handleTerms(w http.ResponseWriter, r *http.Request)                {}
func (h *handler) handleGetSchools(w http.ResponseWriter, r *http.Request)           {}
func (h *handler) handleGetSchoolsForTerm(w http.ResponseWriter, r *http.Request)    {}
func (h *handler) handleGetDepartments(w http.ResponseWriter, r *http.Request)       {}
func (h *handler) handleGetCourses(w http.ResponseWriter, r *http.Request)           {}
func (h *handler) handleGetSections(w http.ResponseWriter, r *http.Request)          {}
func (h *handler) handleCourseForSection(w http.ResponseWriter, r *http.Request)     {}
func (h *handler) handleSearchFind(w http.ResponseWriter, r *http.Request)           {}
func (h *handler) handleGetCourseOpts(w http.ResponseWriter, r *http.Request)        {}
func (h *handler) handleGetMatchingSchedules(w http.ResponseWriter, r *http.Request) {}
func (h *handler) handleScheduleNew(w http.ResponseWriter, r *http.Request)          {}
func (h *handler) handleScheduleGet(w http.ResponseWriter, r *http.Request)          {}
func (h *handler) handleScheduleICal(w http.ResponseWriter, r *http.Request)         {}
func (h *handler) handleScheduleOld(w http.ResponseWriter, r *http.Request)          {}
func (h *handler) handleScheduleImage(w http.ResponseWriter, r *http.Request)        {}
func (h *handler) handleRMP(w http.ResponseWriter, r *http.Request)                  {}
