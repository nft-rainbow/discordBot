package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/nft-rainbow/discordBot/database"
	"github.com/nft-rainbow/discordBot/models"
	"github.com/nft-rainbow/discordBot/service"
	"github.com/nft-rainbow/discordBot/utils"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
)
var s *discordgo.Session

func initConfig() {
	viper.SetConfigName("config")             // name of config file (without extension)
	viper.SetConfigType("yaml")               // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(".")                  // optionally look for config in the working directory
	err := viper.ReadInConfig()               // Find and read the config file
	if err != nil {                           // Handle errors reading the config file
		log.Fatalln(fmt.Errorf("fatal error config file: %w", err))
	}
}

var (
	integerOptionMinValue = 1.0

	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "claim",
			Description: "Command for claiming NFTs",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "custom-mint",
					Description: "Mint a nft through the contract deployed by the admin",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "user_address",
							Description: "The address of the user",
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
				{
					Name:        "easy-mint",
					Description: "Mint a nft through the NFTfactory contract owned by NFTRainbow",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "user_address",
							Description: "The address of the user",
							Required:    true,
						},
					},
					Type: discordgo.ApplicationCommandOptionSubCommand,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"claim": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			options := i.ApplicationCommandData().Options
			resp := ""
			userAddress := options[0].Options[0].Value.(string)
			startFlag := ""
			switch options[0].Name {
			case "custom-mint":
				startFlag = "Start to mint using custom-mint model. Please wait patiently."
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: startFlag,
					},
				})
				resp = handleCustomMint(userAddress)
			case "easy-mint":
				startFlag = "Start to mint using easy-mint model. Please wait patiently."
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: startFlag,
					},
				})
				resp = handleEasyMint(userAddress)
			}

			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: resp,
			})
		},
	}
)

func init() {
	initConfig()
	var err error
	s, err = discordgo.New("Bot " + viper.GetString("botToken"))
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
	database.ConnectDB()
}


func init() {
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %v#%v", s.State.User.Username, s.State.User.Discriminator)
	})
	err := s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
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

func handleCustomMint(userAddress string) string{
	_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
	if err != nil {
		return err.Error()
	}

	contractAddress := viper.GetString("customMint.contractAddress")
	_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, contractAddress)
	if err != nil {
		return err.Error()
	}

	err = checkRestrain(userAddress, database.CustomMintBucket)
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.CustomMintBucket)
		return err.Error()
	}

	token, err := service.Login()
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.CustomMintBucket)
		return err.Error()
	}

	metadataUri, err := service.CreateMetadata(token, viper.GetString("customMint.fileUrl"), viper.GetString("customMint.name"), viper.GetString("customMint.description"))
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.CustomMintBucket)
		return err.Error()
	}
	resp , err := service.SendCustomMintRequest(token, models.CustomMintDto{
		models.ContractInfoDto{
			Chain: viper.GetString("chainType"),
			ContractType: viper.GetString("customMint.contractType"),
			ContractAddress: contractAddress,
		},
		models.MintItemDto{
			MintToAddress: userAddress,
			MetadataUri: metadataUri,
		},
	})
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.CustomMintBucket)
		return err.Error()
	}

	return fmt.Sprintf("Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", resp.UserAddress, resp.NFTAddress, resp.Advertise)
}

func handleEasyMint(userAddress string)string {
	_, err := utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
	if err != nil {

		return err.Error()
	}
	err = checkRestrain(userAddress, database.EasyMintBucket)
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.EasyMintBucket)
		return err.Error()
	}

	token, err := service.Login()
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.EasyMintBucket)
		return err.Error()
	}

	resp , err := service.SendEasyMintRequest(token, models.EasyMintMetaDto{
		Chain: viper.GetString("chainType"),
		Name: viper.GetString("easyMint.name"),
		Description: viper.GetString("easyMint.description"),
		MintToAddress: userAddress,
		FileUrl: viper.GetString("easyMint.fileUrl"),
	})
	if err != nil {
		_ = database.InsertDB(userAddress, []byte("0"), database.EasyMintBucket)
		return err.Error()
	}
	return fmt.Sprintf("Congratulate on minting NFT for %s successfully. Check this link to view it: %s \n  %s", resp.UserAddress, resp.NFTAddress, resp.Advertise)
}








