# DiscordBot

# DiscordBot
DiscordNFTBot
## Description

This is a NFTRainbow-based discord bot, which helps users in the discord to mint NFTs easily. On the other hand, this bot can be used on the community activities to increase community activity.

## References
[NFTRainbow Console](https://console.nftrainbow.xyz/login)
[NFTRainbow Doc](https://docs.nftrainbow.xyz/)
[NFTRainbow Git](https://github.com/nft-rainbow)

## Functions
For the admin of the bot, he can choose to use the default erc721 contract, or to deploy his own erc721 or erc1155 contract. To achieve this target, the admin can use the provided CMD to upload file to 
obtain the `file_url`, which is used by the admin to deploy the contract through the provided CMD. 

For the users in the discord channel where the bot is deployed, they can mint their own NFTs through the contract provided by the admin.

## Run
### CMD
````
cd botCMD
````
Generate the `config.yaml`
````
cp config-sample.yaml config.yaml
````
Input the `app_id` and `app_secret`, which can be obtained from the [NFTRainbow console](https://console.nftrainbow.xyz/login)

Generate the binary file 
````
make build
````
Upload file to server to obtain the `file_url`
````
botCMD upload [file_path]
# file_path is the uploaded file path
````
Deploy the contract
````
botCMD deploy [name] [symbol] [type] [appAddress]
# name if the name of the contract
# symbol is the symbol of the NFT
# type is the type of the contract including erc721 and erc1155
# appAddress is the address of the NFTRainbow app`,
````

### Bot configuration
Generate the `config.yaml`
````
cp config-sample.yaml config.yaml
````
Config the yaml 
- Input the `app_id` and `app_secret`
- Input the `token` which can be obtained from the discord. This can refer to <https://www.writebots.com/discord-bot-token/>
- Input the default mint configuration including `file_url`, `name`, `description`

Run the project 
````
go run main.go
````

### How to mint the NFTs
#### EasyMint
After the users in the discord channel can input the `/claim [user_address] (file_url)` to the chat frame, the bot will return the NFT information in several seconds.

|  Parameters Name   | Meaning  | Required or Optional | 
|  ----  | ----  | ---- | 
| user_address  | The blockchain address of the user |required |
| file_url | The file_url of the metadata |optional |

#### CustomMint
After the users in the discord channel can input the `/claim customNFT [user_address] [contract_address] [name] [description] (file_url)` to the chat frame, the bot will return the NFT information in several seconds.

|  Parameters Name   | Meaning  | Required or Optional | 
|  ----  | ----  | ---- | 
| user_address  | The blockchain address of the user |required |
| contract_address  | The contract address of the user |required |
| name  | The name of the metadata |required |
| description | The description of the metadata |required |
| file_url | The file_url of the metadata |optional |
## Supported Chains
<https://docs.nftrainbow.xyz/docs/faqs#mu-qian-nftrainbow-zhi-chi-na-xie-lian:~:text=FAQs-,%E7%9B%AE%E5%89%8D%20NFTRainbow%20%E6%94%AF%E6%8C%81%E5%93%AA%E4%BA%9B%E9%93%BE%3F,-%E6%A0%91%E5%9B%BE%E9%93%BE>