package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetObject(query string, result interface{}) error {
	response, err := http.Get(query)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	return nil
}
