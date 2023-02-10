package patching

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"github.com/krateoplatformops/patch-provider/internal/functions"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	errFmtRequiredField            = "%s is required by type %s"
	errFmtExpandingArrayFieldPaths = "cannot expand To %s"
)

// Patch patches the "to" resource, using a source field
// on the "from" resource. Values may be transformed if any are defined on
// the patch.
func Patch(cr *v1alpha1.Patch, from, to *unstructured.Unstructured) error {
	if cr.Spec.From == nil {
		return errors.Errorf(errFmtRequiredField, "from.objectReference", "spec")
	}

	if cr.Spec.From.FieldPath == nil {
		return errors.Errorf(errFmtRequiredField, "from.fieldPath", "spec")
	}

	// Default to patching the same field.
	if cr.Spec.To.FieldPath == nil {
		cr.Spec.To.FieldPath = cr.Spec.From.FieldPath
	}

	fromFieldPath := helpers.String(cr.Spec.From.FieldPath)
	in, err := fieldpath.Pave(from.Object).GetValue(fromFieldPath)
	if err != nil {
		return err
	}

	out, err := transformEventually(cr, in)
	if err != nil {
		return err
	}

	var mo *v1alpha1.MergeOptions
	if cr.Spec.MergeOptions != nil {
		mo = cr.Spec.MergeOptions
	}

	// Patch all expanded fields if the ToFieldPath contains wildcards
	toFieldPath := helpers.String(cr.Spec.To.FieldPath)
	if strings.Contains(toFieldPath, "[*]") {
		return patchFieldValueToMultiple(toFieldPath, out, to, mo)
	}

	return fieldpath.Pave(to.Object).MergeValue(toFieldPath, out, mo)
}

func transformEventually(cr *v1alpha1.Patch, input any) (any, error) {
	fn := helpers.String(cr.Spec.To.Transform)
	if len(fn) == 0 {
		return input, nil
	}

	buf := bytes.NewBufferString("")
	tpl := template.New(cr.GetName()).Funcs(functions.Map())
	tpl, err := tpl.Parse(fmt.Sprintf("{{ %s . }}", fn))
	if err != nil {
		return nil, err
	}

	err = tpl.Execute(buf, input)
	return buf.String(), err
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
