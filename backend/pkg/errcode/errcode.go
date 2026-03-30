package errcode

import "fmt"

// Common error codes.
const (
	ErrSuccess      = 0
	ErrInternal     = 500
	ErrBadRequest   = 400
	ErrUnauthorized = 401
	ErrForbidden    = 403
	ErrNotFound     = 404
	ErrConflict     = 409
	ErrTooManyReqs  = 429
	ErrFileTooLarge = 413

	// Business error codes (1000+)
	ErrInvalidInput     = 1001
	ErrEmailExists      = 1002
	ErrInvalidCreds     = 1003
	ErrTokenExpired     = 1004
	ErrTokenInvalid     = 1005
	ErrFileTypeNotAllow = 1006
)

// AppError represents a business error with code and message.
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// New creates a new AppError.
func New(httpStatus, code int, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Predefined errors.
var (
	ErrInternalServer      = New(500, ErrInternal, "internal server error")
	ErrBadRequestMsg       = New(400, ErrBadRequest, "bad request")
	ErrUnauthorizedMsg     = New(401, ErrUnauthorized, "unauthorized")
	ErrForbiddenMsg        = New(403, ErrForbidden, "forbidden")
	ErrNotFoundMsg         = New(404, ErrNotFound, "resource not found")
	ErrEmailAlreadyExists  = New(409, ErrEmailExists, "email already registered")
	ErrInvalidCredentials  = New(401, ErrInvalidCreds, "invalid credentials")
	ErrTokenExpiredMsg     = New(401, ErrTokenExpired, "token expired")
	ErrTokenInvalidMsg     = New(401, ErrTokenInvalid, "invalid token")
	ErrRefreshTokenExpired = New(401, ErrTokenExpired, "refresh token expired")
	ErrFileTypeNotAllowed  = New(400, ErrFileTypeNotAllow, "file type not allowed")
	ErrFileTooLargeMsg     = New(413, ErrFileTooLarge, "file too large")
	ErrTooManyRequests     = New(429, ErrTooManyReqs, "too many requests")
)

// FromHTTPStatus maps an HTTP status code to the default application error code.
func FromHTTPStatus(statusCode int) int {
	switch statusCode {
	case 400:
		return ErrBadRequest
	case 401:
		return ErrUnauthorized
	case 403:
		return ErrForbidden
	case 404:
		return ErrNotFound
	case 409:
		return ErrConflict
	case 413:
		return ErrFileTooLarge
	case 429:
		return ErrTooManyReqs
	default:
		if statusCode >= 500 {
			return ErrInternal
		}
		if statusCode >= 400 {
			return statusCode
		}
		return ErrInternal
	}
}
