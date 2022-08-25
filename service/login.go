package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

func Login() (string, error) {
	data := make(map[string]string)
	data["app_id"] = viper.GetString("app.appId")
	data["app_secret"] = viper.GetString("app.appSecret")
	b, _ := json.Marshal(data)
	fmt.Println("Start to login")
	req, err := http.NewRequest("POST", viper.GetString("host") + "v1/login", bytes.NewBuffer(b))
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
	t := make(map[string]interface{})
	err = json.Unmarshal(content, &t)
	if err != nil {
		return "", err
	}
	if t["code"] != nil {
		return "", errors.New(t["message"].(string))
	}

	return t["token"].(string), nil
}
