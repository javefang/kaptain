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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/kaptain"
)

var newCluster api.Cluster
var etcdServers string
var authenticationTokenWebhookConfigFile string

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new cluster",
	Long: `Create a new Kubernetes Cluster. This will generate all TLS assets and
config files required by Kubernetes, upload them to a store (see "kaptain -h") 
to be used later.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if newCluster.ObjectMeta.Name == "" {
			return fmt.Errorf("--name must be set")
		}

		if newCluster.Spec.DNSDomain == "" {
			log.Infof("--dns-domain not specified, using cluster name '%s'", newCluster.ObjectMeta.Name)
			newCluster.Spec.DNSDomain = newCluster.ObjectMeta.Name
		}

		if newCluster.Spec.MasterPublicName == "" {
			log.Infof("--apiserver not specified, using default 'api.%s'", newCluster.ObjectMeta.Name)
			newCluster.Spec.MasterPublicName = fmt.Sprintf("api.%s", newCluster.ObjectMeta.Name)
		}

		// Cloud provider
		switch newCluster.Spec.CloudProvider {
		case "aws":
		case "vsphere":
			vopts := newCluster.Spec.VSphereOpts
			if vopts.Username == "" {
				return fmt.Errorf("--vsphere-username must be set")
			}
			if vopts.Password == "" {
				return fmt.Errorf("--vsphere-password must be set")
			}
			if vopts.Server == "" {
				return fmt.Errorf("--vsphere-server must be set")
			}
			if vopts.DataCenter == "" {
				return fmt.Errorf("--vsphere-datacenter must be set")
			}
			if vopts.DataStore == "" {
				return fmt.Errorf("--vsphere-datastore must be set")
			}
			if vopts.WorkingDir == "" {
				return fmt.Errorf("--vsphere-workingdir must be set")
			}
			// create cloud-config file at /var/lib/kubernetes/cloud.conf
			newCluster.Spec.CloudConfig = "/var/lib/kubernetes/cloud.conf"
			newCluster.Spec.WorkerCloudConfig = "/var/lib/kubelet/cloud.conf" // TODO: document that this should be provided by orchestration tool that deploys the node
		default:
			return fmt.Errorf("--cloud-provider must be one of 'aws' or 'vsphere'")
		}

		// Authentication token webhook
		if authenticationTokenWebhookConfigFile != "" {
			webhookConfig, err := ioutil.ReadFile(authenticationTokenWebhookConfigFile)
			if err != nil {
				return err
			}
			webhookConfigData := base64.StdEncoding.EncodeToString(webhookConfig)
			// TODO: validate the schema of "webhookConfig" (currently stored as []byte only)
			newCluster.Spec.AuthenticationTokenWebhookOpts.ConfigDataBase64 = webhookConfigData
		}

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// create cluster
		client := kaptain.KaptainClient{
			Registry: api.NewClusterRegistry(storeUrl),
		}

		newCluster.Spec.EtcdCluster = newEtcdCluster(etcdServers)

		inflateOptions := kaptain.InflateClusterOptions{
			UpdateSpec:          true,
			UpdatePKIs:          true,
			UpdateTokens:        true,
			UpdateAssetManifest: true,
		}
		kaptain.InflateCluster(&newCluster, &inflateOptions)

		if err := client.Create(&newCluster, false); err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
	newCluster = api.NewCluster()
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	createCmd.Flags().StringVarP(&newCluster.ObjectMeta.Name, "name", "n", "", "Cluster Name")
	createCmd.Flags().StringVar(&newCluster.Spec.KubeVersion, "kube-version", kaptain.DefaultKubeVersion, "Specify Kubernetes Version")
	createCmd.Flags().StringVar(&newCluster.Spec.DNSDomain, "dns-domain", "", "DNS Domain (default <cluster_name>)")
	createCmd.Flags().StringVar(&newCluster.Spec.CloudProvider, "cloud-provider", kaptain.DefaultCloudProvider, "Cloud Provider (aws or vsphere)")
	createCmd.Flags().StringVar(&etcdServers, "etcd-servers", "etcd-k8s-0,etcd-k8s-1,etcd-k8s-2", "Comma-separated ETCD server hostnames")
	createCmd.Flags().StringVar(&newCluster.Spec.DockerOpts.KubeImageProxy, "docker-kube-image-proxy", kaptain.DefaultKubeImageProxy, "Set this flag to use a proxy to download gcr.io images (e.g. gcr.io/google_containers/kube-apiserver)")
	createCmd.Flags().StringArrayVar(&newCluster.Spec.DockerOpts.InsecureRegistries, "docker-insecure-registry", []string{}, "Insecure Docker registries to allow")
	createCmd.Flags().StringArrayVar(&newCluster.Spec.DockerOpts.RegistryMirrors, "docker-registry-mirror", []string{}, "Docker registry mirror to add")
	createCmd.Flags().StringVar(&newCluster.Spec.MasterPublicName, "apiserver", "", "Kubernetes API server name (default api.<cluster_name>)")
	createCmd.Flags().IntVar(&newCluster.Spec.MasterPort, "apiserver-port", kaptain.DefaultMasterPort, "Kubernetes API server listen port")
	createCmd.Flags().StringVar(&newCluster.Spec.VSphereOpts.Username, "vsphere-username", "", "VSphere username")
	createCmd.Flags().StringVar(&newCluster.Spec.VSphereOpts.Password, "vsphere-password", "", "VSphere password")
	createCmd.Flags().StringVar(&newCluster.Spec.VSphereOpts.Server, "vsphere-server", "", "VSphere server")
	createCmd.Flags().StringVar(&newCluster.Spec.VSphereOpts.DataCenter, "vsphere-datacenter", "", "VSphere datacenter")
	createCmd.Flags().StringVar(&newCluster.Spec.VSphereOpts.DataStore, "vsphere-datastore", "", "VSphere datastore")
	createCmd.Flags().StringVar(&newCluster.Spec.VSphereOpts.WorkingDir, "vsphere-workingdir", "", "VSphere working directory")
	createCmd.Flags().StringVar(&authenticationTokenWebhookConfigFile, "authentication-token-webhook-config-file", "", "Kubernetes Authentication Webhook Config File, see https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication")
	createCmd.Flags().StringVar(&newCluster.Spec.AuthenticationTokenWebhookOpts.CacheTTL, "authentication-token-webhook-cache-ttl", "2m0s", "Kubernetes Authentication Webhook Cache TTL")
	createCmd.Flags().BoolVar(&newCluster.Spec.PodSecurityPolicyOpts.Enabled, "enable-pod-security-policy", false, "Enable PodSecurityPolicy, see 'cluster/pod-security-policy' for set up details")
}

func newEtcdCluster(etcdServers string) api.EtcdCluster {
	servers := strings.Split(etcdServers, ",")

	etcdCluster := api.EtcdCluster{}
	etcdCluster.Members = make([]api.EtcdMember, len(servers))

	for i, s := range servers {
		etcdCluster.Members[i] = api.EtcdMember{
			Hostname: s,
		}
	}

	return etcdCluster
}

func makeArrayFromCommaSeparatedString(commaSeparatedString string) []string {
	if commaSeparatedString == "" {
		return make([]string, 0)
	}
	return strings.Split(commaSeparatedString, ",")
}
