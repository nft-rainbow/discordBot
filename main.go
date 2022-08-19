package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/nft-rainbow/discordBot/utils"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var easyMintRestrain map[string]int = make(map[string]int)
var customRestrain map[string]int = make(map[string]int)

const advertise = "Powered by NFTRainbow"

func initConfig() {
	viper.SetConfigName("config")             // name of config file (without extension)
	viper.SetConfigType("yaml")               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")                  // optionally look for config in the working directory
	err := viper.ReadInConfig()               // Find and read the config file
	if err != nil {                           // Handle errors reading the config file
		log.Fatalln(fmt.Errorf("fatal error config file: %w", err))
	}
}

func init() {
	initConfig()
}

func main() {
	s, _ := discordgo.New("Bot " + viper.GetString("bot.token"))
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// /claim easyMint <userAddress> (fileUrl)
		if strings.Contains(m.Content, "/claim easyMint") {
			userAddress := strings.Split(m.Content, " ")[2]
			_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			if easyMintRestrain[userAddress] > 0 {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("@%s One account can only obtain a NFT", m.Author.Username))
				return
			}

			var fileUrl string
			if len(strings.Split(m.Content, " ")) >= 4 {
				fileUrl = strings.Split(m.Content, " ")[3]
				if _, err = url.ParseRequestURI(fileUrl); err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
					return
				}
			}else {
				fileUrl = viper.GetString("easyMint.fileUrl")
			}

			easyMintRestrain[userAddress] ++
			token, err := login()
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				easyMintRestrain[userAddress] --
				return
			}
			if token == "" {
				_, _ = s.ChannelMessageSend(m.ChannelID, "get access token failed")
				easyMintRestrain[userAddress] --
				return
			}

			resp , err := sendEasyMintRequest(token, models.EasyMintMetaDto{
				Chain: viper.GetString("easyMint.chain"),
				Name: viper.GetString("easyMint.name"),
				Description: viper.GetString("easyMint.description"),
				MintToAddress: userAddress,
				FileUrl: fileUrl,
			})
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				easyMintRestrain[userAddress] --
				return
			}
			_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content:   fmt.Sprintf("<@%s> Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", m.Author.ID, resp.UserAddress, resp.NFTAddress, resp.Advertise),
				Reference: m.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{nil, nil, []string{m.Author.ID},
				},
			})
		}
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// /claim customNFT <userAddress> <contract_address> <name> <description> (file_url)
		if strings.Contains(m.Content, "/claim customNFT") {
			contents := strings.Split(m.Content, " ")
			userAddress := contents[2]
			_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			contractAddress := contents[3]
			_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, contractAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			name, description := contents[4], contents[5]

			if customRestrain[userAddress] > 0 {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("@%s One account can only obtain a NFT", m.Author.ID))
				return
			}

			var fileUrl string
			if len(strings.Split(m.Content, " ")) >= 7 {
				fileUrl = strings.Split(m.Content, " ")[6]
				if _, err = url.ParseRequestURI(fileUrl); err != nil {
					_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
					return
				}
			}else {
				fileUrl = viper.GetString("customMint.fileUrl")
			}

			fmt.Println(fileUrl)

			customRestrain[userAddress] ++
			token, err := login()
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				customRestrain[userAddress] --
				return
			}
			if token == "" {
				_, _ = s.ChannelMessageSend(m.ChannelID, "get access token failed")
				customRestrain[userAddress] --
				return
			}

			metadataUri, err := createMetadata(token, fileUrl, name, description)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				customRestrain[userAddress] --
				return
			}

			resp , err := sendCustomMintRequest(token, models.CustomMintDto{
				Chain: viper.GetString("customMint.chain"),
				MintToAddress: userAddress,
				ContractAddress: contractAddress,
				MetadataUri: metadataUri,
			})
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				customRestrain[userAddress] --
				return
			}
			_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content:   fmt.Sprintf("<@%s> Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", m.Author.ID, resp.UserAddress, resp.NFTAddress, resp.Advertise),
				Reference: m.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{nil, nil, []string{m.Author.ID},
				},
			})
		}
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.Contains(m.Content, "/claim file") {
			downloadUrl := m.Attachments[0].URL
			fmt.Println(downloadUrl)
			if _, err := url.ParseRequestURI(downloadUrl); err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			token, err := login()
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			fileUrl, err := uploadFile(token, downloadUrl, m.Attachments[0].Filename)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content:   fmt.Sprintf("<@%s> Congratulate on uploading files successfully. You can use the url to mint your own NFT: %s ! \n %s", m.Author.ID, fileUrl, advertise),
				Reference: m.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{nil, nil, []string{m.Author.ID},
				},
			})
		}
	})

	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// /claim contract <userAddress> <name> <symbol>
		if strings.Contains(m.Content, "/claim contract") {
			contents := strings.Split(m.Content, " ")
			userAddress, contractName := contents[2], contents[3]
			_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			if len(contents) < 5 {
				_, _ = s.ChannelMessageSend(m.ChannelID, "Please provide symbol in erc721")
				return
			}

			token, err := login()
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			address, err := deployContract(token, contractName, contents[4], userAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content:   fmt.Sprintf("<@%s> Congratulate on deploying erc721 contract successfully for %s. Your contract address is %s! \n %s", m.Author.ID, userAddress, address, advertise),
				Reference: m.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{nil, nil, []string{m.Author.ID},
				},
			})

		}
	})

	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")
}


