package kaptain

import (
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/utils/fileutil"
	"github.com/javefang/kaptain/pkg/utils/pkiutil"
	"github.com/javefang/kaptain/pkg/utils/secretutil"
)

// InflateClusterOptions defines options to update the cluster specs
type InflateClusterOptions struct {
	UpdateSpec          bool
	UpdatePKIs          bool
	UpdateTokens        bool
	UpdateAssetManifest bool
}

// InflateCluster inflate the cluster by setting default values, prepare role manifest, generate PKIs and token secrets
func InflateCluster(cluster *api.Cluster, opts *InflateClusterOptions) error {
	// Initialise random seed, otherwise all secrets will be generated with the same value
	rand.Seed(time.Now().UnixNano())

	b, _ := yaml.Marshal(cluster)
	log.Infof(string(b))

	if opts.UpdateSpec {
		inflateClusterDefaults(cluster)
	}

	if opts.UpdateAssetManifest {
		err := inflateAssetManifest(cluster)
		if err != nil {
			return err
		}
	}

	if opts.UpdatePKIs {
		inflatePKIs(cluster)
	}

	if opts.UpdateTokens {
		inflateTokens(cluster)
	}

	return nil
}

func inflateClusterDefaults(cluster *api.Cluster) {
	log.Infof("Inflating cluster spec")

	// set Pod CIDR
	if cluster.Spec.PodCIDR == "" {
		cluster.Spec.PodCIDR = DefaultPodCIDR
	}

	if cluster.Spec.ServiceCIDR == "" {
		cluster.Spec.ServiceCIDR = DefaultServiceCIDR
	}

	if cluster.Spec.DNSClusterIP == "" {
		cluster.Spec.DNSClusterIP = DefaultDNSClusterIP
	}
}

func inflateAssetManifest(cluster *api.Cluster) error {
	log.Infof("Inflating asset manifest")
	majorMinorVersion, err := getMajorMinorVersion(cluster.Spec.KubeVersion)
	if err != nil {
		return err
	}
	manifest, err := getManifest(majorMinorVersion)
	if err != nil {
		log.Errorf("Error reading manifest: %v", err)
		return err
	}
	cluster.AssetManifest = manifest.Spec

	return nil
}

func inflatePKIs(cluster *api.Cluster) {
	log.Infof("Inflating PKIs")

	// etcd-ca
	var etcdCA *pkiutil.CertCombo
	if etcdCAPair, exists := cluster.Secrets.PKIs["etcd-ca"]; !exists {
		log.Infof("Creating new CA: etcd-ca")
		// create new CA
		etcdCACsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "ETCD CA",
			},
			ValidFor: defaultCAExpiry,
			Profile:  pkiutil.None,
		}
		etcdCA = pkiutil.InitCA(etcdCACsr)
		cluster.Secrets.PKIs["etcd-ca"] = makeCertPair(etcdCA)
	} else {
		// use existing CA
		log.Infof("Use existing PKI: etcd-ca")
		etcdCA = makeCertCombo(&etcdCAPair)
	}

	// kube-ca
	var kubeCA *pkiutil.CertCombo
	if kubeCAPair, exists := cluster.Secrets.PKIs["kube-ca"]; !exists {
		log.Infof("Creating new CA: kube-ca")
		// create new CA
		kubeCACsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "Kube CA",
			},
			ValidFor: defaultCAExpiry,
			Profile:  pkiutil.None,
		}
		kubeCA = pkiutil.InitCA(kubeCACsr)
		cluster.Secrets.PKIs["kube-ca"] = makeCertPair(kubeCA)
	} else {
		// use existing CA
		log.Infof("Use existing PKI: kube-ca")
		kubeCA = makeCertCombo(&kubeCAPair)
	}

	// etcd-server
	if _, exists := cluster.Secrets.PKIs["etcd-server"]; !exists {
		etcdMemberCount := len(cluster.Spec.EtcdCluster.Members)
		etcdAltNames := make([]string, etcdMemberCount*2)
		for i, v := range cluster.Spec.EtcdCluster.Members {
			fullHostname := fmt.Sprintf("%s.%s", v.Hostname, cluster.Spec.DNSDomain)
			etcdAltNames[i] = v.Hostname
			etcdAltNames[i+etcdMemberCount] = fullHostname
		}
		etcdCsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "etcd",
			},
			AltNames: etcdAltNames,
			Profile:  pkiutil.Server,
			ValidFor: defaultCertExpiry,
		}
		etcd := pkiutil.MakeCert(etcdCsr, etcdCA)
		cluster.Secrets.PKIs["etcd-server"] = makeCertPair(etcd)
	}

	// ** etcd-client (client)
	if _, exists := cluster.Secrets.PKIs["etcd-client"]; !exists {
		etcdClientCsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "apiserver",
			},
			Profile:  pkiutil.Client,
			ValidFor: defaultCertExpiry,
		}
		etcdClient := pkiutil.MakeCert(etcdClientCsr, etcdCA)
		cluster.Secrets.PKIs["etcd-client"] = makeCertPair(etcdClient)
	}

	// kubernetes
	if _, exists := cluster.Secrets.PKIs["kubernetes"]; !exists {
		kubeCsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "kubernetes",
			},
			AltNames: []string{
				cluster.Spec.MasterPublicName,
				"kubernetes",
				"kubernetes.default",
				"kubernetes.default.svc",
				"kubernetes.default.svc.cluster",
				"kubernetes.default.svc.cluster.local",
				"localhost",
				"127.0.0.1",
				DefaultMasterServiceIP,
			},
			Profile:  pkiutil.Server,
			ValidFor: defaultCertExpiry,
		}
		kube := pkiutil.MakeCert(kubeCsr, kubeCA)
		cluster.Secrets.PKIs["kubernetes"] = makeCertPair(kube)
	}

	// kube-controller-manager
	if _, exists := cluster.Secrets.PKIs["kube-controller-manager"]; !exists {
		controllerManagerCsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "system:kube-controller-manager",
			},
			Profile:  pkiutil.Client,
			ValidFor: defaultCertExpiry,
		}
		controllerManager := pkiutil.MakeCert(controllerManagerCsr, kubeCA)
		cluster.Secrets.PKIs["kube-controller-manager"] = makeCertPair(controllerManager)
	}

	// kube-scheduler
	if _, exists := cluster.Secrets.PKIs["kube-scheduler"]; !exists {
		schedulerCsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "system:kube-scheduler",
			},
			Profile:  pkiutil.Client,
			ValidFor: defaultCertExpiry,
		}
		scheduler := pkiutil.MakeCert(schedulerCsr, kubeCA)
		cluster.Secrets.PKIs["kube-scheduler"] = makeCertPair(scheduler)
	}

	// kube-proxy (client)
	if _, exists := cluster.Secrets.PKIs["kube-proxy"]; !exists {
		kubeProxyCsr := pkiutil.CSRParams{
			Subject: pkix.Name{
				CommonName: "system:kube-proxy",
			},
			Profile:  pkiutil.Client,
			ValidFor: defaultCertExpiry,
		}
		kubeProxy := pkiutil.MakeCert(kubeProxyCsr, kubeCA)
		cluster.Secrets.PKIs["kube-proxy"] = makeCertPair(kubeProxy)
	}
}

