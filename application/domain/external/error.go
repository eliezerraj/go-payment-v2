package external

import (
	"context"
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
)

const RequestIDHeaderName = "x-request-id"
const SERVER_ERROR = "server_error"
const BUSSINESS_ERROR = "business_error"
const CLIENT_ERROR = "client_error"

// Error struct represents a structured error response. the literal - means ommited.
type Error struct {
	OriginalError error      `json:"-"`
	InnerError    InnerError `json:"error"`
	StatusCode    int        `json:"-"`
	Type          string     `json:"type"`
	Reason        string     `json:"reason"`
}

// InnerError struct represents the inner details of an error
type InnerError struct {
	Id           string           `json:"id"`
	Code         string           `json:"code"`
	Description  string           `json:"description"`
	ErrorDetails ValidationErrors `json:"error_details,omitempty"`
}

//----------------------------------
// ValidationError represents a single validation error
// ---------------------------------
type ValidationError struct {
	Attribute string   `json:"attribute"`
	Messages  []string `json:"messages"`
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var errMsgs []string
	for _, err := range ve {
		errMsgs = append(errMsgs, strings.Join(err.Messages, ", "))
	}
	return strings.Join(errMsgs, "; ")
}

//---------------------------------
// ValidationError represents a single validation error
// ---------------------------------
func NewResponseError(	ctx context.Context,
						statusCode int, 
						err error, 
						code, 
						description, 
						reason, 
						typeErr string) *Error {

	requestID, ok := ctx.Value(RequestIDHeaderName).(string)
	if !ok {
		requestID = uuid.NewString()
	}

	var errResp *Error

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		logger.ErrorOutCtx("context timeout", zap.Any("stack_trace", err))
		
		return &Error{
			OriginalError: err,
			Reason:        reason,
			InnerError: InnerError{
				Id:          requestID,
				Code:        code,
				Description: fiber.ErrRequestTimeout.Message,
			},
			StatusCode: fiber.StatusRequestTimeout,
			Type:       typeErr,
		}
	} else {
		errResp = &Error{
			OriginalError: err,
			InnerError: InnerError{
				Id:          requestID,
				Code:        code,
				Description: description,
			},
			StatusCode: statusCode,
			Type:       typeErr,
			Reason:     reason,
		}
	}

	return errResp
}