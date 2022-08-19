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
  "github.com/spf13/cobra"
  "log"
  "os"

  "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
  Use:   "botCMD",
  Short: "A brief description of your application",
  Long: `The bot cmd helps the admin of the bot to upload `,
}


func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
  cobra.OnInitialize(initConfig)
}


// initConfig reads in config file and ENV variables if set.
func initConfig() {
  viper.SetConfigName("config")             // name of config file (without extension)
  viper.SetConfigType("yaml")               // REQUIRED if the config file does not have the extension in the name
  viper.AddConfigPath(".")                  // optionally look for config in the working directory
  viper.AddConfigPath("..")
  err := viper.ReadInConfig()               // Find and read the config file
  if err != nil {                           // Handle errors reading the config file
    log.Fatalln(fmt.Errorf("fatal error config file: %w", err))
  }
}

func init() {
  initConfig()
}

