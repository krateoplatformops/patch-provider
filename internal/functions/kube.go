package functions

import (
	"context"

	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func fetchKnown(kc client.Client, gv, k string) func(ns, n string, path string) (any, error) {
	return func(ns, n string, path string) (any, error) {
		return fetchUnknown(kc)(gv, k, ns, n, path)
	}
}

func fetchUnknown(kc client.Client) func(gv, k string, ns, n string, path string) (any, error) {
	return func(gv, k string, ns, n string, path string) (any, error) {
		if ns == "" {
			ns = "default"
		}

		u := unstructured.Unstructured{}
		u.SetGroupVersionKind(gvk(gv, k))

		err := kc.Get(context.TODO(), types.NamespacedName{
			Namespace: ns,
			Name:      n,
		}, &u)
		if err != nil {
			return nil, err
		}

		if len(path) == 0 {
			return u.Object, nil
		}

		in, err := fieldpath.Pave(u.Object).GetValue(path)
		if err != nil {
			return nil, err
		}
		return in, nil
	}
}

func gvk(groupVersion, kind string) schema.GroupVersionKind {
	gv, err := schema.ParseGroupVersion(groupVersion)
	if err != nil {
		panic(err)
	}

	return gv.WithKind(kind)
}
