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
	"os/user"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/kaptain"
)

// exportConfigCmd represents the export command
var exportConfigCmd = &cobra.Command{
	Use:   "export-config",
	Short: "Export kubeconfig for accessing a cluster",
	Long: `Export credentials of a specific user for accessing a cluster.
For example, the following command will export the credential of user
'viewer' on cluster 'dev.test.waws'

$ kaptain export-config --name=dev.test.waws --user=admin

By default, the config is exported to '~/.kube/config'. If the target file
already exists, Kaptain will merge in the new config and set the context.
Otherwise, it will create a new file.
`,
	Run: func(cmd *cobra.Command, args []string) {
		flagset := cmd.Flags()

		clusterName, err := flagset.GetString("name")
		if err != nil {
			panic(err)
		}

		username, err := flagset.GetString("user")
		if err != nil {
			panic(err)
		}

		kubeconfig, err := flagset.GetString("kubeconfig")
		if err != nil {
			panic(err)
		}

		force, err := flagset.GetBool("force")
		if err != nil {
			panic(err)
		}

		client := kaptain.KaptainClient{
			Registry: api.NewClusterRegistry(storeUrl),
		}

		if err := client.ExportConfig(clusterName, kubeconfig, username, force); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(exportConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	homeDir, err := getHomeDir()
	if err != nil {
		log.Fatalf("failed to get home directory")
		os.Exit(1)
	}
	defaultKubeConfigFile := path.Join(homeDir, ".kube", "config")

	exportConfigCmd.Flags().StringP("name", "n", "", "Cluster name of the credential to be exported")
	exportConfigCmd.Flags().StringP("user", "u", "admin", "Username of the credential to be exported")
	exportConfigCmd.Flags().StringP("kubeconfig", "k", defaultKubeConfigFile, "specify path to the output kubeconfig")
	exportConfigCmd.Flags().BoolP("force", "f", false, "overwrite existing kubeconfig")

	exportConfigCmd.MarkFlagRequired("name")
}

func getHomeDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	return usr.HomeDir, nil
}
