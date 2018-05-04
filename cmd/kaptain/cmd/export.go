// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
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

	"github.com/ghodss/yaml"

	"github.com/spf13/cobra"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/kaptain"
)

// exportCmd represents the get command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export a cluster spec to a file",
	Long: `Export a cluster spec to a file.
	
	Foe example, to export all cluster spec for cluster 'dev.test.waws' to a file 
	'cluster.yaml' on local disk:
	
	$ kaptain export -n dev.test.waws > cluster.yaml
	
	This file contains all cluster spec, PKIs, token secrets and file/addon manifests.
	You can keep it in version control, edit it and recreate a cluster later.
	See 'kaptain import -h'.
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

		cluster, err := client.Get(clusterName)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		data, err := yaml.Marshal(cluster)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(data))
	},
}

func init() {
	RootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	exportCmd.Flags().StringP("name", "n", "", "Cluster Name")

	exportCmd.MarkFlagRequired("name")
}
