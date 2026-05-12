package juphttp

import "fmt"

// APIError is a structured Jupiter API error.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Endpoint   string
	Body       string
	Retryable  bool
}

func (e *APIError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Code != "" {
		return fmt.Sprintf("jupiter %s failed: status=%d code=%s message=%s request_id=%s", e.Endpoint, e.StatusCode, e.Code, e.Message, e.RequestID)
	}
	return fmt.Sprintf("jupiter %s failed: status=%d message=%s request_id=%s", e.Endpoint, e.StatusCode, e.Message, e.RequestID)
}

// IsRetryable reports whether err is a retryable APIError.
func IsRetryable(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Retryable
}
