package api

import (
	"fmt"
	"path"
	"time"

	"github.com/ghodss/yaml"
	log "github.com/sirupsen/logrus"
	"github.com/javefang/kaptain/pkg/store"
)

const clusterSpecFile = "cluster.yaml"
const defaultCAExpiry = time.Hour * 24 * 365 * 5
const defaultCertExpiry = time.Hour * 24 * 365
const defaultTokenLength = 32

type ClusterRegistry struct {
	store store.Store
}

func NewClusterRegistry(storeUrl string) *ClusterRegistry {
	s := store.CreateStoreFromUrlOrDie(storeUrl)

	return &ClusterRegistry{
		store: s,
	}
}

func makeClusterSpecPath(clusterName string) string {
	return path.Join(clusterName, clusterSpecFile)
}

func makeClusterFilesPath(clusterName string, role string) string {
	return path.Join(clusterName, "roles", fmt.Sprintf("%s.yaml", role))
}

func (reg *ClusterRegistry) List() ([]string, error) {
	log.Debug("Listing clusters")
	clusterNames, err := reg.store.List("")
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}
	return clusterNames, nil
}

func (reg *ClusterRegistry) Exists(clusterName string) (bool, error) {
	log.Debugf("Checking cluster existency for '%s'", clusterName)
	exists, err := reg.store.Exists(makeClusterSpecPath(clusterName))
	if err != nil {
		return false, fmt.Errorf("failed to check cluster '%s': %v", clusterName, err)
	}
	return exists, nil
}

func (reg *ClusterRegistry) Get(clusterName string) (*Cluster, error) {
	log.Debugf("Getting cluster details for '%s'", clusterName)
	data, err := reg.store.Get(makeClusterSpecPath(clusterName))
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster '%s': %v", clusterName, err)
	}

	cluster := Cluster{}
	err = yaml.Unmarshal(data, &cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cluster spec '%s': %v", clusterName, err)
	}

	return &cluster, nil
}

func (reg *ClusterRegistry) Create(cluster *Cluster, force bool) error {
	clusterName := cluster.Name
	if clusterName == "" {
		return fmt.Errorf("cluster name cannot be empty")
	}

	exists, err := reg.Exists(clusterName)
	if err != nil {
		return fmt.Errorf("failed to check if cluster '%s' exists: %v", clusterName, err)
	}

	if exists && !force {
		return fmt.Errorf("failed to create cluster '%s': cluster with the same name already exists (use -f to override)", clusterName)
	}

	log.Debugf("Creating new cluster '%s'", cluster.Name)

	// prepare and write cluster spec
	data, err := yaml.Marshal(cluster)
	if err != nil {
		panic(err)
	}

	log.Debugf("Cluster: \n%s", string(data))

	err = reg.store.Set(makeClusterSpecPath(cluster.Name), data)
	if err != nil {
		return fmt.Errorf("failed to create cluster '%s': %v", cluster.Name, err)
	}

	return nil
}

func (reg *ClusterRegistry) Delete(clusterName string) error {
	log.Debugf("Deleting cluster '%s'", clusterName)

	if err := reg.store.DeleteAll(clusterName); err != nil {
		return fmt.Errorf("failed to cleanup storage when deleting cluster '%s': %v", clusterName, err)
	}

	return nil
}

func (reg *ClusterRegistry) GetFiles(clusterName string, role string) (*ClusterFiles, error) {
	log.Debugf("Get cluster files for '%s' as role '%s'", clusterName, role)
	data, err := reg.store.Get(makeClusterFilesPath(clusterName, role))
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster file for cluster '%s' and role '%s': %v", clusterName, role, err)
	}

	var clusterFiles ClusterFiles
	err = yaml.Unmarshal(data, &clusterFiles)

	return &clusterFiles, err
}

func (reg *ClusterRegistry) SetFiles(clusterName string, role string, clusterFiles *ClusterFiles) error {
	log.Debugf("Set cluster files for '%s' as role '%s'", clusterName, role)

	data, err := yaml.Marshal(clusterFiles)
	if err != nil {
		panic(err)
	}

	err = reg.store.Set(makeClusterFilesPath(clusterName, role), data)
	if err != nil {
		return fmt.Errorf("failed to write cluster files for %s: %v", role, err)
	}

	return nil
}
