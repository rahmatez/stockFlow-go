package response

import (
	"encoding/json"
	"net/http"
)

type Meta struct {
	Page  int   `json:"page,omitempty"`
	Limit int   `json:"limit,omitempty"`
	Total int64 `json:"total,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	write(w, status, Envelope{Success: true, Data: data})
}

func JSONWithMeta(w http.ResponseWriter, status int, data interface{}, meta Meta) {
	write(w, status, Envelope{Success: true, Data: data, Meta: &meta})
}

func Error(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Envelope{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	})
}

func write(w http.ResponseWriter, status int, body Envelope) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
