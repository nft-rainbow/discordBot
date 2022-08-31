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
````

|  Parameters Name   | Meaning  | Required or Optional | 
|  ----  | ----  | ---- | 
| name  | The name of the NFT |required |
| symbol | The symbol of the NFT |required |
| type | The type of contracts including erc721 and erc1155 |required |
| appAddress | The address of the app account,which can be obtained from the NFTRainbow console |required |

### Bot configuration
Generate the `config.yaml`
````
cp config-sample.yaml config.yaml
````
Config the yaml 
- Input the `app_id` and `app_secret`
- Input the `botToken` which can be obtained from the discord. This can refer to <https://www.writebots.com/discord-bot-token/>
- Input the default mint configuration including `file_url`, `name`, `description` and so on.
- If the admin of the bot want to use his own contract to mint, the `contractAddress` is required to call customMint. Please input the parameter.

Run the project 
````
go run main.go
````

### How to mint the NFTs
#### EasyMint
After the users in the discord channel can input the `/claim easy-mint [user_address]` to the chat frame, the bot will return the NFT information in several seconds.

|  Parameters Name   | Meaning  | Required or Optional | 
|  ----  | ----  | ---- | 
| user_address  | The blockchain address of the user |required |

#### CustomMint
After the users in the discord channel can input the `/claim custom-mint [user_address]` to the chat frame, the bot will return the NFT information in several seconds.

|  Parameters Name   | Meaning  | Required or Optional | 
|  ----  | ----  | ---- | 
| user_address  | The blockchain address of the user |required |

## Supported Chains
[Present Supported Chains](https://docs.nftrainbow.xyz/docs/faqs#mu-qian-nftrainbow-zhi-chi-na-xie-lian:~:text=FAQs-,%E7%9B%AE%E5%89%8D%20NFTRainbow%20%E6%94%AF%E6%8C%81%E5%93%AA%E4%BA%9B%E9%93%BE%3F,-%E6%A0%91%E5%9B%BE%E9%93%BE)
