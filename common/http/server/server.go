package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"mysql-agent/common/logger"
	"mysql-agent/controller/domain"
	"net/http"
)

func NewHttpServer(ip string, port int) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf("%s:%d", ip, port),
		Handler: newRouter(),
	}
}

func newRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	for index := range OpenRouters {
		openRouter := OpenRouters[index]
		path := openRouter.Pattern
		method := openRouter.Method

		if method == POST {
			router.POST(path, func(c *gin.Context) {
				process(c, openRouter)
			})
		}
		if method == GET {
			router.GET(path, func(c *gin.Context) {
				process(c, openRouter)
			})
		}
		if method == PUT {
			router.PUT(path, func(c *gin.Context) {
				process(c, openRouter)
			})
		}
		if method == DELETE {
			router.DELETE(path, func(c *gin.Context) {
				process(c, openRouter)
			})
		}
		if method == PATCH {
			router.PATCH(path, func(c *gin.Context) {
				process(c, openRouter)
			})
		}
	}
	return router
}

func process(c *gin.Context, route Router) {
	handlerFunc := route.Func
	ctx := metadata.NewOutgoingContext(context.Background(), nil)

	buf := new(bytes.Buffer)
	buf.ReadFrom(c.Request.Body)
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(buf.Bytes()))
	logger.Info("receive request, method: %s, url: %s, reqBody: %s", route.Method, c.Request.URL.String(), buf.String())
	res, err := handlerFunc(ctx, c)
	if err != nil {
		logger.Error("handle error, error: %s", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, &domain.BaseResponse{
			Code:    string(http.StatusInternalServerError),
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, res)
}
