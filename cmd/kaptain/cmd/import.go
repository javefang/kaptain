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
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/kaptain"
)

var importInflateClusterOpts kaptain.InflateClusterOptions

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import a cluster from a file",
	Long: `Import a cluster spec from a file.
For example, to import all cluster spec from a file 'cluster.yaml' on local disk:
	
$ kaptain import -f cluster.yaml

The cluster name is determined by the metadata in 'cluster.yaml'. The command
will throw error if a cluster with the same name already exists in the registry.

For how to generate the 'cluster.yaml' file, see 'kaptain export -h'.
`,
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.Flags()

		inFile, err := flagset.GetString("file")
		if err != nil {
			panic(err)
		}

		// read the file
		data, err := ioutil.ReadFile(inFile)
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}

		// parse the cluster
		cluster := api.Cluster{}
		if err = yaml.Unmarshal(data, &cluster); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		kaptain.InflateCluster(&cluster, &importInflateClusterOpts)

		// create the cluster if it doesn't exist on the registry yet
		client := kaptain.KaptainClient{
			Registry: api.NewClusterRegistry(storeUrl),
		}

		if _, err := client.Get(cluster.Name); err == nil {
			log.Errorf("Unable to import cluster: a cluster with the same name already exists")
			os.Exit(1)
		}

		if err := client.Create(&cluster, true); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	importCmd.Flags().StringP("file", "f", "", "Cluster spec file to be applied")
	importCmd.Flags().BoolVar(&importInflateClusterOpts.UpdateSpec, "update-spec", false, "Update missing cluster spec")
	importCmd.Flags().BoolVar(&importInflateClusterOpts.UpdatePKIs, "update-pkis", false, "Update missing PKIs")
	importCmd.Flags().BoolVar(&importInflateClusterOpts.UpdateTokens, "update-tokens", false, "Update missing tokens")
	importCmd.Flags().BoolVar(&importInflateClusterOpts.UpdateAssetManifest, "update-asset-manifest", false, "Update asset manifest")
	importCmd.MarkFlagRequired("file")
}
