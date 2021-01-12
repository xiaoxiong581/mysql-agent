package main

import (
	"fmt"
)

var (
	// version 项目版本信息
	version string
	// goVersion Go版本信息
	goVersion string
	// buildTime 构建时间
	buildTime string
)

func Version() string {
	return fmt.Sprintf(
		"Version: %s\n"+
			"Go Version: %s\n"+
			"Build Time: %s", version, goVersion, buildTime)
}
