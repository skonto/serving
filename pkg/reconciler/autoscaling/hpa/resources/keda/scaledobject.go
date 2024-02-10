package keda

import (
	context "context"

	v1alpha1 "github.com/kedacore/keda/v2/pkg/generated/informers/externalversions/keda/v1alpha1"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
	factory "knative.dev/serving/pkg/reconciler/autoscaling/hpa/resources/keda/factory"
)

func init() {
	injection.Default.RegisterInformer(withInformer)
}

// Key is used for associating the Informer inside the context.Context.
type Key struct{}

func withInformer(ctx context.Context) (context.Context, controller.Informer) {
	f := factory.Get(ctx)
	inf := f.Keda().V1alpha1().ScaledObjects()
	return context.WithValue(ctx, Key{}, inf), inf.Informer()
}

// Get extracts the typed informer from the context.
func Get(ctx context.Context) v1alpha1.ScaledObjectInformer {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch KEDA v1alpha1.ScaledObjectInformer from context.")
	}
	return untyped.(v1alpha1.ScaledObjectInformer)
}
