package dto

import "github.com/gin-gonic/gin"

// ErrorResponse стандартный формат ошибки для всех API-ответов.
type ErrorResponse struct {
	RequestID string `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID запроса для трассировки"`
	Code      string `json:"code" example:"VALIDATION_ERROR" description:"Машиночитаемый код ошибки"`
	Message   string `json:"message" example:"Некорректный формат email" description:"Человекочитаемое описание ошибки"`
}

// NewErrorResponse создаёт ErrorResponse с request_id из контекста Gin.
func NewErrorResponse(c *gin.Context, code error, message string) ErrorResponse {
	return ErrorResponse{
		RequestID: c.GetHeader("X-Request-Id"),
		Code:      code.Error(),
		Message:   message,
	}
}
