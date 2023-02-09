package patching

import (
	"context"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func Diff(cr *v1alpha1.Patch, from, to *unstructured.Unstructured) (string, error) {
	in, err := fieldpath.Pave(from.Object).GetValue(helpers.String(cr.Spec.From.FieldPath))
	if err != nil {
		return "", err
	}

	// Default to patching the same field.
	if cr.Spec.To.FieldPath == nil {
		cr.Spec.To.FieldPath = cr.Spec.From.FieldPath
	}

	out, err := fieldpath.Pave(to.Object).GetValue(helpers.String(cr.Spec.To.FieldPath))
	if err != nil {
		return "", err
	}

	return cmp.Diff(in, out), nil
}

func FromObjectReference(ctx context.Context, kc client.Client, cr *v1alpha1.Patch) (*unstructured.Unstructured, error) {
	if cr.Spec.From == nil || cr.Spec.From.ObjectReference == nil {
		return nil, errors.New("missing 'from' objectReference")
	}

	return resolveObjectReference(ctx, kc, cr.Spec.From.ObjectReference)
}

func ToObjectReference(ctx context.Context, kc client.Client, cr *v1alpha1.Patch) (*unstructured.Unstructured, error) {
	if cr.Spec.To == nil || cr.Spec.To.ObjectReference == nil {
		return resolveObjectReference(ctx, kc, cr.Spec.From.ObjectReference)
	}

	return resolveObjectReference(ctx, kc, cr.Spec.To.ObjectReference)
}

func resolveObjectReference(ctx context.Context, kc client.Client, ref *v1alpha1.ObjectReference) (*unstructured.Unstructured, error) {
	if ref == nil {
		return nil, nil
	}

	gvk, err := schema.ParseGroupVersion(helpers.StringOrDefault(ref.ApiVersion, "v1"))
	if err != nil {
		return nil, errors.Wrapf(err, "parsing object reference 'apiVersion' field")
	}

	res := &unstructured.Unstructured{}
	res.SetGroupVersionKind(gvk.WithKind(ref.Kind))
	err = kc.Get(ctx, types.NamespacedName{
		Name:      ref.Name,
		Namespace: helpers.StringOrDefault(ref.Namespace, "default"),
	}, res)

	return res, err
}
