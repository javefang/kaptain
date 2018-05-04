package sailor

import (
	"fmt"
	"path"

	log "github.com/sirupsen/logrus"
	"github.com/javefang/kaptain/pkg/utils/fileutil"
)

func makeClusterFilesPath(clusterName string, role string) string {
	return path.Join(clusterName, "roles", fmt.Sprintf("%s.yaml", role))
}

func (c *SailorClient) Provision() error {
	logCtx := log.Fields{
		"cluster": c.ClusterName,
		"role":    c.Role,
	}

	log.WithFields(logCtx).Infof("SAILOR: provisioning node")

	log.WithFields(logCtx).Debug("SAILOR: fetching cluster files")
	clusterFiles, err := c.Registry.GetFiles(c.ClusterName, c.Role)
	if err != nil {
		return err
	}

	log.WithFields(logCtx).Infof("SAILOR: writing all files with prefix: %s", c.Prefix)
	return fileutil.WriteAll(c.Prefix, clusterFiles.Spec.ClusterFiles)
}
