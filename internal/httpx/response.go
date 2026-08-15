// Package httpx berisi helper HTTP: response envelope standar, router
// tipis di atas net/http.ServeMux (Go 1.22+) dengan dukungan chaining
// middleware, dan util pengambilan context request (klaim JWT, dsb).
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

// Envelope adalah bentuk response JSON standar seluruh API.
type Envelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("[httpx] gagal menulis response JSON: %v", err)
	}
}

func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, Envelope{Success: true, Data: data})
}

func OKMeta(w http.ResponseWriter, data interface{}, meta interface{}) {
	JSON(w, http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, Envelope{Success: true, Data: data})
}

func Err(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, Envelope{Success: false, Error: &ErrorInfo{Code: code, Message: message}})
}

func BadRequest(w http.ResponseWriter, message string) {
	Err(w, http.StatusBadRequest, "bad_request", message)
}
func Unauthorized(w http.ResponseWriter, message string) {
	Err(w, http.StatusUnauthorized, "unauthorized", message)
}
func Forbidden(w http.ResponseWriter, message string) {
	Err(w, http.StatusForbidden, "forbidden", message)
}
func NotFound(w http.ResponseWriter, message string) {
	Err(w, http.StatusNotFound, "not_found", message)
}
func Conflict(w http.ResponseWriter, message string) {
	Err(w, http.StatusConflict, "conflict", message)
}
func TooManyRequests(w http.ResponseWriter, message string) {
	Err(w, http.StatusTooManyRequests, "rate_limited", message)
}
func InternalError(w http.ResponseWriter, err error) {
	log.Printf("[httpx] internal error: %v", err)
	Err(w, http.StatusInternalServerError, "internal_error", "Terjadi kesalahan pada server")
}