func sendEasyMintRequest(token string, dto models.EasyMintMetaDto) (*models.MintResp, error){
	b, err := json.Marshal(dto)
	if err != nil {
		panic(err)
		return nil, err
	}

	req, _ := http.NewRequest("POST", viper.GetString("easyMint.url"), bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
		return  nil, err
	}

	var tmp models.MintTask
	content, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return nil, err
	}

	id, err := getTokenId(tmp.ID, token)
	if err != nil {
		return nil, err
	}

	res := &models.MintResp{
		UserAddress: dto.MintToAddress,
		Advertise: advertise,
		NFTAddress: viper.GetString("mintRespPrefix") + strconv.Itoa(int(id)),
	}

	defer resp.Body.Close()
	return res, nil
}

func sendCustomMintRequest(token string, dto models.CustomMintDto) (*models.MintResp, error){
	b, err := json.Marshal(dto)
	if err != nil {
		panic(err)
		return nil, err
	}

	req, _ := http.NewRequest("POST", viper.GetString("customMint.url"), bytes.NewBuffer(b))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer " + token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
		return  nil, err
	}

	var tmp models.MintTask
	content, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(content, &tmp)
	if err != nil {
		return nil, err
	}

	id, err := getTokenId(tmp.ID, token)
	if err != nil {
		return nil, err
	}

	res := &models.MintResp{
		UserAddress: dto.MintToAddress,
		Advertise: advertise,
		NFTAddress: viper.GetString("mintRespPrefix") +  dto.ContractAddress + "/" + strconv.Itoa(int(id)),
	}

	defer resp.Body.Close()
	return res, nil
}

func createMetadata(token, fileUrl, name, description string) (string, error) {
	metadata := models.Metadata{
		Name: name,
		Description: description,
		Image: fileUrl,
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	req, _ := http.NewRequest("POST", viper.GetString("customMint.metadataUrl"), bytes.NewBuffer(b))
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

	return tmp.MetadataURI, nil
}

func login() (string, error) {
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

func getTokenId(id uint, token string) (uint64, error) {
	t := models.MintTask{}
	for t.TokenId == 0 {
		req, err := http.NewRequest("GET", viper.GetString("infoUrl") + strconv.Itoa(int(id)),nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer " + token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return 0, err
		}

		err = json.Unmarshal(content, &t)
		if err != nil {
			return 0, err
		}
		time.Sleep(10 * time.Second)
	}
	return t.TokenId, nil
}

func uploadFile(token, fileUrl, name string) (string, error){
	proxy := func(_ *http.Request) (*url.URL, error) {
		return url.Parse("http://127.0.0.1:7890")
	}

	transport := &http.Transport{Proxy: proxy}

	client := &http.Client{Transport: transport}
	resp, err := client.Get(fileUrl)

	fmt.Println(fileUrl)
	if err != nil {
		panic(err)
		return "", err
	}
	defer resp.Body.Close()
	contentType := resp.Header.Get("Content-Type")
	fmt.Println(contentType)
	if !strings.Contains(contentType, "image") {
		return "", errors.New("only support to upload images")
	}

	bodyBuffer := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuffer)

	fileWriter, _ := bodyWriter.CreateFormFile("file", name)


	io.Copy(fileWriter, resp.Body)

	contentType1 := bodyWriter.FormDataContentType()
	bodyWriter.Close()

	req, _ := http.NewRequest("POST", viper.GetString("fileUploadUrl"), bodyBuffer)
	fmt.Println(req)

	req.Header.Add("Authorization", "Bearer " + token)
	req.Header.Add("content-type", contentType1)
	fmt.Println(contentType1)
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

	fmt.Println(t.FileUrl)

	return t.FileUrl, nil
}

func deployContract(token, name, symbol, owner string) (string, error){
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

