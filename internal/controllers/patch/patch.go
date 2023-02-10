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
	"github.com/krateoplatformops/provider-runtime/pkg/logging"
	"github.com/krateoplatformops/provider-runtime/pkg/ratelimiter"
	"github.com/krateoplatformops/provider-runtime/pkg/reconciler/managed"
	"github.com/krateoplatformops/provider-runtime/pkg/resource"

	"github.com/krateoplatformops/patch-provider/apis/patch/v1alpha1"
	"github.com/krateoplatformops/patch-provider/internal/fieldpath"
	"github.com/krateoplatformops/patch-provider/internal/patching"
)

const (
	errNotPatch = "managed resource is not a Patch custom resource"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.PatchGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.PatchGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:     mgr.GetClient(),
			log:      log,
			recorder: recorder,
		}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(recorder)))

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

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
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

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Patch)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPatch)
	}

	in, err := patching.FromFieldPathValue(ctx, e.kube, cr)
	if err != nil {
		return managed.ExternalObservation{}, resource.Ignore(fieldpath.IsNotFound, err)
	}

	out, err := patching.ToFieldPathValue(ctx, e.kube, cr)
	if err != nil {
		return managed.ExternalObservation{}, resource.Ignore(fieldpath.IsNotFound, err)
	}

	diff := cmp.Diff(in, out)
	e.log.Debug("Computed drift", "diff", diff)
	if len(diff) == 0 {
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: false,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) error {
	/*
		cr, ok := mg.(*v1alpha1.Patch)
		if !ok {
			return errors.New(errNotPatch)
		}

		cr.SetConditions(rtv1.Creating())

		spec := cr.Spec.DeepCopy()

		e.log.Debug("Repo created", "org", spec.Org, "name", spec.Name)
		e.rec.Eventf(cr, corev1.EventTypeNormal, "RepoCreated", "Repo '%s/%s' created", spec.Org, spec.Name)
	*/
	return nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) error {
	return nil // noop
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
