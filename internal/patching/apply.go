package patching

import (
	"strings"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	errFmtRequiredField            = "%s is required by type %T"
	errFmtExpandingArrayFieldPaths = "cannot expand To %s"
)

// ApplyFromFieldPathPatch patches the "to" resource, using a source field
// on the "from" resource. Values may be transformed if any are defined on
// the patch.
func Apply(p v1alpha1.Patch, from, to *unstructured.Unstructured) error {
	if p.Spec.From == nil {
		return errors.Errorf(errFmtRequiredField, "from", p)
	}

	if p.Spec.From.FieldPath == nil {
		return errors.Errorf(errFmtRequiredField, "from.fieldPath", p)
	}

	// Default to patching the same field.
	if p.Spec.To == nil {
		p.Spec.To = p.Spec.From
	}

	if p.Spec.To.FieldPath == nil {
		p.Spec.To.FieldPath = p.Spec.From.FieldPath
	}

	in, err := fieldpath.Pave(from.Object).GetValue(*p.Spec.From.FieldPath)
	if err != nil {
		return err
	}

	var mo *v1alpha1.MergeOptions
	if p.Spec.MergeOptions != nil {
		mo = p.Spec.MergeOptions
	}

	// Apply transform pipeline
	out, err := ResolveTransforms(p, in)
	if err != nil {
		return err
	}

	// Patch all expanded fields if the ToFieldPath contains wildcards
	if strings.Contains(*p.Spec.To.FieldPath, "[*]") {
		return patchFieldValueToMultiple(*p.Spec.To.FieldPath, out, to, mo)
	}

	return fieldpath.Pave(to.Object).MergeValue(*p.Spec.To.FieldPath, out, mo)
}

// TODO: ResolveTransforms applies a list of transforms to a patch value.
func ResolveTransforms(c v1alpha1.Patch, input any) (any, error) {
	//var err error
	//for i, t := range c.Transforms {
	//	if input, err = Resolve(t, input); err != nil {
	//		// TODO(negz): Including the type might help find the offending transform faster.
	//		return nil, errors.Wrapf(err, errFmtTransformAtIndex, i)
	//	}
	//}
	return input, nil
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
