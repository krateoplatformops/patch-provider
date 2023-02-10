package patching

import (
	"context"
	"strings"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errFmtExpandingArrayFieldPaths = "cannot expand To %s"
)

// Patch patches the "to" resource, using a source field
// on the "from" resource. Values may be transformed if any are defined on
// the patch.
func Patch(ctx context.Context, kc client.Client, cr *v1alpha1.Patch) (*unstructured.Unstructured, error) {
	if cr.Spec.From.FieldPath == nil {
		return nil, errors.New("from.fieldPath is required")
	}

	from, err := From(ctx, kc, cr)
	if err != nil {
		return nil, err
	}

	to, err := To(ctx, kc, cr)
	if err != nil {
		return nil, err
	}

	in, err := fieldpath.Pave(from.Object).GetValue(helpers.String(cr.Spec.From.FieldPath))
	if err != nil {
		return to, err
	}

	out, err := transformEventually(cr, in)
	if err != nil {
		return to, err
	}

	var mo *v1alpha1.MergeOptions
	if cr.Spec.MergeOptions != nil {
		mo = cr.Spec.MergeOptions
	}

	// Patch all expanded fields if the ToFieldPath contains wildcards
	toFieldPath := helpers.StringOrDefault(cr.Spec.To.FieldPath, helpers.String(cr.Spec.From.FieldPath))
	if strings.Contains(toFieldPath, "[*]") {
		return to, patchFieldValueToMultiple(toFieldPath, out, to, mo)
	}

	return to, fieldpath.Pave(to.Object).MergeValue(toFieldPath, out, mo)
}

// patchFieldValueToMultiple, given a path with wildcards in an array index,
// expands the arrays paths in the "to" object and patches the value into each
// of the resulting fields, returning any errors as they occur.
func patchFieldValueToMultiple(fieldPath string, value any, to *unstructured.Unstructured, mo *v1alpha1.MergeOptions) error {
	paved := fieldpath.Pave(to.UnstructuredContent())

	arrayFieldPaths, err := paved.ExpandWildcards(fieldPath)
	if err != nil {
		return err
	}

	if len(arrayFieldPaths) == 0 {
		return errors.Errorf(errFmtExpandingArrayFieldPaths, fieldPath)
	}

	for _, field := range arrayFieldPaths {
		if err := paved.MergeValue(field, value, mo); err != nil {
			return err
		}
	}

	return nil
}
