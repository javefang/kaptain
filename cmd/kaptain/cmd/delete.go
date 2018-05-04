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
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/kaptain"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a cluster spec from the registry",
	Long: `Delete a cluster spec from the registry.
This will delete the cluster spec from the registry. 
The operation cannot be undone.	

$ kaptain delete -n dev.example.com
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

		if _, err := client.Get(clusterName); err != nil {
			log.Errorf("Unable to delete cluster: cluster '%s' not found", clusterName)
			os.Exit(1)
		}

		if err := client.Delete(clusterName); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	deleteCmd.Flags().StringP("name", "n", "", "Cluster Name")

	deleteCmd.MarkFlagRequired("name")
}
