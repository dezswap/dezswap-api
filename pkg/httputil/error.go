package httputil

import "github.com/gin-gonic/gin"

// NewError
func NewError(ctx *gin.Context, status int, err error) {
	er := HTTPError{
		Code:    status,
		Message: err.Error(),
	}
	ctx.JSON(status, er)
}

// HTTPError example
type HTTPError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"bad request"`
}

// HTTPError example
type BadRequestError struct {
	Code    int    `json:"code" example:"400"`
	Message string `json:"message" example:"bad request"`
}

// HTTPError example
type NotFoundError struct {
	Code    int    `json:"code" example:"404"`
	Message string `json:"message" example:"not found"`
}

// HTTPError example
type InternalServerError struct {
	Code    int    `json:"code" example:"500"`
	Message string `json:"message" example:"internal server error"`
}
