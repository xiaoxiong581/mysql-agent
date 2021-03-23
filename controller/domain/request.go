package domain

/*type MysqlInstallReq struct {
    User    string `json:"user"`
    Pwd     string `json:"pwd"`
    Address string `json:"address"`
    //Version string `json:"version"`
}*/

type AddInstanceReq struct {
    Port     int    `json:"port" binding:"required"`
    ServerId int    `json:"serverId" binding:"required"`
    DataDir  string `json:"dataDir" binding:"required"`
    Pwd      string `json:"pwd"`
}

type ModifyInstanceReq struct {
    Port       int               `json:"port" binding:"required"`
    Properties map[string]string `json:"properties"`
}

type ModifyInstancePwdReq struct {
    Port   int    `json:"port" binding:"required"`
    OldPwd string `json:"oldPwd"`
    NewPwd string `json:"newPwd"`
}

type OperateInstanceReq struct {
    Port int `json:"port" binding:"required"`
}
