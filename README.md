# DiscordBot
This is a discord bot, which helps users in the discord to mint NFTs easily.

# Run
cp `config.yaml`
````
cp config-sample.yaml config.yaml
````

Config the yaml 
- Input the `app_id` and `app_secret`
- Input the `token` which can be obtained from the discord
- Input the mint configuration including `file_url`, `name`, `description`

Run the project 
````
go run main.go
````

# How to mint the NFTs
Input the `/claim <user_address>` in the target discord server, the bot will return the NFT info after several seconds. 