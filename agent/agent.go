package main

import (
	"context"
	"flag"
	"fmt"
	"mysql-agent/common/env"
	"mysql-agent/common/http/server"
	"mysql-agent/common/logger"
	"mysql-agent/crontask"
	"mysql-agent/externalfile"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var printVersion bool
	flag.BoolVar(&printVersion, "version", false, "print program build version")
	flag.Parse()
	if printVersion {
		fmt.Printf("%s\n", Version())
		os.Exit(0)
	}

	logger.StartLogger("mysql-agent.log", "info")
	httpServer := server.NewHttpServer(*env.AgentIp, *env.AgentPort)

	go func() {
		logger.Info("agent begin to listen %s:%d", *env.AgentIp, *env.AgentPort)
		certDirs := "externalfile"
		if err := externalfile.RestoreAssets("./", certDirs); err != nil {
			logger.Error("restore http cert fail, error: %s", err.Error())
			return
		}
		serverCrtPath := strings.Join([]string{certDirs, "cert", "server.crt"}, string(filepath.Separator))
		serverKeyPath := strings.Join([]string{certDirs, "cert", "server.key"}, string(filepath.Separator))
		if err := httpServer.ListenAndServeTLS(serverCrtPath, serverKeyPath); err != nil {
			logger.Error("listen http server fail, error: %s", err.Error())
		}
	}()
	crontask.StartCron()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logger.Info("Shutting down server...")

	crontask.StopCron()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error(err.Error())
	}
	logger.Info("Success shutting server.")
}
