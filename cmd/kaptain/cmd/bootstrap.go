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
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/kaptain"
)

// bootstrapCmd represents the bootstrap command
var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap the cluster",
	Long: `Bootstrap the a fresh cluster deployed by Terraform.
This command perform the following:
- Configure cluster networking with Calico
- Configure RBAC permissions
- Configure storageclass
- Configure limit-range
- Install Heapster

$ kaptain bootstrap -n dev.test.waws
`,
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.Flags()

		clusterName, err := flagset.GetString("name")
		if err != nil {
			panic(err)
		}

		client := kaptain.KaptainClient{
			Registry: api.NewClusterRegistry(storeUrl),
		}

		if err := client.Bootstrap(clusterName); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(bootstrapCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// bootstrapCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// bootstrapCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	bootstrapCmd.Flags().StringP("name", "n", "", "Cluster name of the cluster to be bootstrapped")

	bootstrapCmd.MarkFlagRequired("name")
}
