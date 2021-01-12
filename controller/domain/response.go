package domain

const (
    SUCCESS_CODE    = "0"
    SUCCESS_MESSAGE = "success"
)

type BaseResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type ListInstanceResponse struct {
    BaseResponse
    Datas []InstanceConfig `json:"datas"`
}
