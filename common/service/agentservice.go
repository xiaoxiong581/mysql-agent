package service

import (
    "flag"
    "fmt"
    "mysql-agent/common/logger"
    "os"
    "os/exec"
    "strconv"
    "strings"
)

var ConfPath = flag.String("ConfPath", "/etc/my.cnf", "mysql config file path")

func CheckFileIsExist(filePath string) (bool, error) {
    _, err := os.Stat(filePath)
    if err != nil {
        return false, err
    }
    if err == nil || os.IsExist(err) {
        return true, nil
    }
    return false, nil
}

func CheckPortValid(port int) string {
    command := fmt.Sprintf("cat %s | grep \"port=%d\" | wc -l", *ConfPath, port)
    output, err := ExecuteSingleCmd(command)
    if err != "" {
        logger.Error("check port valid failed")
        return err
    }
    output = strings.ReplaceAll(output, "\n", "")
    count, err1 := strconv.Atoi(output)
    if err1 != nil {
        error := fmt.Sprintf("convert check port result failed, err: %s", err1.Error())
        logger.Error(error)
        return error
    }
    if count > 0 {
        error := fmt.Sprintf("port %d is in use, please update it", port)
        logger.Error(error)
        return error
    }

    return ""
}

func QueryInstanceConfigRange(portStr string) (int, int, string) {
    if portStr == "" {
        logger.Info("port or ConfPath is null")
        return 0, 0, ""
    }

    cmd := fmt.Sprintf("cat %s | grep -n '\\[mysqld'", *ConfPath)
    output, error := ExecuteSingleCmd(cmd)
    outputList := strings.Split(output, "\n")
    if len(outputList) == 0 {
        logger.Info("query mysqld is null")
        return 0, 0, ""
    }

    begin := 0
    beginIndex := 0
    end := 0
    condi := fmt.Sprintf("[mysqld@%s]", portStr)
    isHas := false
    for index, result := range outputList {
        if strings.Contains(result, condi) {
            isHas = true
            beginStr := strings.Split(result, ":")[0]
            beginInt, err := strconv.Atoi(beginStr)
            if err != nil {
                error := fmt.Sprintf("convert config begin line to int failed, output: %s, err: %s", beginStr, err.Error())
                logger.Error(error)
                return 0, 0, error
            }
            beginIndex = index
            begin = beginInt
            break
        }
    }
    if !isHas {
        logger.Info("query mysqld not suit for %s", portStr)
        return 0, 0, ""
    }
    if beginIndex >= len(outputList) || outputList[beginIndex+1] == "" {
        cmd = fmt.Sprintf("cat %s | wc -l", *ConfPath)
        output, error = ExecuteSingleCmd(cmd)
        if error != "" {
            return 0, 0, error
        }
        output = strings.ReplaceAll(output, "\n", "")
        totalLine, err := strconv.Atoi(output)
        if err != nil {
            error := fmt.Sprintf("convert config total lines result to int failed, output: %s, err: %s", output, err.Error())
            logger.Error(error)
            return 0, 0, error
        }
        return begin, totalLine, ""
    }

    nextStr := outputList[beginIndex+1]
    endStr := strings.Split(nextStr, ":")[0]
    end, err := strconv.Atoi(endStr)
    if err != nil {
        error := fmt.Sprintf("convert config end line to int failed, output: %s, err: %s", endStr, err.Error())
        logger.Error(error)
        return 0, 0, error
    }

    return begin, end - 1, ""
}

func ExecuteMultiCmd(commands []string) (string, string) {
    output := ""
    for _, cmd := range commands {
        output, err := ExecuteSingleCmd(cmd)
        if err != "" {
            return output, err
        }
    }
    return output, ""
}

func ExecuteSingleCmd(command string) (string, string) {
    logger.Info("begin to execute command [%s]", command)
    cmdExec := exec.Command("/bin/sh", "-c", command);
    result, err := cmdExec.Output()
    outputStr := string(result)
    logger.Info("execute command [%s] result: \n[%s]", command, outputStr)
    if err != nil {
        error := fmt.Sprintf("execute [%s] command failed, err: %s", command, err.Error())
        logger.Error(error)
        return outputStr, error
    }
    return outputStr, ""
}
