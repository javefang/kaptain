package kaptain

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/utils/fileutil"
	"github.com/javefang/kaptain/pkg/utils/kubeutil"
)

func createFilesFromClusterSpec(role string, cluster *api.Cluster) (*api.ClusterFiles, error) {
	r := createRenderer(cluster)

	switch role {
	case "etcd":
		return createEtcdFiles(r)
	case "master":
		return createMasterFiles(r)
	case "worker":
		return createWorkerFiles(r)
	case "bootstrapper":
		return createBootstrapperFiles(r)
	default:
		return nil, fmt.Errorf("Invalid role: %s", role)
	}
}

func createEtcdFiles(r *renderer) (*api.ClusterFiles, error) {
	r.renderX509Cert("etcd-ca", EtcdCACert)
	r.renderX509Cert("etcd-server", EtcdServerCert)
	r.renderX509Key("etcd-server", EtcdServerKey)

	return r.clusterFiles, r.err
}

func createMasterFiles(r *renderer) (*api.ClusterFiles, error) {
	// common files
	createKubeNodeFiles(r)

	// master specific files
	r.renderNodeFile("sysconfig.kubelet.master", SysconfigKubeletKaptainExtra)

	// PKIs
	r.renderX509Cert("etcd-ca", KubeEtcdCA)
	r.renderX509Cert("etcd-client", KubeEtcdClientCert)
	r.renderX509Key("etcd-client", KubeEtcdClientKey)
	r.renderX509Cert("kube-ca", KubeCACert)
	r.renderX509Key("kube-ca", KubeCAKey)
	r.renderX509Cert("kubernetes", KubeCert)
	r.renderX509Key("kubernetes", KubeKey)

	// Tokens
	r.renderTokenCsv(KubeTokenCsv)

	// Kubeconfigs
	r.renderKubeConfig(makeKubeConfig(r.cluster, "kube-controller-manager"), KubeControllerManagerConfig)
	r.renderKubeConfig(makeKubeConfig(r.cluster, "kube-scheduler"), KubeSchedulerConfig)
	r.renderKubeConfig(makeKubeletMasterConfig(r.cluster), KubeletConfig)

	// Manifests
	r.renderNodeFile("manifest.kube-apiserver", KubeManifestApiserver)
	r.renderNodeFile("manifest.kube-controller-manager", KubeManifestControllerManager)
	r.renderNodeFile("manifest.kube-scheduler", KubeManifestScheduler)

	// Cloud provider specific config
	if r.cluster.Spec.CloudProvider == "vsphere" {
		r.renderNodeFile("config.cloud-config.vsphere", KubeCloudConfig)
	}

	// Auth Token Webhook specific config
	if r.cluster.Spec.AuthenticationTokenWebhookOpts.ConfigDataBase64 != "" {
		r.renderDataBase64(r.cluster.Spec.AuthenticationTokenWebhookOpts.ConfigDataBase64, AuthTokenWebhookConfig)
	}

	return r.clusterFiles, r.err
}

func createWorkerFiles(r *renderer) (*api.ClusterFiles, error) {
	// common files
	createKubeNodeFiles(r)

	// worker specific files
	r.renderNodeFile("sysconfig.kubelet.worker", SysconfigKubeletKaptainExtra)
	r.renderKubeConfig(makeKubeletBootstrapConfig(r.cluster), KubeletBootstrapConfig)

	return r.clusterFiles, r.err
}

func createBootstrapperFiles(r *renderer) (*api.ClusterFiles, error) {
	r.renderAddon("rbac-kube-system")
	r.renderAddon("rbac-node-bootstrap")
	r.renderAddon("calico")
	r.renderAddon("coredns")
	r.renderAddon("heapster")
	r.renderAddon("node-problem-detector")

	switch r.cluster.Spec.CloudProvider {
	case "aws":
		r.renderAddon("storageclass.aws")
	case "vsphere":
		r.renderAddon("storageclass.vsphere")
	}

	return r.clusterFiles, r.err
}

func createKubeNodeFiles(r *renderer) {
	r.renderNodeFile("config.docker-daemon", DockerDaemonConfig)
	r.renderNodeFile("sysconfig.docker", SysconfigDocker)
	r.renderNodeFile("sysconfig.kubelet", SysconfigKubeletKaptain)
	r.renderNodeFile("sysconfig.kube-proxy", SysconfigKubeProxyKaptain)

	r.renderKubeConfig(makeKubeConfig(r.cluster, "kube-proxy"), KubeProxyConfig)
}

// Util: create a ClusterFile
func createClusterFile(path string, data []byte) *api.ClusterFile {
	return &api.ClusterFile{
		Path:       path,
		DataBase64: base64.StdEncoding.EncodeToString(data),
	}
}

// Util: make token.csv file for the apiserver
func makeTokenCsv(tokenSecretMap map[string]api.TokenSecret) ([]byte, error) {
	rows := make([][]string, len(tokenSecretMap))
	i := 0
	for _, v := range tokenSecretMap {
		groupStr := strings.Join(v.Groups, ",")
		rows[i] = []string{v.Token, v.Username, strconv.Itoa(v.UID), groupStr}
		i++
	}
	return rowsToCSV(rows)
}

