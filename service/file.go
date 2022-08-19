package service

import (
	"bytes"
	"encoding/json"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

func UploadFile(token, path string) (string, error){
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, _ := bodyWriter.CreateFormFile("file", file.Name())

	io.Copy(fileWriter, file)

	contentType := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	req, _ := http.NewRequest("POST", viper.GetString("host") + "v1/files", bodyBuffer)

	req.Header.Add("Authorization", "Bearer " + token)
	req.Header.Add("Content-Type", contentType)
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var t models.UploadFilesResponse

	err = json.Unmarshal(body, &t)
	if err != nil {
		return "", err
	}

	return t.FileUrl, nil
}
