package kaptain

import "time"

const (
	// role=etcd
	EtcdCACert     = "etc/pki/tls/certs/etcd-ca.pem"
	EtcdServerCert = "etc/pki/tls/certs/etcd-server.pem"
	EtcdServerKey  = "etc/pki/tls/private/etcd-server-key.pem"

	// role=master|worker
	DockerDaemonConfig        = "etc/docker/daemon.json"
	SysconfigDocker           = "etc/sysconfig/docker"
	SysconfigKubeletKaptain   = "etc/sysconfig/kubelet-kaptain"
	SysconfigKubeProxyKaptain = "etc/sysconfig/kube-proxy-kaptain"
	KubeProxyConfig           = "var/lib/kube-proxy/kubeconfig"

	// role=master
	KubeEtcdCA                  = "var/lib/kubernetes/etcd-ca.pem"
	KubeEtcdClientCert          = "var/lib/kubernetes/etcd-client.pem"
	KubeEtcdClientKey           = "var/lib/kubernetes/etcd-client-key.pem"
	KubeCACert                  = "var/lib/kubernetes/ca.pem"
	KubeCAKey                   = "var/lib/kubernetes/ca-key.pem"
	KubeCert                    = "var/lib/kubernetes/kubernetes.pem"
	KubeKey                     = "var/lib/kubernetes/kubernetes-key.pem"
	KubeTokenCsv                = "var/lib/kubernetes/token.csv"
	KubeBasicAuthCsv            = "var/lib/kubernetes/basic_auth.csv"
	KubeletConfig               = "var/lib/kubelet/kubeconfig"
	KubeControllerManagerConfig = "var/lib/kubernetes/kube-controller-manager.kubeconfig"
	KubeSchedulerConfig         = "var/lib/kubernetes/kube-scheduler.kubeconfig"
	KubeCloudConfig             = "var/lib/kubernetes/cloud.conf"
	AuthTokenWebhookConfig      = "var/lib/kubernetes/authn-webhook-config"

	KubeManifestApiserver         = "etc/kubernetes/manifests/kube-apiserver.yaml"
	KubeManifestControllerManager = "etc/kubernetes/manifests/kube-controller-manager.yaml"
	KubeManifestScheduler         = "etc/kubernetes/manifests/kube-scheduler.yaml"

	// role=worker
	SysconfigKubeletKaptainExtra = "etc/sysconfig/kubelet-kaptain-extra"
	KubeletBootstrapConfig       = "var/lib/kubelet/bootstrap.kubeconfig"
)

// default values
const DefaultMasterServiceIP = "100.64.0.1"
const DefaultDNSClusterIP = "100.64.0.10"
const DefaultServiceCIDR = "100.64.0.0/16"
const DefaultPodCIDR = "100.200.0.0/16"
const DefaultEtcdMemberCount = 3
const DefaultKubeVersion = "v1.10.1"
const DefaultCloudProvider = "aws"
const DefaultKubeImageProxy = "gcr.io"
const DefaultMasterPort = 6443
const DefaultClusterDomain = "cluster.local"

const clusterSpecFile = "cluster.yaml"
const defaultCAExpiry = time.Hour * 24 * 365 * 5
const defaultCertExpiry = time.Hour * 24 * 365
const defaultTokenLength = 32
const defaultContextName = "default"
