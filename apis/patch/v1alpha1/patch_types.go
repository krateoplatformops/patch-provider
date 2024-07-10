package v1alpha1

import (
	"github.com/imdario/mergo"
	prv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ObjectReference struct {
	// +optional
	ApiVersion *string `json:"apiVersion,omitempty"`

	Kind string `json:"kind"`

	// +optional
	Namespace *string `json:"namespace,omitempty"`

	Name string `json:"name"`
}

type FieldObjectReference struct {
	ObjectReference *ObjectReference `json:"objectReference,omitempty"`

	// +optional
	FieldPath *string `json:"fieldPath,omitempty"`
}

type FieldObjectReferenceWithTransforms struct {
	FieldObjectReference `json:",inline"`

	// +optional
	Transform string `json:"transform,omitempty"`
}

// PatchSpec defines the desired state of Patch
type PatchSpec struct {
	prv1.ManagedSpec `json:",inline"`

	// From is the path of the field on the resource whose value is
	// to be used as input.
	// +optional
	From *FieldObjectReference `json:"from,omitempty"`

	// To is the path of the field on the resource whose value will be changed.
	// +optional
	To *FieldObjectReferenceWithTransforms `json:"to,omitempty"`

	// MergeOptions specifies merge options on a field path.
	// +optional
	MergeOptions *MergeOptions `json:"mergeOptions,omitempty"`
}

// PatchStatus defines the observed state of Patch
type PatchStatus struct {
	prv1.ManagedStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster,categories={krateo}
//+kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
//+kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"

// Patch is the Schema for the patches API
type Patch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PatchSpec   `json:"spec,omitempty"`
	Status PatchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PatchList contains a list of Patch
type PatchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Patch `json:"items"`
}

// MergeOptions Specifies merge options on a field path
type MergeOptions struct {
	// Specifies that already existing values in a merged map should be preserved
	// +optional
	KeepMapValues *bool `json:"keepMapValues,omitempty"`
	// Specifies that already existing elements in a merged slice should be preserved
	// +optional
	AppendSlice *bool `json:"appendSlice,omitempty"`
}

// MergoConfiguration the default behavior is to replace maps and slices
func (mo *MergeOptions) MergoConfiguration() []func(*mergo.Config) {
	config := []func(*mergo.Config){mergo.WithOverride}
	if mo == nil {
		return config
	}

	if mo.KeepMapValues != nil && *mo.KeepMapValues {
		config = config[:0]
	}
	if mo.AppendSlice != nil && *mo.AppendSlice {
		config = append(config, mergo.WithAppendSlice)
	}
	return config
}

// IsAppendSlice returns true if mo.AppendSlice is set to true
func (mo *MergeOptions) IsAppendSlice() bool {
	return mo != nil && mo.AppendSlice != nil && *mo.AppendSlice
}
