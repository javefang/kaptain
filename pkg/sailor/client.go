package sailor

import (
	"github.com/javefang/kaptain/pkg/api"
)

type SailorClient struct {
	Role        string
	ClusterName string
	Prefix      string
	Registry    *api.ClusterRegistry
}
