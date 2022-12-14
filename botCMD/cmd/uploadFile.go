/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/nft-rainbow/discordBot/service"
	"github.com/spf13/cobra"
)

var uploadFileCmd = &cobra.Command{
	Use:   "upload",
	Short: "upload your file to obtain the file_url",
	Long: `In order to config the mint service in discord, the admin of the bot can choose to upload its own file to NFTRainbow server to obtain the file_url through this cmd.`,
	Example: `botCMD upload [file_path]
- file_path The path of the uploaded file`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token, err := service.Login()
		if err != nil {
			fmt.Println(err)
			return
		}
		fileUrl, err := service.UploadFile(token, args[0])
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(fileUrl)
	},
}

func init() {
	rootCmd.AddCommand(uploadFileCmd)
}
