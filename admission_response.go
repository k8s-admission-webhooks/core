package core

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// LabelName name label
	LabelName = "app.kubernetes.io/name"
	// LabelInstance instance label
	LabelInstance = "app.kubernetes.io/instance"
	// LabelVersion version label
	LabelVersion = "app.kubernetes.io/version"
	// LabelComponent component label
	LabelComponent = "app.kubernetes.io/component"
	// LabelPartOf part-of label
	LabelPartOf = "app.kubernetes.io/part-of"
	// LabelManagedBy managed-by label
	LabelManagedBy = "app.kubernetes.io/managed-by"
)

var (
	// IgnoredNamespaces list of namespaces that should ignored by admission review process
	IgnoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
)

// IsObjectInNamespaces check if an object is in specified set of namespaces
func IsObjectInNamespaces(meta *metav1.ObjectMeta, namespaces []string) bool {
	if namespaces == nil {
		namespaces = IgnoredNamespaces
	}
	for _, ns := range namespaces {
		if meta.Namespace == ns {
			return true
		}
	}
	return false
}

// PatchOperation a patch operation
type PatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// KeyValue create a map[string]string with a single entry
func KeyValue(key string, value string) map[string]string {
	return map[string]string{
		key: value,
	}
}

// NewAddPatch create a new `PatchOperation` for 'add' operation
func NewAddPatch(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Op:    "add",
		Path:  path,
		Value: value,
	}
}

// NewReplacePatch create a new `PatchOperation` for 'replace' operation
func NewReplacePatch(path string, value interface{}) PatchOperation {
	return PatchOperation{
		Op:    "replace",
		Path:  path,
		Value: value,
	}
}

func updateItems(current map[string]string, added map[string]string, path string) []PatchOperation {
	var patches []PatchOperation
	for key, value := range added {
		if current == nil || current[key] == "" {
			patches = append(patches, NewAddPatch(path, KeyValue(key, value)))
		} else {
			patches = append(patches, NewReplacePatch(path+key, value))
		}
	}
	return patches
}

// UpdateAnnotations create a set of packes to update annotations of an object
func UpdateAnnotations(current map[string]string, added map[string]string) []PatchOperation {
	return updateItems(current, added, "/metadata/annotations")
}

// UpdateLabels create a set of packes to update labels of an object
func UpdateLabels(current map[string]string, added map[string]string) []PatchOperation {
	return updateItems(current, added, "/metadata/labels")
}
