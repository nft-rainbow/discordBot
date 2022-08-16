package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/nft-rainbow/discordBot/models"
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


const timeout time.Duration = time.Second * 10

var games map[string]time.Time = make(map[string]time.Time)

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
	//models.ConnectDB()
}

func main() {
	s, _ := discordgo.New("Bot " + viper.GetString("bot.token"))
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if strings.Contains(m.Content, "/claim") {
			token, err := login()
			if token == "" {
				_, _ = s.ChannelMessageSend(m.ChannelID, "get access token failed")
				return
			}
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			userAddress := strings.Split(m.Content, " ")[1]
			resp , err := sendMintRequest(token, models.EasyMintMetaDto{
				Chain: viper.GetString("easymint.chain"),
				Name: viper.GetString("easymint.name"),
				Description: viper.GetString("easymint.description"),
				MintToAddress: userAddress,
				FileUrl: viper.GetString("easymint.file_url"),
			})
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			_, _ = s.ChannelMessageSendReply(m.ChannelID, fmt.Sprintf("UserAddress: %s \n NFTAddress: %s \n %s", resp.UserAddress, resp.NFTAddress, resp.Advertise), m.Reference())

			games[m.ChannelID] = time.Now()
			<-time.After(timeout)
			if time.Since(games[m.ChannelID]) >= timeout {
				archived := true
				locked := true
				_, err := s.ChannelEditComplex(m.ChannelID, &discordgo.ChannelEdit{
					Archived: archived,
					Locked:   locked,
				})
				if err != nil {
					panic(err)
				}
			}
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

	req, _ := http.NewRequest("POST", viper.GetString("easymint.url"), bytes.NewBuffer(b))
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
		NFTAddress: viper.GetString("easymint.mintRespPrefix") + strconv.Itoa(int(id)),
	}

	defer resp.Body.Close()
	return res, nil
}

func login() (string, error) {
	data := make(map[string]string)
	data["app_id"] = viper.GetString("app.app_id")
	data["app_secret"] = viper.GetString("app.app_secret")
	b, _ := json.Marshal(data)

	req, err := http.NewRequest("POST", viper.GetString("app.login_url"), bytes.NewBuffer(b))
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
		req, err := http.NewRequest("GET", viper.GetString("getInfo_url") + strconv.Itoa(int(id)),nil)
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

