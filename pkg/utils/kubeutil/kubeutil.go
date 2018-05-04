package kubeutil

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"time"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/utils/fileutil"
)

// ValidateAndWriteToBuffer converts kubeconfig object to bytes
func ValidateAndWriteToBuffer(config *clientcmdapi.Config) ([]byte, error) {
	if err := clientcmd.Validate(*config); err != nil {
		return nil, fmt.Errorf("Failed validating kubeconfig: %v", err)
	}

	data, err := clientcmd.Write(*config)
	if err != nil {
		return nil, fmt.Errorf("failed to serialise kubeconfig: %v", err)
	}

	return data, nil
}

// KubeApply apply the file (in bytes) to the cluster
func KubeApply(context string, name string, data []byte) error {
	log.Debugf("Running kubectl apply for '%s' under context '%s' (len: %d)", name, context, len(data))

	reader := bytes.NewReader(data)
	cmd := exec.Command("kubectl", "--context", context, "apply", "-f", "-")
	cmd.Stdin = reader
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// KubeWaitForApiserver blocks and retries until apiserver is available
func KubeWaitForApiserver(context string) {
	log.Debugf("Running kubectl to check if apiserver is ready")

	for true {
		cmd := exec.Command("kubectl", "--context", context, "get", "nodes", "--no-headers")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Infof("Apiserver is not ready, waiting 10 seconds before retrying")
			time.Sleep(time.Second * 10)
		} else {
			log.Infof("Apiserver is ready, continuing...")
			break
		}
	}
}

// GetKubeConfig returns the kubeconfig for a specific user
func GetKubeConfig(c *api.Cluster, user string) (*clientcmdapi.Config, error) {
	// parse config from file if file already exist, otherwise, create empty config
	config := clientcmdapi.NewConfig()

	// prepare data
	clusterName := c.Name
	authInfoName := fmt.Sprintf("%s-%s", clusterName, user)
	apiserverURL := fmt.Sprintf("https://%s", c.Spec.MasterPublicName)
	apiserverCAData := c.Secrets.PKIs["kube-ca"].GetCertData()

	tokenSecret, exists := c.Secrets.TokenSecrets[user]
	if !exists {
		return nil, fmt.Errorf("user %s not found", user)
	}

	// set Cluster
	cluster := clientcmdapi.NewCluster()
	cluster.Server = apiserverURL
	cluster.CertificateAuthorityData = apiserverCAData
	config.Clusters[clusterName] = cluster

	// set AuthInfo
	authInfo := clientcmdapi.NewAuthInfo()
	authInfo.Token = tokenSecret.Token
	config.AuthInfos[authInfoName] = authInfo

	// set Context
	context := clientcmdapi.NewContext()
	context.Cluster = clusterName
	context.AuthInfo = authInfoName
	config.Contexts[clusterName] = context

	// set current context
	config.CurrentContext = clusterName

	return config, nil
}

// ExportKubeConfig exports the specified kubeconfig to a file on disk
func ExportKubeConfig(c *api.Cluster, filename string, user string, overwrite bool) error {
	// prepare new config
	newConfig, err := GetKubeConfig(c, user)
	if err != nil {
		return fmt.Errorf("failed to export kube config: %v", err)
	}

	// parse config from file if file already exist, otherwise, create empty config
	config, err := clientcmd.LoadFromFile(filename)
	if err != nil {
		log.Infof("creating new kubeconfig file '%s'", filename)
		config = clientcmdapi.NewConfig()
	}

	// prepare data
	clusterName := c.Name
	authInfoName := fmt.Sprintf("%s-%s", clusterName, user)

	// set Cluster
	if config.Clusters[clusterName] != nil && !overwrite {
		return fmt.Errorf("failed to set cluster '%s' in '%s': already exists", clusterName, filename)
	}
	config.Clusters[clusterName] = newConfig.Clusters[clusterName]

	// set AuthInfo
	if config.AuthInfos[authInfoName] != nil && !overwrite {
		return fmt.Errorf("failed to set authInfo '%s' in '%s': already exists", clusterName, filename)
	}
	config.AuthInfos[authInfoName] = newConfig.AuthInfos[authInfoName]

	// set Context
	if config.Contexts[clusterName] != nil && !overwrite {
		return fmt.Errorf("failed to set context '%s' in '%s': already exists", clusterName, filename)
	}
	config.Contexts[clusterName] = newConfig.Contexts[clusterName]

	// set current context
	config.CurrentContext = clusterName

	// validate and write
	data, err := ValidateAndWriteToBuffer(config)
	if err != nil {
		return err
	}

	if err := fileutil.EnsureDirExists(path.Dir(filename)); err != nil {
		return err
	}

	if err := fileutil.Write(data, filename); err != nil {
		return err
	}

	log.Infof("added cluster '%s' to kubeconfig '%s'", clusterName, filename)

	return nil
}
