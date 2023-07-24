package component

import (
	"github.com/gin-gonic/gin"
)

type GokuApiResponse struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

func (gar GokuApiResponse) Render(context *gin.Context, code int, result interface{}, err error) {

	if err != nil {
		gar.Message = err.Error()
		gar.Code = 1
		context.JSON(code, gar)
	} else {
		gar.Code = 0
		gar.Data = result
		gar.Message = "success"
		context.JSON(code, gar)
	}

}
