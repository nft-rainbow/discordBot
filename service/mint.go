package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

func SendEasyMintRequest(dto models.MintReq) (*models.MintResp, error){
	b, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	fmt.Println("Start to easy mint")
	req, _ := http.NewRequest("POST", viper.GetString("host") + "v1/mints/easy/urls", bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return  nil, err
	}

	var res models.MintResp
	content, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &res)
	if err != nil {
		return nil, err
	}
	if res.Message != "" {
		return nil, errors.New(res.Message)
	}

	defer resp.Body.Close()
	return &res, nil
}

func SendCustomMintRequest(dto models.MintReq) (*models.MintResp, error){
	b, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	fmt.Println("Start to custom mint")
	req, _ := http.NewRequest("POST", viper.GetString("host") + "user/mint/custom", bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return  nil, err
	}

	var res models.MintResp
	content, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &res)
	if err != nil {
		return nil, err
	}
	if res.Message != "" {
		return nil, errors.New(res.Message)
	}

	defer resp.Body.Close()
	return &res, nil
}

