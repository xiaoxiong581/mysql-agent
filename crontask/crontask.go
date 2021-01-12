package crontask

import (
	"mysql-agent/common/cron"
)

var crontab *cron.Crontab

func StartCron() {
	crontab = cron.NewCrontab()
	/*if err := crontab.Add(REGISTER_TO_SERVER_TASK, "@every 1m", RegisterToServerTask); err != nil {
		logger.Error("add registerToServerTask fail, error: %s", err.Error())
	}*/

	crontab.Start()
}

func StopCron() {
	if crontab != nil {
		crontab.Stop()
	}
}