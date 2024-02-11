package factory

import (
	context "context"

	"github.com/kedacore/keda/v2/pkg/generated/clientset/versioned"
	informers "github.com/kedacore/keda/v2/pkg/generated/informers/externalversions"
	controller "knative.dev/pkg/controller"
	injection "knative.dev/pkg/injection"
	logging "knative.dev/pkg/logging"
)

func init() {
	injection.Default.RegisterInformerFactory(withInformerFactory)
}

// Key is used as the key for associating information with a context.Context.
type Key struct{}

func withInformerFactory(ctx context.Context) context.Context {
	opts := make([]informers.SharedInformerOption, 0, 1)
	if injection.HasNamespaceScope(ctx) {
		opts = append(opts, informers.WithNamespace(injection.GetNamespaceScope(ctx)))
	}
	vc, _ := versioned.NewForConfig(injection.GetConfig(ctx))
	return context.WithValue(ctx, Key{},
		informers.NewSharedInformerFactoryWithOptions(vc, controller.GetResyncPeriod(ctx), opts...))
}

// Get extracts the InformerFactory from the context.
func Get(ctx context.Context) informers.SharedInformerFactory {
	untyped := ctx.Value(Key{})
	if untyped == nil {
		logging.FromContext(ctx).Panic(
			"Unable to fetch k8s.io/client-go/informers.SharedInformerFactory from context.")
	}
	return untyped.(informers.SharedInformerFactory)
}
