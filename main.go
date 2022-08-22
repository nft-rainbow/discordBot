package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/nft-rainbow/discordBot/database"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/nft-rainbow/discordBot/service"
	"github.com/nft-rainbow/discordBot/utils"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)


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
	database.ConnectDB()
	s, _ := discordgo.New("Bot " + viper.GetString("botToken"))
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")
	})
	s.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// /claim easyMint <userAddress>
		if strings.Contains(m.Content, "/claim easyMint") {
			userAddress := strings.Split(m.Content, " ")[2]
			_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
			if err != nil {
				_, _ = s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}
			err = checkRestrain(userAddress, database.EasyMintBucket)
			if err != nil {
				processErrorMessage(s,m, err.Error(), "", nil)
				return
			}

			token, err := service.Login()
			if err != nil {
				processErrorMessage(s,m, err.Error(), userAddress, database.EasyMintBucket)
				return
			}

			resp , err := service.SendEasyMintRequest(token, models.EasyMintMetaDto{
				Chain: viper.GetString("chainType"),
				Name: viper.GetString("easyMint.name"),
				Description: viper.GetString("easyMint.description"),
				MintToAddress: userAddress,
				FileUrl: viper.GetString("easyMint.fileUrl"),
			})
			if err != nil {
				processErrorMessage(s,m, err.Error(), userAddress, database.EasyMintBucket)
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
		// /claim customNFT <userAddress>
		if strings.Contains(m.Content, "/claim customNFT") {
			contents := strings.Split(m.Content, " ")
			userAddress := contents[2]
			_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
			if err != nil {
				processErrorMessage(s,m, err.Error(), "", database.CustomMintBucket)
				return
			}

			contractAddress := viper.GetString("customMint.contractAddress")
			_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, contractAddress)
			if err != nil {
				processErrorMessage(s,m, err.Error(), "", database.CustomMintBucket)
				return
			}

			err = checkRestrain(userAddress, database.CustomMintBucket)
			if err != nil {
				processErrorMessage(s,m, err.Error(), userAddress, database.CustomMintBucket)
				return
			}

			token, err := service.Login()
			if err != nil {
				processErrorMessage(s,m, err.Error(), userAddress, database.CustomMintBucket)
				return
			}

			metadataUri, err := service.CreateMetadata(token, viper.GetString("customMint.fileUrl"), viper.GetString("customMint.name"), viper.GetString("customMint.description"))
			if err != nil {
				processErrorMessage(s,m, err.Error(), userAddress, database.CustomMintBucket)
				return
			}

			resp , err := service.SendCustomMintRequest(token, models.CustomMintDto{
				Chain: viper.GetString("chainType"),
				MintToAddress: userAddress,
				ContractAddress: contractAddress,
				MetadataUri: metadataUri,
				ContractType: viper.GetString("customMint.contractType"),
			})
			if err != nil {
				processErrorMessage(s,m, err.Error(), userAddress, database.CustomMintBucket)
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

func checkRestrain(address string, mintType []byte) error{
	count, err := database.GetCount(address, mintType)
	if err != nil {
		return err
	}
	if count == nil {
		err = database.InsertDB(address, []byte("1"), mintType)
		if err != nil {
			return err
		}
		return nil
	}

	if !bytes.Equal(count, []byte("0")) {
		return errors.New("This address has minted the NFT")
	}

	return nil
}

func processErrorMessage(s *discordgo.Session, m *discordgo.MessageCreate, message, address string, mintType []byte) {
	_, _ = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content:   fmt.Sprintf("<@%s> %s", m.Author.ID, message),
		Reference: m.Reference(),
		AllowedMentions: &discordgo.MessageAllowedMentions{nil, nil, []string{m.Author.ID},
		},
	})
	if address != "" {
		_ = database.InsertDB(address, []byte("0"), mintType)
	}
	return
}








