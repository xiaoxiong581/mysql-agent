package env

import "flag"

var AgentIp = flag.String("ip", "0.0.0.0", "agent ip")
var AgentPort = flag.Int("port", 30033, "agent port")
var ConfPath = flag.String("confPath", "/etc/my.cnf", "mysql config file path")