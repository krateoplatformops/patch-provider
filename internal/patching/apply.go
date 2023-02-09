package patching

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Apply changes to the supplied object. The object will be created if it does
// not exist, or patched if it does. If the object does exist, it will only be
// patched if the passed object has the same or an empty resource version.
func Apply(ctx context.Context, kc client.Client, o client.Object) error {
	m, ok := o.(metav1.Object)
	if !ok {
		return errors.New("cannot access object metadata")
	}

	if m.GetName() == "" && m.GetGenerateName() != "" {
		return errors.Wrap(kc.Create(ctx, o), "cannot create object")
	}

	desired := o.DeepCopyObject()

	err := kc.Get(ctx, types.NamespacedName{Name: m.GetName(), Namespace: m.GetNamespace()}, o)
	if kerrors.IsNotFound(err) {
		return errors.Wrap(kc.Create(ctx, o), "cannot create object")
	}
	if err != nil {
		return errors.Wrap(err, "cannot get object")
	}

	return errors.Wrap(kc.Patch(ctx, o, &patch{desired}), "cannot patch object")
}

type patch struct{ from runtime.Object }

func (p *patch) Type() types.PatchType                { return types.MergePatchType }
func (p *patch) Data(_ client.Object) ([]byte, error) { return json.Marshal(p.from) }
