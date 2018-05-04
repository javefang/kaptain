package kaptain

import (
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/utils/kubeutil"
)

type KaptainClient struct {
	Registry *api.ClusterRegistry
}

func (client *KaptainClient) List() error {
	clusterNames, err := client.Registry.List()
	if err != nil {
		return fmt.Errorf("failed to list clusters: %v", err)
	}

	printClusterNames(clusterNames)

	return nil
}

func (client *KaptainClient) Create(cluster *api.Cluster, force bool) error {
	// TODO: check if cluster exists

	// write cluster spec
	if err := client.Registry.Create(cluster, force); err != nil {
		return err
	}

	// write cluster files
	for _, role := range []string{"etcd", "master", "worker", "bootstrapper"} {
		clusterFiles, err := createFilesFromClusterSpec(role, cluster)
		if err != nil {
			return fmt.Errorf("Failed to render cluster files for %s: %v", role, err)
		}

		err = client.Registry.SetFiles(cluster.Name, role, clusterFiles)
		if err != nil {
			return fmt.Errorf("failed to write cluster files for %s: %v", role, err)
		}
	}

	return nil
}

// TODO: create func (client *KaptainClient) Apply(cluster *api.Cluster) error

func (client *KaptainClient) Delete(clusterName string) error {
	// TODO: check if cluster exists

	if err := client.Registry.Delete(clusterName); err != nil {
		return err
	}

	// TODO: add confirmation

	return nil
}

func (client *KaptainClient) Get(clusterName string) (*api.Cluster, error) {
	cluster, err := client.Registry.Get(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to read cluster: %v", err)
	}

	return cluster, nil
}

func (client *KaptainClient) ExportConfig(clusterName string, kubeConfigFilePath string, username string, overwrite bool) error {
	// TODO: check if cluster exists

	cluster, err := client.Registry.Get(clusterName)
	if err != nil {
		return fmt.Errorf("failed to read cluster: %v", err)
	}

	err = kubeutil.ExportKubeConfig(cluster, kubeConfigFilePath, username, overwrite)
	if err != nil {
		return fmt.Errorf("failed to export cluster config: %v", err)
	}

	return nil
}

func (client *KaptainClient) Bootstrap(clusterName string) error {
	// TODO: check if cluster exists

	addonFiles, err := client.Registry.GetFiles(clusterName, "bootstrapper")
	if err != nil {
		return fmt.Errorf("failed to get addon files for %s: %v", clusterName, err)
	}

	return bootstrap(clusterName, addonFiles)
}

func printClusterNames(clusters []string) {
	data := make([][]string, len(clusters))
	for i, c := range clusters {
		data[i] = []string{c}
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}
