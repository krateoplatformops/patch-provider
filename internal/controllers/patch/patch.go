package patch

import (
	"context"
	"errors"

	"github.com/google/go-cmp/cmp"
	rtv1 "github.com/krateoplatformops/provider-runtime/apis/common/v1"
	"k8s.io/client-go/tools/record"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/krateoplatformops/provider-runtime/pkg/controller"
	"github.com/krateoplatformops/provider-runtime/pkg/event"
	"github.com/krateoplatformops/provider-runtime/pkg/helpers"
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"github.com/krateoplatformops/patch-provider/internal/patching"
)

const (
	errNotPatch = "reconciler resource is not a Patch custom resource"
)

// Setup adds a controller that reconciles Token reconciler resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := reconciler.ControllerName(v1alpha1.PatchGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := reconciler.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PatchGroupVersionKind),
		reconciler.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		reconciler.WithPollInterval(o.PollInterval),
		reconciler.WithLogger(log),
		reconciler.WithRecorder(event.NewAPIRecorder(recorder)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Patch{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube     client.Client
	log      logging.Logger
	recorder record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (reconciler.ExternalClient, error) {
	return &external{
		kube: c.kube,
		log:  c.log,
		rec:  c.recorder,
	}, nil
}

type external struct {
	kube client.Client
	log  logging.Logger
	rec  record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (reconciler.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Patch)
	if !ok {
		return reconciler.ExternalObservation{}, errors.New(errNotPatch)
	}

	if cr.GetDeletionTimestamp() != nil && cr.GetCondition(rtv1.TypeReady).Reason == rtv1.ReasonDeleting {
		return reconciler.ExternalObservation{
			ResourceExists:   false,
			ResourceUpToDate: true,
		}, nil
	}

	if cr.Spec.From.FieldPath == nil {
		return reconciler.ExternalObservation{}, errors.New("from.fieldPath is required")
	}

	from, err := patching.From(ctx, e.kube, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, resource.IgnoreNotFound(err)
	}

	to, err := patching.To(ctx, e.kube, cr)
	if err != nil {
		return reconciler.ExternalObservation{}, resource.IgnoreNotFound(err)
	}

	in, err := fieldpath.Pave(from.Object).GetValue(helpers.String(cr.Spec.From.FieldPath))
	if err != nil {
		return reconciler.ExternalObservation{}, resource.Ignore(fieldpath.IsNotFound, err)
	}

	if in, err = patching.TransformEventually(cr, in); err != nil {
		return reconciler.ExternalObservation{}, err
	}

	// if 'to' fieldPath is not specified, use the same 'from' fieldPath.
	toFieldPath := helpers.StringOrDefault(cr.Spec.To.FieldPath, helpers.String(cr.Spec.To.FieldPath))
	out, err := fieldpath.Pave(to.Object).GetValue(toFieldPath)
	if err != nil {
		return reconciler.ExternalObservation{}, resource.Ignore(fieldpath.IsNotFound, err)
	}

	diff := cmp.Diff(in, out)
	if len(diff) == 0 {
		cr.SetConditions(rtv1.Available())
		return reconciler.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	return reconciler.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	return nil // noop
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Patch)
	if !ok {
		return errors.New(errNotPatch)
	}

	obj, err := patching.Patch(ctx, e.kube, cr)
	if err != nil {
		return err
	}
	cr.SetConditions(rtv1.Creating())

	return patching.Apply(ctx, e.kube, obj)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Patch)
	if !ok {
		return errors.New(errNotPatch)
	}

	cr.SetConditions(rtv1.Deleting())

	//e.log.Debug("Repo deleted", "org", spec.Org, "name", spec.Name)
	//e.rec.Eventf(cr, corev1.EventTypeNormal, "RepDeleted", "Repo '%s/%s' deleted", spec.Org, spec.Name)

	return nil
}
