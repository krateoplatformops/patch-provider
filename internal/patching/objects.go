package patching

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/tmpl"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func From(ctx context.Context, kc client.Client, cr *v1alpha1.Patch) (*unstructured.Unstructured, error) {
	if cr.Spec.From == nil || cr.Spec.From.ObjectReference == nil {
		return nil, errors.New("from.objectReference is required")
	}

	return resolveObjectReference(ctx, kc, cr.Spec.From.ObjectReference)
}

func To(ctx context.Context, kc client.Client, cr *v1alpha1.Patch) (*unstructured.Unstructured, error) {
	if cr.Spec.To == nil || cr.Spec.To.ObjectReference == nil {
		return nil, errors.New("to.objectReference is required")
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

// func TransformEventually(cr *v1alpha1.Patch, input any) (any, error) {
// 	if len(cr.Spec.To.Transforms) == 0 {
// 		return input, nil
// 	}

// 	fn := strings.Join(cr.Spec.To.Transforms, " | ")

// 	buf := bytes.NewBufferString("")
// 	tpl := template.New(cr.GetName()).Funcs(functions.Map())
// 	tpl, err := tpl.Parse(fmt.Sprintf("{{ %s }}", fn))
// 	if err != nil {
// 		return nil, err
// 	}

// 	err = tpl.Execute(buf, input)
// 	return buf.String(), err
// }

func TransformEventually(cr *v1alpha1.Patch, input any) (any, error) {
	if len(cr.Spec.To.Transform) == 0 {
		return input, nil
	}

	tpl, err := tmpl.New("${", "}")
	if err != nil {
		return nil, err
	}

	mapInput := make(map[string]any)

	binput, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(binput, &mapInput)
	if err != nil {
		return nil, err
	}

	result, err := tpl.Execute(cr.Spec.To.Transform, mapInput)

	return result, err
}
