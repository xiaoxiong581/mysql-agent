package cron

import (
	"github.com/robfig/cron/v3"
	"mysql-agent/common/logger"
	"sync"
)

type Crontab struct {
	mutex sync.Mutex
	cron  *cron.Cron
	ids   map[string]cron.EntryID
}

func NewCrontab() *Crontab {
	return &Crontab{
		cron: cron.New(),
		ids:  make(map[string]cron.EntryID),
	}
}

func (c *Crontab) Start() {
	c.cron.Start()
}

func (c *Crontab) Stop() {
	c.cron.Stop()
}

func (c *Crontab) Add(id string, cronExp string, jobFunc func()) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.ids[id]; ok {
		logger.Warn("cron job is exist, id: %s", id)
		return nil
	}

	entryId, err := c.cron.AddFunc(cronExp, jobFunc)
	if err != nil {
		logger.Error("add cron job error, id: %s, error: %s", id, err.Error())
		return err
	}
	c.ids[id] = entryId
	logger.Info("add cron job %s success", id)
	return nil
}

func (c *Crontab) del(id string) {
	if _, ok := c.ids[id]; !ok {
		logger.Warn("cron job %s is not exist", id)
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entryId, ok := c.ids[id]
	if !ok {
		logger.Warn("cron job %s is not exist", id)
		return
	}

	c.cron.Remove(entryId)
	delete(c.ids, id)
	logger.Info("delete cron job %s success", id)
}
