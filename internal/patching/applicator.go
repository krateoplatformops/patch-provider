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

// An Applicator applies changes to an object by either creating or
// patching it in a Kubernetes API server.
type Applicator struct {
	client client.Client
}

// NewApplicator returns an Applicator that applies changes to an
// object by either creating or patching it in a Kubernetes API server.
func NewApplicator(c client.Client) *Applicator {
	return &Applicator{client: c}
}

// Apply changes to the supplied object. The object will be created if it does
// not exist, or patched if it does. If the object does exist, it will only be
// patched if the passed object has the same or an empty resource version.
func (a *Applicator) Apply(ctx context.Context, o client.Object) error {
	m, ok := o.(metav1.Object)
	if !ok {
		return errors.New("cannot access object metadata")
	}

	if m.GetName() == "" && m.GetGenerateName() != "" {
		return errors.Wrap(a.client.Create(ctx, o), "cannot create object")
	}

	desired := o.DeepCopyObject()

	err := a.client.Get(ctx, types.NamespacedName{Name: m.GetName(), Namespace: m.GetNamespace()}, o)
	if kerrors.IsNotFound(err) {
		return errors.Wrap(a.client.Create(ctx, o), "cannot create object")
	}
	if err != nil {
		return errors.Wrap(err, "cannot get object")
	}

	return errors.Wrap(a.client.Patch(ctx, o, &patch{desired}), "cannot patch object")
}

type patch struct{ from runtime.Object }

func (p *patch) Type() types.PatchType                { return types.MergePatchType }
func (p *patch) Data(_ client.Object) ([]byte, error) { return json.Marshal(p.from) }
