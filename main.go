package main

import (
	"fmt"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/nft-rainbow/discordBot/service"
	"github.com/nft-rainbow/discordBot/utils"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

var easyMintRestrain map[string]int = make(map[string]int)
var customRestrain map[string]int = make(map[string]int)

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
			token, err := service.Login()
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

			resp , err := service.SendEasyMintRequest(token, models.EasyMintMetaDto{
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

			customRestrain[userAddress] ++
			token, err := service.Login()
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

			metadataUri, err := service.CreateMetadata(token, fileUrl, name, description)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				customRestrain[userAddress] --
				return
			}

			resp , err := service.SendCustomMintRequest(token, models.CustomMintDto{
				Chain: viper.GetString("customMint.chain"),
				MintToAddress: userAddress,
				ContractAddress: contractAddress,
				MetadataUri: metadataUri,
				ContractType: viper.GetString("customMint.type"),
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






