package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"mysql-agent/common/logger"
	"net/http"
	"strings"
)

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
}

func Get(url string, params map[string]string, headers map[string]string) (string, error) {
	if len(params) > 0 {
		ParseUrl(url, params)
	}

	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Error("create request failed, url: %s, err: %s", url, err.Error())
		return "", err
	}

	return sendRest(request, headers, err, url)
}

func Post(url string, requestBody interface{}, headers map[string]string) (string, error) {
	requestBytes, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("marshal request failed, url: %s, err: %s", url, err.Error())
		return "", err
	}
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestBytes))
	if err != nil {
		logger.Error("create request failed, url: %s, err: %s", url, err.Error())
		return "", err
	}

	return sendRest(request, headers, err, url)
}

func sendRest(request *http.Request, headers map[string]string, err error, url string) (string, error) {
	request.Header.Set("content-type", "application/json")
	request.Header.Set("accept", "application/json")
	if len(headers) > 0 {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}
	client := &http.Client{Transport: transport}
	response, err := client.Do(request)
	if err != nil {
		logger.Error("send rest to %s failed, error: %s", url, err.Error())
		return "", err
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		logger.Error("send rest to %s failed, status: %d", url, response.StatusCode)
		return "", err
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logger.Error("read response body failed, url: %s, error: %s", url, err.Error())
		return "", err
	}
	return string(body), nil
}

func ParseUrl(url string, params map[string]string) {
	if strings.Contains(url, "?") {
		url += "&"
	} else {
		url += "?"
	}
	var _params []string
	for key, value := range params {
		_params = append(_params, key+"="+value)
	}
	url += strings.Join(_params, "&")
}
