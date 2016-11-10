package utils

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"gottb.io/goru/errors"
)

func HTTPRequestJSON(method, url string, data interface{}, headers map[string]string) (int, []byte, error) {
	client := &http.Client{}
	body, err := json.Marshal(data)
	if err != nil {
		return 0, nil, errors.Wrap(err)
	}
	bodyReader := bytes.NewBuffer(body)
	request, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return 0, nil, errors.Wrap(err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, nil, errors.Wrap(err)
	}
	defer response.Body.Close()
	responseContent, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, nil, errors.Wrap(err)
	}
	return response.StatusCode, responseContent, nil
}
