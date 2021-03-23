package mysql

import (
    "bufio"
    "context"
    "fmt"
    "github.com/gin-gonic/gin"
    "mysql-agent/common/env"
    "mysql-agent/common/http/client"
    "mysql-agent/common/logger"
    "mysql-agent/common/resultcode"
    "mysql-agent/common/service"
    "mysql-agent/controller/domain"
    "os"
    "strconv"
    "strings"
    "sync"
    "time"
)

const MY_CNF_TEMPLATE = "# For advice on how to change settings please see\n" +
    "# http://dev.mysql.com/doc/refman/5.7/en/server-configuration-defaults.html\n" +
    "\n" +
    "[mysqld]\n" +
    "#\n" +
    "# Remove leading # and set to the amount of RAM for the most important data\n" +
    "# cache in MySQL. Start at 70% of total RAM for dedicated server, else 10%.\n" +
    "# innodb_buffer_pool_size = 128M\n" +
    "#\n" +
    "# Remove leading # to turn on a very important data integrity option: logging\n" +
    "# changes to the binary log between backups.\n" +
    "# log_bin\n" +
    "#\n" +
    "# Remove leading # to set options mainly useful for reporting servers.\n" +
    "# The server defaults are faster for transactions and fast SELECTs.\n" +
    "# Adjust sizes as needed, experiment to find the optimal values.\n" +
    "# join_buffer_size = 128M\n" +
    "# sort_buffer_size = 2M\n" +
    "# read_rnd_buffer_size = 2M\n" +
    "\n" +
    "# Disabling symbolic-links is recommended to prevent assorted security risks\n" +
    "symbolic-links=0\n" +
    "log-error=/var/log/mysqld.log\n" +
    "socket=mysql.sock\n" +
    "\n" +
    "gtid-mode=on\n" +
    "enforce-gtid-consistency=on\n" +
    "log_bin=mysql-bin\n" +
    "\n" +
    "character_set_server=utf8\n" +
    "init_connect='SET NAMES utf8'\n" +
    "\n" +
    "session_track_gtids=OWN_GTID\n" +
    "session_track_state_change=TRUE\n" +
    "\n" +
    "#performance-schema-instrument='transaction=ON'\n" +
    "#performance-schema-consumer-events-transactions-current=ON\n" +
    "#performance-schema-consumer-events-transactions-history=ON\n"

const INSTANCE_TEMPLATE = "[mysqld@rep_port]\n" +
    "datadir=rep_datadir\n" +
    "port=rep_port\n" +
    "server-id=rep_serverid\n"

var conf_lock sync.Mutex

