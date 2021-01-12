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
}

type ModifyInstanceReq struct {
    Port       int               `json:"port" binding:"required"`
    Properties map[string]string `json:"properties"`
}
