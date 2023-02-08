package v1alpha1

import (
	prv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
)

// GetCondition of this Patch.
func (mg *Patch) GetCondition(ct prv1.ConditionType) prv1.Condition {
	return mg.Status.GetCondition(ct)
}

// GetDeletionPolicy of this Patch.
func (mg *Patch) GetDeletionPolicy() prv1.DeletionPolicy {
	return mg.Spec.DeletionPolicy
}

// SetConditions of this Patch.
func (mg *Patch) SetConditions(c ...prv1.Condition) {
	mg.Status.SetConditions(c...)
}

// SetDeletionPolicy of this Patch.
func (mg *Patch) SetDeletionPolicy(r prv1.DeletionPolicy) {
	mg.Spec.DeletionPolicy = r
}