func Install(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    /*req := &domain.MysqlInstallReq{}
      if err := c.ShouldBindJSON(req); err != nil {
          logger.Error("convert remoteInstall req to json failed, err: %s", err.Error())
          return domain.BaseResponse{
              Code:    resultcode.RequestIllegal,
              Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err.Error()),
          }, nil
      }*/

    if isExist, err := service.CheckFileIsExist(*env.ConfPath); isExist {
        if err != nil {
            error := fmt.Sprintf("check conf file [%s] failed, err: %s", *env.ConfPath, err.Error())
            logger.Error(error)
            return domain.BaseResponse{
                Code:    resultcode.SystemInternalException,
                Message: error,
            }, nil
        }
        error := fmt.Sprintf("conf file [%s] is exist, mysql is installed", *env.ConfPath)
        logger.Error(error)
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], error),
        }, nil
    }

    conf_lock.Lock()
    defer conf_lock.Unlock()
    logger.Info("begin to install mysql")
    commands := []string{
        "wget https://dev.mysql.com/get/mysql57-community-release-el7-8.noarch.rpm",
        "rpm -ivh mysql57-community-release-el7-8.noarch.rpm || true",
        "yum install mysql-server -y",
        "systemctl disable mysqld",
        fmt.Sprintf("cp %s %s.bak || true", *env.ConfPath, *env.ConfPath),
        fmt.Sprintf("echo \"%s\" > %s", MY_CNF_TEMPLATE, *env.ConfPath),
    }
    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    logger.Info("end to install mysql")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func UnInstall(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    logger.Info("begin to unInstall mysql")
    file, err := os.Open(*env.ConfPath)
    if err != nil {
        logger.Error("list instance failed, open conf file fail, err: %s", err.Error())
        return domain.ListInstanceResponse{
            BaseResponse: domain.BaseResponse{
                Code:    resultcode.SystemInternalException,
                Message: err.Error(),
            },
        }, nil
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    instanceName := ""
    for scanner.Scan() {
        lineStr := scanner.Text()
        lineStr = strings.TrimSpace(lineStr)
        if strings.HasPrefix(lineStr, "#") {
            continue
        }
        if strings.HasPrefix(lineStr, "[mysqld@") {
            instanceName = lineStr
            instanceName = strings.ReplaceAll(instanceName, "[", "")
            instanceName = strings.ReplaceAll(instanceName, "]", "")
            service.ExecuteMultiCmd([]string{fmt.Sprintf("systemctl stop %s || true", instanceName), fmt.Sprintf("systemctl disable %s || true", instanceName)})
            continue
        }
        if strings.HasPrefix(lineStr, "[mysqld") || instanceName == "" {
            continue
        }
        if confs := strings.Split(lineStr, "="); len(confs) == 2 {
            confs[0] = strings.TrimSpace(confs[0])
            confs[1] = strings.TrimSpace(confs[1])
            if strings.EqualFold(confs[0], "datadir") {
                service.ExecuteSingleCmd("rm -rf " + confs[1])
            }
        }
    }

    commands := []string{
        "yum remove mysql-community* -y",
        "rpm -e mysql57-community-release-el7-8.noarch || true",
        "rm -f " + *env.ConfPath,
    }
    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    logger.Info("end to unInstall mysql")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func AddInstance(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    req := &domain.AddInstanceReq{}
    if err := c.ShouldBindJSON(req); err != nil {
        logger.Error("convert addInstance req to json failed, err: %s", err.Error())
        return domain.AddInstanceRsp{
            BaseResponse: domain.BaseResponse{
                Code:    resultcode.RequestIllegal,
                Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err.Error()),
            },
        }, nil
    }
    port := req.Port
    if err := service.CheckPortValid(port); err != "" {
        return domain.AddInstanceRsp{
            BaseResponse: domain.BaseResponse{
                Code:    resultcode.RequestIllegal,
                Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err),
            },
        }, nil
    }

    portStr := strconv.Itoa(port)
    instanceConfig := strings.ReplaceAll(INSTANCE_TEMPLATE, "rep_port", portStr)
    instanceConfig = strings.ReplaceAll(instanceConfig, "rep_datadir", req.DataDir)
    instanceConfig = strings.ReplaceAll(instanceConfig, "rep_serverid", strconv.Itoa(req.ServerId))

    conf_lock.Lock()
    logger.Info("begin to add instance")
    commands := []string{
        "mkdir -p " + req.DataDir,
        fmt.Sprintf("echo \"%s\" >> %s", instanceConfig, *env.ConfPath),
        "systemctl start mysqld@" + portStr,
        "systemctl enable mysqld@" + portStr,
    }

    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        conf_lock.Unlock()
        return domain.AddInstanceRsp{
            BaseResponse: domain.BaseResponse{
                Code:    resultcode.SystemInternalException,
                Message: error,
            },
        }, nil
    }

    condi := "A temporary password is generated for root@localhost: "
    output, error := service.ExecuteSingleCmd(fmt.Sprintf("cat %s/mysqld.log | grep \"%s\" | tail -n 1", req.DataDir, condi))
    if error != "" {
        conf_lock.Unlock()
        return domain.AddInstanceRsp{
            BaseResponse: domain.BaseResponse{
                Code:    domain.SUCCESS_CODE,
                Message: "add instance success, but get root password failed, err: " + error,
            },
        }, nil
    }
    output = strings.ReplaceAll(output, "\n", "")
    if output == "" {
        conf_lock.Unlock()
        return domain.AddInstanceRsp{
            BaseResponse: domain.BaseResponse{
                Code:    domain.SUCCESS_CODE,
                Message: "add instance success, but get root password is null",
            },
        }, nil
    }

    condi = "root@localhost: "
    tempPwd := output[strings.LastIndex(output, condi)+len(condi):]
    if req.Pwd != "" {
        if ok, error := service.ModifyInstancePwd(port, tempPwd, req.Pwd); !ok {
            conf_lock.Unlock()
            logger.Error("modify Instance Pwd failed, begin to delete instance, err: %s", error)
            client.Delete(fmt.Sprintf("https://%s:%d/v1/mysqlagent/mysql/instance/delete?port=%d", *env.AgentIp, *env.AgentPort, port), nil, nil)
            return domain.BaseResponse{
                Code:    resultcode.SystemInternalException,
                Message: error,
            }, nil
        }
        tempPwd = req.Pwd
    }

    conf_lock.Unlock()
    logger.Info("end to add instance")
    return domain.AddInstanceRsp{
        BaseResponse: domain.BaseResponse{
            Code:    domain.SUCCESS_CODE,
            Message: domain.SUCCESS_MESSAGE,
        },
        Data: domain.InstanceInfo{Password: tempPwd},
    }, nil
}

