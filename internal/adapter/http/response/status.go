package response

type StatusResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewStatus(status string) StatusResponse {
	return StatusResponse{Status: status}
}

func NewError(message string) ErrorResponse {
	return ErrorResponse{Error: message}
}
