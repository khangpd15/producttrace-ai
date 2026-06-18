package response

type ApiResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Error   any    `json:"error,omitempty"`
}

func ResponseError(msg string, err any) ApiResponse {
	return ApiResponse{
		Success: false,
		Message: msg,
		Data:    nil,
		Error:   err,
	}
}

func ResponseSuccess(msg string, data any) ApiResponse {
	return ApiResponse{
		Success: true,
		Message: msg,
		Data:    data,
		Error:   nil,
	}
}
