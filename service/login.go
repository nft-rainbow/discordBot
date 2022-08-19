package service

import (
	"bytes"
	"encoding/json"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

func Login() (string, error) {
	data := make(map[string]string)
	data["app_id"] = viper.GetString("app.appId")
	data["app_secret"] = viper.GetString("app.appSecret")
	b, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", viper.GetString("app.loginUrl"), bytes.NewBuffer(b))
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return  "", err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	t := make(map[string]string)
	err = json.Unmarshal(content, &t)
	if err != nil {
		return "", err
	}

	return t["token"], nil
}
