package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/nft-rainbow/discordBot/utils"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var games map[string]int = make(map[string]int)

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
		if strings.Contains(m.Content, "/claim") {
			userAddress := strings.Split(m.Content, " ")[1]

			_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			if games[userAddress] > 0 {
				_, _ = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("@%s One account can only obtain a NFT", m.Author.Username))
				return
			}

			games[userAddress] ++
			token, err := login()
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				games[userAddress] --
				return
			}
			if token == "" {
				_, _ = s.ChannelMessageSend(m.ChannelID, "get access token failed")
				games[userAddress] --
				return
			}

			resp , err := sendMintRequest(token, models.EasyMintMetaDto{
				Chain: viper.GetString("easyMint.chain"),
				Name: viper.GetString("easyMint.name"),
				Description: viper.GetString("easyMint.description"),
				MintToAddress: userAddress,
				FileUrl: viper.GetString("easyMint.fileUrl"),
			})
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				games[userAddress] --
				return
			}
			_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Content:   fmt.Sprintf("@%s UserAddress: %s \n NFTAddress: %s \n %s", m.Author.Username, resp.UserAddress, resp.NFTAddress, resp.Advertise),
				Reference: m.Reference(),
				AllowedMentions: &discordgo.MessageAllowedMentions{nil, nil, []string{m.Author.ID},
				},
			})
			//_, _ = s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", resp.UserAddress, resp.NFTAddress, resp.Advertise), m.Reference())

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

func sendMintRequest(token string, dto models.EasyMintMetaDto) (*models.MintResp, error){
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
		Advertise: "Powered by NFTRainbow",
		NFTAddress: viper.GetString("easyMint.mintRespPrefix") + strconv.Itoa(int(id)),
	}

	defer resp.Body.Close()
	return res, nil
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
		req, err := http.NewRequest("GET", viper.GetString("easyMint.infoUrl") + strconv.Itoa(int(id)),nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer " + token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
			return 0, err
		}
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)

			return 0, err
		}

		fmt.Println(string(content))

		err = json.Unmarshal(content, &t)
		if err != nil {
			panic(err)
			return 0, err
		}
		time.Sleep(10 * time.Second)
	}
	return t.TokenId, nil
}

