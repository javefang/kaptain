package kaptain

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/javefang/kaptain/pkg/api"
	"github.com/javefang/kaptain/pkg/utils/kubeutil"
)

// Bootstrap initialise the cluster
func bootstrap(clusterName string, addonFiles *api.ClusterFiles) error {
	// waiting for apiserver to get ready
	kubeutil.KubeWaitForApiserver(clusterName)

	// install addons
	for _, file := range addonFiles.Spec.ClusterFiles {
		data, err := file.GetData()
		if err != nil {
			return fmt.Errorf("failed to deserialise data for %s: %v", file.Path, err)
		}

		err = kubeutil.KubeApply(clusterName, file.Path, data)
		if err != nil {
			return fmt.Errorf("failed to apply kube addon %s: %v", file.Path, err)
		}
	}

	log.Info("Cluster bootstrapped! The containers might take a few minutes to starts.")

	return nil
}
