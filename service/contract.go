package service

import (
	"bytes"
	"encoding/json"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/nft-rainbow/discordBot/utils"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func DeployContract(token, name, symbol, owner string) (string, error){
	contract := models.ContractDeployDto{
		Chain: utils.CONFLUX_TEST,
		Name: name,
		Symbol: symbol,
		OwnerAddress: owner,
		Type: viper.GetString("deployContract.type"),
	}

	b, err := json.Marshal(contract)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", viper.GetString("deployContract.deployUrl"), bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return  "", err
	}

	var tmp models.Contract

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return "", err
	}

	address, err := getContractAddress(tmp.ID, token)
	if err != nil {
		return "", err
	}

	return address, nil
}

func getContractAddress(id uint, token string) (string, error){
	t := models.Contract{}
	for t.Address == "" {
		req, err := http.NewRequest("GET", viper.GetString("deployContract.infoUrl") + strconv.Itoa(int(id)),nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer " + token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		err = json.Unmarshal(content, &t)
		if err != nil {
			return "", err
		}
		time.Sleep(10 * time.Second)
	}
	return t.Address, nil
}
