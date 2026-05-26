package api

type ErrorResponse struct {
	Error string `json:"error"`
	Msg   string `json:"msg"`
	Arg   string `json:"arg,omitempty"`
}

