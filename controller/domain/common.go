package domain

type InstanceConfig struct {
    Name string `json:"name"`
    Port int `json:"port"`
    Status string `json:"status"`
    Properties map[string]string `json:"properties"`
}
