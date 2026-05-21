package protocol

import "net/http"

type Request struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
}

type Response struct {
	ID     string      `json:"id"`
	Status int         `json:"status"`
	Header http.Header `json:"header"`
	Body   []byte      `json:"body"`
	Error  string      `json:"error,omitempty"`
}
