package responses

// ErrorEnvelope is a generic error wrapper for HTTP responses.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains a standardized error shape for clients.
type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}
