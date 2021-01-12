package crontask

import (
	"flag"
	"fmt"
	"mysql-agent/common/http/client"
	"mysql-agent/common/logger"
)

const REGISTER_TO_SERVER_TASK = "registerToServerTask"
var serverUrl = flag.String("server", "127.0.0.1:30033", "server address")

func RegisterToServerTask() {
	heartBeatUrl := fmt.Sprintf("%s%s%s", "https://", *serverUrl, "/v1/mysqlagent/health/healthcheck")
	res, err := client.Get(heartBeatUrl, nil, nil)
	if err != nil {
		logger.Warn("register to server fail, error: %s", err.Error())
		return
	}
	logger.Info("register to server success, response: %s", res)
}
