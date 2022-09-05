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
			var resp *models.MintResp
			userAddress := options[0].Options[0].Value.(string)
			startFlag := ""
			var err error
			switch options[0].Name {
			case "custom-mint":
				startFlag = "Start to mint using custom-mint model. Please wait patiently."
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: startFlag,
					},
				})
				resp, err = handleCustomMint(userAddress)
			case "easy-mint":

				startFlag = "Start to mint using easy-mint model. Please wait patiently."
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: startFlag,
					},
				})

				resp, err = handleEasyMint(userAddress)
			}
			if err != nil {
				s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Embeds: failMessageEmbed(err.Error()),
				})
				return
			}
			//button := discordgo.Button{
			//	Label: "VIEW IN CONFLUX SCAN",
			//	Style: discordgo.LinkButton,
			//	URL: resp.NFTAddress,
			//	Disabled: false,
			//}

			s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				//Components: []discordgo.MessageComponent{
				//	discordgo.ActionsRow{
				//		Components: []discordgo.MessageComponent{button},
				//	},
				//},
				Embeds: successfulMessageEmbed(resp),
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
	status, err := database.GetStatus(address, mintType)
	if err != nil {
		return err
	}

	if bytes.Equal(status, []byte("Success")) {
		return errors.New("This account has minted NFT")
	}
	if bytes.Equal(status, []byte("Minting")) {
		return errors.New("This account is minting NFT")
	}

	return nil
}

func handleCustomMint(userAddress string) (*models.MintResp, error){
	var err error
	defer func() {
		status, _ := database.GetStatus(userAddress, database.CustomMintBucket)
		if err != nil && !bytes.Equal(status, []byte("Success")) {
			_ = database.InsertDB(userAddress, []byte("NoMinting"), database.CustomMintBucket)
		}
	}()
	_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
	if err != nil {
		return nil, err
	}

	contractAddress := viper.GetString("customMint.contractAddress")
	_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, contractAddress)
	if err != nil {
		return nil, err
	}

	err = checkRestrain(userAddress, database.CustomMintBucket)
	if err != nil {
		return nil, err
	}
	_ = database.InsertDB(userAddress, []byte("Minting"), database.CustomMintBucket)

	token, err := service.Login()
	if err != nil {
		return nil, err
	}

	metadataUri, err := service.CreateMetadata(token, viper.GetString("customMint.fileUrl"), viper.GetString("customMint.name"), viper.GetString("customMint.description"))
	if err != nil {
		return nil, err
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
		return nil, err
	}
	_ = database.InsertDB(userAddress, []byte("Success"), database.CustomMintBucket)

	return resp, err
}

func handleEasyMint(userAddress string)(*models.MintResp, error) {
	var err error
	defer func() {
		status, _ := database.GetStatus(userAddress, database.EasyMintBucket)
		if err != nil && !bytes.Equal(status, []byte("Success")) {
			_ = database.InsertDB(userAddress, []byte("NoMinting"), database.EasyMintBucket)
		}
	}()
	_, err = utils.CheckCfxAddress(utils.CONFLUX_TEST, userAddress)
	if err != nil {
		return nil, err
	}
	err = checkRestrain(userAddress, database.EasyMintBucket)
	if err != nil {
		return nil, err
	}
	_ = database.InsertDB(userAddress, []byte("Minting"), database.EasyMintBucket)

	token, err := service.Login()
	if err != nil {
		return nil, err
	}

	resp , err := service.SendEasyMintRequest(token, models.EasyMintMetaDto{
		Chain: viper.GetString("chainType"),
		Name: viper.GetString("easyMint.name"),
		Description: viper.GetString("easyMint.description"),
		MintToAddress: userAddress,
		FileUrl: viper.GetString("easyMint.fileUrl"),
	})
	if err != nil {
		return nil, err
	}
	_ = database.InsertDB(userAddress, []byte("Success"), database.EasyMintBucket)
	return resp, nil
}

func successfulMessageEmbed(resp *models.MintResp) []*discordgo.MessageEmbed{
	embeds := []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type: discordgo.EmbedTypeRich,
			Title: ":rainbow: Mint NFT successfully  :rainbow:",
			Description: "Congratulate on minting NFT successfully! The NFT information is showed in the following.",
			Image: &discordgo.MessageEmbedImage{
				URL: "https://img0.baidu.com/it/u=2475308105,1312864556&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=889",
			},
			Provider: &discordgo.MessageEmbedProvider{
				Name: "come",
				URL: "https://img0.baidu.com/it/u=2475308105,1312864556&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=889",
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name: "Mints Time",
					Value: resp.Time,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "Contract",
					Value: resp.Contract,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "Token ID",
					Value: resp.TokenID,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name: "NFT URL",
					Value: fmt.Sprintf("[VIEW IN CONFLUX SCAN](%s)", resp.NFTAddress),
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name: "Advertise",
					Value: viper.GetString("advertise"),
					Inline: false,
				},
			},
			Author: &discordgo.MessageEmbedAuthor{
				Name: "NFTRainbow",
				URL: "https://docs.nftrainbow.xyz/",
				IconURL: "https://img0.baidu.com/it/u=2475308105,1312864556&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=889",
			},
		},
	}

	return embeds
}

func failMessageEmbed(message string) []*discordgo.MessageEmbed{
	embeds := []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type: discordgo.EmbedTypeRich,
			Title: ":scream: Failed to Mint NFT  :scream:",
			Description: "There is problem during minting NFT. ",
			Image: &discordgo.MessageEmbedImage{
				URL: "https://gimg2.baidu.com/image_search/src=http%3A%2F%2Ftva1.sinaimg.cn%2Fbmiddle%2F006APoFYly1g55m70z1uvj30hs0hidhd.jpg&refer=http%3A%2F%2Ftva1.sinaimg.cn&app=2002&size=f9999,10000&q=a80&n=0&g=0n&fmt=auto?sec=1664935347&t=223d106a8cbc9c825b5a34ff36b3678c",
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name: "Error message",
					Value: message,
					Inline: false,
				},
				&discordgo.MessageEmbedField{
					Name: "Advertise",
					Value: viper.GetString("advertise"),
					Inline: false,
				},
			},
			Author: &discordgo.MessageEmbedAuthor{
				Name: "NFTRainbow",
				URL: "https://docs.nftrainbow.xyz/",
				IconURL: "https://img0.baidu.com/it/u=2475308105,1312864556&fm=253&fmt=auto&app=138&f=JPEG?w=500&h=889",
			},
		},
	}

	return embeds
}








