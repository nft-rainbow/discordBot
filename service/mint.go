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
	"strconv"
	"time"
)

func SendEasyMintRequest(token string, dto models.EasyMintMetaDto) (*models.MintResp, error){
	b, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	fmt.Println("Start to easy mint")
	req, _ := http.NewRequest("POST", viper.GetString("host") + "v1/mints/easy/urls", bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return  nil, err
	}

	var tmp models.MintTask
	content, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.ErrMessage != "" {
		return nil, errors.New(tmp.ErrMessage)
	}
	id, err := getTokenId(tmp.ID, token)
	if err != nil {
		return nil, err
	}

	res := &models.MintResp{
		UserAddress: dto.MintToAddress,
		Contract: viper.GetString("easyMint.contract"),
		NFTAddress: viper.GetString("easyMint.mintRespPrefix") + viper.GetString("easyMint.contract") + "/" + id,
		TokenID: id,
		Time: tmp.BaseModel.CreatedAt.String(),
	}

	defer resp.Body.Close()
	return res, nil
}

func SendCustomMintRequest(token string, dto models.CustomMintDto) (*models.MintResp, error){
	b, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}

	fmt.Println("Start to custom mint")
	req, _ := http.NewRequest("POST", viper.GetString("host") + "v1/mints/", bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return  nil, err
	}

	var tmp models.MintTask
	content, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return nil, err
	}
	if tmp.ErrMessage != "" {
		return nil, errors.New(tmp.ErrMessage)
	}

	id, err := getTokenId(tmp.ID, token)
	if err != nil {
		return nil, err
	}

	res := &models.MintResp{
		UserAddress: dto.MintToAddress,
		NFTAddress: viper.GetString("customMint.mintRespPrefix") +  dto.ContractAddress + "/" + id,
		Contract: dto.ContractAddress,
		TokenID: id,
		Time: tmp.BaseModel.CreatedAt.String(),
	}

	defer resp.Body.Close()
	return res, nil
}

func CreateMetadata(token, fileUrl, name, description string) (string, error) {
	metadata := models.Metadata{
		Name: name,
		Description: description,
		Image: fileUrl,
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	fmt.Println("Start to create metadata")
	req, _ := http.NewRequest("POST", viper.GetString("host") + "v1/metadata/", bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
		return  "", err
	}

	var tmp models.CreateMetadataResponse
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return "", err
	}
	if tmp.Message != "" {
		return "", errors.New(tmp.Message)
	}

	return tmp.MetadataURI, nil
}

func getTokenId(id uint, token string) (string, error) {
	t := models.MintTask{}
	fmt.Println("Start to get token id")
	for t.TokenId == "" && t.Status != 1{
		req, err := http.NewRequest("GET", viper.GetString("host") + "v1/mints/" + strconv.Itoa(int(id)),nil)
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
		if t.Error != "" {
			return "", errors.New(t.Error)
		}
		time.Sleep(10 * time.Second)
	}
	return t.TokenId, nil
}