func DeleteInstance(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    portStr := c.Query("port")
    if portStr == "" {
        error := "port is null"
        logger.Error("delete instance failed, err: %s", error)
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], error),
        }, nil
    }

    conf_lock.Lock()
    defer conf_lock.Unlock()
    logger.Info("begin to delete instance")
    beginLine, endLine, error := service.QueryInstanceConfigRange(portStr)
    if error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    if beginLine == 0 && endLine == 0 {
        return domain.BaseResponse{
            Code:    domain.SUCCESS_CODE,
            Message: "no suit for it, no need to delete",
        }, nil
    }

    output, error := service.ExecuteSingleCmd(fmt.Sprintf("sed -n '%d,%dp' %s | grep datadir", beginLine, endLine, *env.ConfPath))
    if error != "" {
        logger.Error("delete instance failed, failed to query instance datadir, error: %s", error)
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }
    commands := []string{
        "systemctl stop mysqld@" + portStr,
        "systemctl disable mysqld@" + portStr,
        fmt.Sprintf("sed -i '%d, %dd' %s", beginLine, endLine, *env.ConfPath),
    }
    if confs := strings.Split(output, "="); len(confs) == 2 {
        dataDir := strings.TrimSpace(confs[1])
        commands = append(commands, "rm -rf "+dataDir)
    }

    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    logger.Info("end to delete instance")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func ListInstance(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    logger.Info("begin to list instance")
    //file, err := os.Open("D://my.cnf")
    file, err := os.Open(*env.ConfPath)
    if err != nil {
        logger.Error("list instance failed, open conf file fail, err: %s", err.Error())
        return domain.ListInstanceResponse{
            BaseResponse: domain.BaseResponse{
                Code:    resultcode.SystemInternalException,
                Message: err.Error(),
            },
        }, nil
    }
    defer file.Close()
    datas := []domain.InstanceConfig{}
    var instanceConfig domain.InstanceConfig
    scanner := bufio.NewScanner(file)
    instanceName := ""
    for scanner.Scan() {
        lineStr := scanner.Text()
        lineStr = strings.TrimSpace(lineStr)
        if strings.HasPrefix(lineStr, "#") {
            continue
        }
        if strings.HasPrefix(lineStr, "[mysqld@") {
            datas = paddingInstanceStatus(datas, instanceConfig)
            instanceName = lineStr
            instanceName = strings.ReplaceAll(instanceName, "[", "")
            instanceName = strings.ReplaceAll(instanceName, "]", "")
            instanceConfig = domain.InstanceConfig{
                Name:       instanceName,
                Properties: make(map[string]string),
            }
            continue
        }
        if strings.HasPrefix(lineStr, "[mysqld") || instanceName == "" {
            continue
        }
        if confs := strings.Split(lineStr, "="); len(confs) == 2 {
            confs[0] = strings.TrimSpace(confs[0])
            confs[1] = strings.TrimSpace(confs[1])
            instanceConfig.Properties[confs[0]] = confs[1]
            if strings.EqualFold(confs[0], "port") {
                instanceConfig.Port, _ = strconv.Atoi(confs[1])
            }
        }
    }
    datas = paddingInstanceStatus(datas, instanceConfig)

    logger.Info("end to list instance")
    return domain.ListInstanceResponse{
        BaseResponse: domain.BaseResponse{
            Code:    domain.SUCCESS_CODE,
            Message: domain.SUCCESS_MESSAGE,
        },
        Datas: datas,
    }, nil
}

