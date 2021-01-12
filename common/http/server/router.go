package server

import (
    "context"
    "github.com/gin-gonic/gin"
    "mysql-agent/controller/impl/health"
    "mysql-agent/controller/impl/mysql"
)

const (
    POST   = "post"
    GET    = "get"
    PUT    = "put"
    DELETE = "delete"
    PATCH  = "patch"
)

type httpHandler func(ctx context.Context, c *gin.Context) (interface{}, error)
type Router struct {
    Method  string
    Pattern string
    Func    httpHandler
}

var OpenRouters = []Router{
    {GET, "/v1/mysqlagent/health/healthcheck", health.Healthcheck},
    {POST, "/v1/mysqlagent/mysql/install", mysql.Install},
    {POST, "/v1/mysqlagent/mysql/uninstall", mysql.UnInstall},
    {POST, "/v1/mysqlagent/mysql/instance/add", mysql.AddInstance},
    {DELETE, "/v1/mysqlagent/mysql/instance/delete", mysql.DeleteInstance},
    {GET, "/v1/mysqlagent/mysql/instance/list", mysql.ListInstance},
    {POST, "/v1/mysqlagent/mysql/instance/modify", mysql.ModifyInstance},
    {POST, "/v1/mysqlagent/mysql/instance/start", mysql.StartInstance},
    {POST, "/v1/mysqlagent/mysql/instance/stop", mysql.StopInstance},
}
