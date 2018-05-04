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
	"log"
	"os"

	"github.com/javefang/kaptain/pkg/api"
	"github.com/spf13/cobra"
)

// provisionCmd represents the provision command
var provisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Prepare the current Kubernetes node",
	Long: `Follow Kaptain's instruction to download the correct config files 
	and TLS assets to configure the current Kubernetes node`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if sailorClient.ClusterName == "" {
			return fmt.Errorf("--name must be set")
		}

		switch sailorClient.Role {
		case "etcd":
		case "master":
		case "worker":
		default:
			return fmt.Errorf("--role must be one of etcd, master or worker")
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		sailorClient.Registry = api.NewClusterRegistry(storeUrl)

		if err := sailorClient.Provision(); err != nil {
			log.Fatalf("failed to provision node: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(provisionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will provision for this command
	// and all subcommands, e.g.:
	// provisionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// provisionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	provisionCmd.Flags().StringVar(&sailorClient.Role, "role", "", "Sailor role ('etcd', 'master' or 'worker')")
	provisionCmd.Flags().StringVar(&sailorClient.Prefix, "prefix", "/", "Base directory for writing all the files to")
}
