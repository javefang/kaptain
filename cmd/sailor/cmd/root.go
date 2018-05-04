// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/javefang/kaptain/pkg/sailor"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var storeUrl string
var sailorClient sailor.SailorClient

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "sailor",
	Short: "Follow cluster created by Kaptain and provision config files.",
	Long: `
Sailor relies on a backend store to save state. The default is "s3://aws.all.kaptain?region=eu-west-1".
To change it, set environment variable "KAPTAIN_STORE" to the desired store URL. E.g. The following example
set Sailor to use Vault as the store.

$ export VAULT_ADDR="https://vault.service.consul:8200"
$ export KAPTAIN_STORE="vault://project/kaptain?role_id=1234-1234-1234-1234&secret_id=<redacted>"
$ sailor provision --role=etcd -n dev.example.com

For details usage, please see help of each sub-command.
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.sailor.yaml)")
	RootCmd.PersistentFlags().StringVarP(&sailorClient.ClusterName, "name", "n", "", "Cluster name")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// log settings
	log.SetLevel(log.InfoLevel)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".sailor" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".sailor")
	}

	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("kaptain")
	viper.SetDefault("store", "s3://aws.all.kaptain?region=eu-west-1")
	viper.SetDefault("log", "info")

	// store settings
	storeUrl = viper.GetString("store")

	// log settings
	switch logLevel := viper.GetString("log"); logLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		panic(fmt.Errorf("Unknown log level: %s", logLevel))
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