func ModifyInstance(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    req := &domain.ModifyInstanceReq{}
    if err := c.ShouldBindJSON(req); err != nil {
        logger.Error("convert modifyInstance req to json failed, err: %s", err.Error())
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err.Error()),
        }, nil
    }

    portStr := strconv.Itoa(req.Port)
    portStrOfProps, ok := req.Properties["port"]
    if !ok || portStr != portStrOfProps {
        error := "port of properties not exist or not equals port of req"
        logger.Error("modifyInstance failed, err: %s", error)
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], error),
        }, nil
    }

    conf_lock.Lock()
    defer conf_lock.Unlock()
    logger.Info("begin to modify instance")
    beginLine, endLine, error := service.QueryInstanceConfigRange(portStr)
    if error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    if beginLine == 0 && endLine == 0 {
        error := "port config is not exist, please check"
        logger.Error("modifyInstance failed, err: %s", error)
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], error),
        }, nil
    }

    var instanceConfigs []string
    instanceConfigs = append(instanceConfigs, fmt.Sprintf("[mysqld@%s]", portStr))
    for key, value := range req.Properties {
        instanceConfigs = append(instanceConfigs, fmt.Sprintf("%s=%s", key, value))
    }
    commands := []string{
        fmt.Sprintf("sed -i '%d, %dd' %s", beginLine, endLine, *env.ConfPath),
        fmt.Sprintf("echo \"%s\n\" >> %s", strings.Join(instanceConfigs, "\n"), *env.ConfPath),
        "systemctl restart mysqld@" + portStr,
    }

    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    logger.Info("end to modify instance")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func ModifyInstancePwd(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    req := &domain.ModifyInstancePwdReq{}
    if err := c.ShouldBindJSON(req); err != nil {
        logger.Error("convert modifyInstancePwd req to json failed, err: %s", err.Error())
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err.Error()),
        }, nil
    }

    if !service.CheckInstanceIsRunning(req.Port) {
        error := fmt.Sprintf("instance %d is not running", req.Port)
        logger.Error("modifyInstancePwd failed, err: %s", error)
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], error),
        }, nil
    }

    if ok, error := service.ModifyInstancePwd(req.Port, req.OldPwd, req.NewPwd); !ok {
        logger.Error("modifyInstancePwd failed, err: %s", error)
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }

    logger.Info("end to modify instance pwd")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func StartInstance(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    req := &domain.OperateInstanceReq{}
    if err := c.ShouldBindJSON(req); err != nil {
        logger.Error("convert startInstance req to json failed, err: %s", err.Error())
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err.Error()),
        }, nil
    }

    if isExist := service.CheckInstanceIsExist(req.Port); !isExist {
        error := "start instance failed, instance not exist"
        logger.Error(error)
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], error),
        }, nil
    }

    logger.Info("begin to start instance")
    portStr := strconv.Itoa(req.Port)
    commands := []string{
        fmt.Sprintf("systemctl restart mysqld@%s || true", portStr),
        fmt.Sprintf("systemctl enable mysqld@%s || true", portStr),
    }

    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }
    logger.Info("end to start instance")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func StopInstance(ctx context.Context, c *gin.Context) (interface{}, error) {
    ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*5))
    defer cancel()

    req := &domain.OperateInstanceReq{}
    if err := c.ShouldBindJSON(req); err != nil {
        logger.Error("convert stopInstance req to json failed, err: %s", err.Error())
        return domain.BaseResponse{
            Code:    resultcode.RequestIllegal,
            Message: fmt.Sprintf(resultcode.ResultMessage[resultcode.RequestIllegal], err.Error()),
        }, nil
    }

    logger.Info("begin to stop instance")
    portStr := strconv.Itoa(req.Port)
    commands := []string{
        fmt.Sprintf("systemctl stop mysqld@%s || true", portStr),
        fmt.Sprintf("systemctl disable mysqld@%s || true", portStr),
    }

    if _, error := service.ExecuteMultiCmd(commands); error != "" {
        return domain.BaseResponse{
            Code:    resultcode.SystemInternalException,
            Message: error,
        }, nil
    }
    logger.Info("end to stop instance")
    return domain.BaseResponse{
        Code:    domain.SUCCESS_CODE,
        Message: domain.SUCCESS_MESSAGE,
    }, nil
}

func paddingInstanceStatus(datas []domain.InstanceConfig, instanceConfig domain.InstanceConfig) []domain.InstanceConfig {
    if instanceConfig.Name != "" {
        output, error := service.ExecuteSingleCmd(fmt.Sprintf("systemctl status %s | grep Active", instanceConfig.Name))
        if error != "" {
            instanceConfig.Status = error
        } else {
            beginIndex := strings.LastIndex(output, "(")
            endIndex := strings.LastIndex(output, ")")
            if beginIndex != -1 && endIndex != -1 {
                instanceConfig.Status = output[beginIndex+1 : endIndex]
            }
        }
        return append(datas, instanceConfig)
    }
    return datas
}
