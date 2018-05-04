package api

import (
	"encoding/base64"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterFiles struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterFilesSpec `json:"spec"`
}

type ClusterFilesSpec struct {
	ClusterFiles []*ClusterFile `json:"files"`
}

func NewClusterFiles() *ClusterFiles {
	cf := ClusterFiles{}
	cf.Kind = "ClusterFiles"
	cf.APIVersion = "v1"
	cf.Labels = map[string]string{}
	cf.Spec.ClusterFiles = []*ClusterFile{}
	return &cf
}

// ClusterFile represents a file to be provisioned on a cluster node
type ClusterFile struct {
	Path       string `json:"path"`
	DataBase64 string `json:"data"`
}

func (cf *ClusterFile) GetData() ([]byte, error) {
	return base64.StdEncoding.DecodeString(cf.DataBase64)
}