func inflateTokens(cluster *api.Cluster) {
	log.Infof("Inflating tokens")

	// kubelet-bootstrap
	if _, exists := cluster.Secrets.TokenSecrets["kubelet-bootstrap"]; !exists {
		cluster.Secrets.TokenSecrets["kubelet-bootstrap"] = api.TokenSecret{
			Username: "kubelet-bootstrap",
			Token:    secretutil.MakeRandomToken(defaultTokenLength),
			UID:      10001,
			Groups:   []string{"system:bootstrappers"},
		}
	}

	// admin
	if _, exists := cluster.Secrets.TokenSecrets["admin"]; !exists {
		cluster.Secrets.TokenSecrets["admin"] = api.TokenSecret{
			Username: "admin",
			Token:    secretutil.MakeRandomToken(defaultTokenLength),
			UID:      1,
			Groups:   []string{"system:masters"},
		}
	}
}

func makeCertPair(certCombo *pkiutil.CertCombo) api.CertPair {
	certData := base64.StdEncoding.EncodeToString(certCombo.ExtractCertData())
	keyData := base64.StdEncoding.EncodeToString(certCombo.ExtractKeyData())

	return api.CertPair{
		CertData: certData,
		KeyData:  keyData,
	}
}

func makeCertCombo(certPair *api.CertPair) *pkiutil.CertCombo {
	certCombo := pkiutil.CertCombo{}
	certCombo.SetCertPEMData(certPair.GetCertData())
	certCombo.SetKeyPEMData(certPair.GetKeyData())
	return &certCombo
}

func getManifest(majorMinorVerion string) (*api.AssetManifest, error) {
	data, err := fileutil.GetAsset(fmt.Sprintf("assets/manifests/%s.yaml", majorMinorVerion))
	if err != nil {
		return nil, fmt.Errorf("Failed to read manifest for version %s: %v", majorMinorVerion, err)
	}

	var manifest api.AssetManifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse manifest for version %s: %v", majorMinorVerion, err)
	}

	return &manifest, nil
}

// Extract the major and minor version from the version string.
// For example: getMajorMinorVersion("v1.2.3") will return "1.2"
func getMajorMinorVersion(version string) (string, error) {
	if !strings.HasPrefix(version, "v") {
		return "", fmt.Errorf("Invalid Kubernetes version %s: it must starts with 'v'", version)
	}

	v := strings.TrimPrefix(version, "v")
	vs := strings.Split(v, ".")

	if len(vs) != 3 {
		return "", fmt.Errorf("Invalid Kubernetes version %s: it must be of form vX.Y.Z", version)
	}

	return strings.Join(vs[0:2], "."), nil
}