// Util: Make kubeconfig with x509 credentials for the specified username
func makeKubeConfig(c *api.Cluster, username string) *clientcmdapi.Config {
	clusterName := c.Name
	apiserverURL := fmt.Sprintf("https://%s", c.Spec.MasterPublicName)
	apiserverCAData := c.Secrets.PKIs["kube-ca"].GetCertData()
	certPair := c.Secrets.PKIs[username]

	cluster := clientcmdapi.NewCluster()
	cluster.Server = apiserverURL
	cluster.CertificateAuthorityData = apiserverCAData

	authInfo := clientcmdapi.NewAuthInfo()
	authInfo.ClientCertificateData = certPair.GetCertData()
	authInfo.ClientKeyData = certPair.GetKeyData()

	context := clientcmdapi.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = username

	config := clientcmdapi.NewConfig()
	config.Clusters[clusterName] = cluster
	config.AuthInfos[username] = authInfo
	config.Contexts["default"] = context

	config.CurrentContext = "default"

	return config
}

// Util: Make bootstrap kubeconfig for kubelet running on worker nodes
func makeKubeletBootstrapConfig(c *api.Cluster) *clientcmdapi.Config {
	username := "kubelet-bootstrap"
	clusterName := c.Name
	apiserverURL := fmt.Sprintf("https://%s", c.Spec.MasterPublicName)
	apiserverCAData := c.Secrets.PKIs["kube-ca"].GetCertData()
	tokenSecret := c.Secrets.TokenSecrets[username]

	cluster := clientcmdapi.NewCluster()
	cluster.Server = apiserverURL
	cluster.CertificateAuthorityData = apiserverCAData

	authInfo := clientcmdapi.NewAuthInfo()
	authInfo.Token = tokenSecret.Token

	context := clientcmdapi.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = username

	config := clientcmdapi.NewConfig()
	config.Clusters[clusterName] = cluster
	config.AuthInfos[username] = authInfo
	config.Contexts[defaultContextName] = context

	config.CurrentContext = defaultContextName

	return config
}

// Util: Make kubeconfig for kubelet on master
func makeKubeletMasterConfig(c *api.Cluster) *clientcmdapi.Config {
	username := "default"
	clusterName := c.Name
	apiserverURL := "http://127.0.0.1:8080"

	cluster := clientcmdapi.NewCluster()
	cluster.Server = apiserverURL

	authInfo := clientcmdapi.NewAuthInfo()

	context := clientcmdapi.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = username

	config := clientcmdapi.NewConfig()
	config.Clusters[clusterName] = cluster
	config.AuthInfos[username] = authInfo
	config.Contexts[defaultContextName] = context

	config.CurrentContext = defaultContextName

	return config
}

// Renderer

func createRenderer(cluster *api.Cluster) *renderer {
	return &renderer{
		cluster:      cluster,
		files:        indexByName(cluster.AssetManifest.Files),
		addons:       indexByName(cluster.AssetManifest.Addons),
		clusterFiles: api.NewClusterFiles(),
	}
}

func indexByName(files []api.NodeFile) map[string]api.NodeFile {
	index := map[string]api.NodeFile{}

	for _, f := range files {
		index[f.Name] = f
	}

	return index
}

type renderer struct {
	cluster *api.Cluster

	files  map[string]api.NodeFile
	addons map[string]api.NodeFile
	err    error

	clusterFiles *api.ClusterFiles
}

func (r *renderer) appendClusterFile(clusterFile *api.ClusterFile) {
	r.clusterFiles.Spec.ClusterFiles = append(r.clusterFiles.Spec.ClusterFiles, clusterFile)
}

func (r *renderer) renderNodeFile(templateName string, path string) {
	if r.err != nil {
		return
	}

	nodeFile, exist := r.files[templateName]
	if !exist {
		r.err = fmt.Errorf("NodeFile template not found: %s", templateName)
		return
	}
	templatePath := fmt.Sprintf("assets/files/%s/%s", nodeFile.Name, nodeFile.Version)
	data, err := fileutil.RenderTemplate(templatePath, r.cluster)
	if err != nil {
		r.err = err
		return
	}

	r.appendClusterFile(createClusterFile(path, data))
}

func (r *renderer) renderAddon(templateName string) {
	if r.err != nil {
		return
	}

	addon, exist := r.addons[templateName]
	if !exist {
		r.err = fmt.Errorf("Addon template not found: %s", templateName)
		return
	}
	templatePath := fmt.Sprintf("assets/addons/%s/%s.yaml", addon.Name, addon.Version)

	path := addon.Name
	data, err := fileutil.RenderTemplate(templatePath, r.cluster)
	if err != nil {
		r.err = err
		return
	}

	r.appendClusterFile(createClusterFile(path, data))
}

func (r *renderer) renderKubeConfig(config *clientcmdapi.Config, path string) {
	if r.err != nil {
		return
	}

	data, err := kubeutil.ValidateAndWriteToBuffer(config)
	if err != nil {
		r.err = err
		return
	}

	r.appendClusterFile(createClusterFile(path, data))
}

func (r *renderer) renderX509Cert(name string, path string) {
	if r.err != nil {
		return
	}

	data := r.cluster.Secrets.PKIs[name].GetCertData()
	r.appendClusterFile(createClusterFile(path, data))
}

func (r *renderer) renderX509Key(name string, path string) {
	if r.err != nil {
		return
	}

	data := r.cluster.Secrets.PKIs[name].GetKeyData()
	r.appendClusterFile(createClusterFile(path, data))
}

func (r *renderer) renderTokenCsv(path string) {
	if r.err != nil {
		return
	}

	data, err := makeTokenCsv(r.cluster.Secrets.TokenSecrets)
	if err != nil {
		r.err = err
		return
	}
	r.appendClusterFile(createClusterFile(path, data))
}

func (r *renderer) renderDataBase64(dataBase64 string, path string) {
	if r.err != nil {
		return
	}

	r.appendClusterFile(&api.ClusterFile{
		Path:       path,
		DataBase64: dataBase64,
	})
}

func rowsToCSV(rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
