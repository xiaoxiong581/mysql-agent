package resultcode

const (
	Success                  = "0"
	SystemInternalException  = "102"
	RequestIllegal           = "103"
	RestError                = "104"
)

var ResultMessage = map[string]string{
	Success:                  "success",
	SystemInternalException:  "system internal exception",
	RequestIllegal:           "request illegal, error: %s",
	RestError:                "send rest to third error",
}
