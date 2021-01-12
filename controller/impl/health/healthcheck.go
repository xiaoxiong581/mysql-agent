package health

import (
	"context"
	"github.com/gin-gonic/gin"
	"mysql-agent/controller/domain"
)

func Healthcheck(ctx context.Context, c *gin.Context) (interface{}, error) {
	return domain.BaseResponse{
		Code:    domain.SUCCESS_CODE,
		Message: domain.SUCCESS_MESSAGE,
	}, nil
}
