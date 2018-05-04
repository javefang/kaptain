package api

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AssetManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AssetManifestSpec `json:"spec"`
}

type AssetManifestSpec struct {
	Addons []NodeFile `json:"addons"`
	Files  []NodeFile `json:"files"`
}
