package apierror

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const timestampLayout = "2006-01-02T15:04:05.000Z07:00"

type Detail struct {
	Field       string   `json:"field"`
	Constraints []string `json:"constraints"`
}

type Envelope struct {
	Timestamp string        `json:"timestamp"`
	Path      string        `json:"path"`
	Error     EnvelopeError `json:"error"`
}

type EnvelopeError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Details []Detail `json:"details,omitempty"`
}

type Error struct {
	Status  int
	Code    string
	Message string
	Details []Detail
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	if e.Cause != nil {
		return e.Cause.Error()
	}

	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

func New(status int, code, message string, details []Detail) *Error {
	return &Error{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}

func Validation(details []Detail) *Error {
	return New(http.StatusBadRequest, "VALIDATION_ERROR", "Request validation failed", details)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, "NOT_FOUND", message, nil)
}

func MethodNotAllowed(message string) *Error {
	return New(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", message, nil)
}

func Internal(err error) *Error {
	return &Error{
		Status:  http.StatusInternalServerError,
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "Internal server error",
		Cause:   err,
	}
}

func Normalize(err error) *Error {
	if err == nil {
		return nil
	}

	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}

	return Internal(err)
}

func Write(w http.ResponseWriter, r *http.Request, err error) {
	appErr := Normalize(err)
	if appErr == nil {
		return
	}

	payload := Envelope{
		Timestamp: time.Now().UTC().Format(timestampLayout),
		Path:      r.URL.Path,
		Error: EnvelopeError{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
	}

	var body bytes.Buffer
	if encodeErr := json.NewEncoder(&body).Encode(payload); encodeErr != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.Status)
	_, _ = w.Write(body.Bytes())
}
