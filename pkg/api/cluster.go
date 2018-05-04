package api

import (
	"encoding/base64"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/javefang/kaptain/pkg/version"
)

const apiNamespace = "k8s.io/kaptain"

// Cluster is the full representation of a Kubernetes cluster that can be used to reproduce a cluster set up.
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec          ClusterSpec       `json:"spec"`
	AssetManifest AssetManifestSpec `json:"assetManifest"`
	Secrets       ClusterSecrets    `json:"secrets"`
}

// ClusterSpec is the spec that defines a the property of a cluster
type ClusterSpec struct {
	KubeVersion                    string                         `json:"kubeVersion"` // Version of Kubernetes
	MasterPublicName               string                         `json:"masterPublicName"`
	MasterPort                     int                            `json:"masterPort"`
	DNSDomain                      string                         `json:"dnsDomain"`
	PodCIDR                        string                         `json:"podCIDR"`
	ServiceCIDR                    string                         `json:"serviceCIDR"`
	DNSClusterIP                   string                         `json:"dnsClusterIP"`
	CloudProvider                  string                         `json:"cloudProvider"`
	CloudConfig                    string                         `json:"cloudConfig"`
	WorkerCloudConfig              string                         `json:"workerCloudConfig"`
	DockerOpts                     DockerOpts                     `json:"dockerOpts"`
	VSphereOpts                    VSphereOpts                    `json:"vsphereOpts"`
	PodSecurityPolicyOpts          PodSecurityPolicyOpts          `json:"podSecurityPolicyOpts"`
	AuthenticationTokenWebhookOpts AuthenticationTokenWebhookOpts `json:"authenticationTokenWebhookOpts"`
	EtcdCluster                    EtcdCluster                    `json:"etcdCluster"`
}

// ClusterSecrets stores PKI and token secrets used to secure the cluster
type ClusterSecrets struct {
	PKIs         map[string]CertPair    `json:"pkis"`
	TokenSecrets map[string]TokenSecret `json:"tokenSecrets"`
}

// DockerOpts is the configurable options for Docker
type DockerOpts struct {
	InsecureRegistries []string `json:"insecureRegistries"`
	RegistryMirrors    []string `json:"registryMirrors"`
	KubeImageProxy     string   `json:"kubeImageProxy"`
}

// VSphereOpts is the configuration options for when VSphere is used as the cloud provider
type VSphereOpts struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Server     string `json:"server"`
	DataCenter string `json:"dataCenter"`
	DataStore  string `json:"dataStore"`
	WorkingDir string `json:"workingDir"`
}

// AuthenticationTokenWebhookOpts is the configurable options for when authentication token webhook is used.
type AuthenticationTokenWebhookOpts struct {
	ConfigDataBase64 string `json:"configDataBase64"` // base64 encoded string of the config file
	CacheTTL         string `json:"cacheTTL"`
}

// PodSecurityPolicyOpts is the configurable options for PodSecurityPolicy
type PodSecurityPolicyOpts struct {
	Enabled bool `json:"enabled"`
}

// EtcdCluster contains information about the ETCD cluster used by Kubernetes
type EtcdCluster struct {
	Members []EtcdMember `json:"members"`
}

// EtcdMember contains information about a single ETCD node of the ETCD cluster used by Kubernetes
type EtcdMember struct {
	Hostname string `json:"hostname"`
}

// CertPair is contains base64 encoded cert and key data
type CertPair struct {
	CertData string `json:"certData"`
	KeyData  string `json:"keyData"`
}

// TokenSecret contains information about a bearing token used for apiserver authentication
type TokenSecret struct {
	Username string   `json:"username"`
	Token    string   `json:"token"`
	UID      int      `json:"uid"`
	Groups   []string `json:"groups"`
}

// NewCluster creates a new Cluster object
func NewCluster() Cluster {
	cluster := Cluster{}
	cluster.Kind = "Cluster"
	cluster.APIVersion = "v1"
	cluster.Annotations = map[string]string{}

	v := version.GetVersion()
	cluster.Annotations[getAnnotationFullName("version")] = v.Version
	cluster.Annotations[getAnnotationFullName("git-commit")] = v.GitCommit
	cluster.Annotations[getAnnotationFullName("git-tree-state")] = v.GitTreeState

	cluster.Secrets.PKIs = make(map[string]CertPair)
	cluster.Secrets.TokenSecrets = make(map[string]TokenSecret)

	return cluster
}

func getAnnotationFullName(field string) string {
	return fmt.Sprintf("%s/%s", apiNamespace, field)
}

// NodeFile represents a specific version of the node config file template from asset files
type NodeFile struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (a NodeFile) String() string {
	return fmt.Sprintf("%s@%s", a.Name, a.Version)
}

// GetCertData returns the x509 certificate data in bytes
func (pair CertPair) GetCertData() []byte {
	data, err := base64.StdEncoding.DecodeString(pair.CertData)
	if err != nil {
		panic(fmt.Errorf("Failed to decode x509 cert data: %v", err))
	}
	return data
}

// GetKeyData returns the x509 key data in bytes
func (pair CertPair) GetKeyData() []byte {
	data, err := base64.StdEncoding.DecodeString(pair.KeyData)
	if err != nil {
		panic(fmt.Errorf("Failed to decode x509 key data: %v", err))
	}
	return data
}
