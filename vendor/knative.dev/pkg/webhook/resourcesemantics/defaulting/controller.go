/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package defaulting

import (
	"context"

	// Injection stuff
	kubeclient "knative.dev/pkg/client/injection/kube/client"
	mwhinformer "knative.dev/pkg/client/injection/kube/informers/admissionregistration/v1/mutatingwebhookconfiguration"
	secretinformer "knative.dev/pkg/injection/clients/namespacedkube/informers/core/v1/secret"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/controller"
	"knative.dev/pkg/system"
	"knative.dev/pkg/webhook"
	"knative.dev/pkg/webhook/resourcesemantics"
)

// NewAdmissionController constructs a reconciler
func NewAdmissionController(
	ctx context.Context,
	name, path string,
	handlers map[schema.GroupVersionKind]resourcesemantics.GenericCRD,
	wc func(context.Context) context.Context,
	disallowUnknownFields bool,
	callbacks ...map[schema.GroupVersionKind]Callback,
) *controller.Impl {
	// This not ideal, we are using a variadic argument to effectively make callbacks optional
	// This allows this addition to be non-breaking to consumers of /pkg
	// TODO: once all sub-repos have adopted this, we might move this back to a traditional param.
	var unwrappedCallbacks map[schema.GroupVersionKind]Callback
	switch len(callbacks) {
	case 0:
		unwrappedCallbacks = map[schema.GroupVersionKind]Callback{}
	case 1:
		unwrappedCallbacks = callbacks[0]
	default:
		panic("NewAdmissionController may not be called with multiple callback maps")
	}

	opts := []OptionFunc{
		WithPath(path),
		WithTypes(handlers),
		WithWrapContext(wc),
		WithCallbacks(unwrappedCallbacks),
	}

	if disallowUnknownFields {
		opts = append(opts, WithDisallowUnknownFields())
	}

	return newController(ctx, name, opts...)
}

func newController(ctx context.Context, name string, optsFunc ...OptionFunc) *controller.Impl {
	client := kubeclient.Get(ctx)
	mwhInformer := mwhinformer.Get(ctx)
	secretInformer := secretinformer.Get(ctx)

	opts := &options{}
	wopts := webhook.GetOptions(ctx)

	for _, f := range optsFunc {
		f(opts)
	}

	// if this environment variable is set, it overrides the value in the Options
	disableNamespaceOwnership := webhook.DisableNamespaceOwnershipFromEnv()
	if disableNamespaceOwnership != nil {
		wopts.DisableNamespaceOwnership = *disableNamespaceOwnership
	}

	key := types.NamespacedName{Name: name}

	wh := &reconciler{
		LeaderAwareFuncs: pkgreconciler.LeaderAwareFuncs{
			// Have this reconciler enqueue our singleton whenever it becomes leader.
			PromoteFunc: func(bkt pkgreconciler.Bucket, enq func(pkgreconciler.Bucket, types.NamespacedName)) error {
				enq(bkt, key)
				return nil
			},
		},

		key:       key,
		path:      opts.path,
		handlers:  opts.types,
		callbacks: opts.callbacks,

		withContext:               opts.wc,
		disallowUnknownFields:     opts.disallowUnknownFields,
		secretName:                wopts.SecretName,
		disableNamespaceOwnership: wopts.DisableNamespaceOwnership,

		client:       client,
		mwhlister:    mwhInformer.Lister(),
		secretlister: secretInformer.Lister(),
	}

	logger := logging.FromContext(ctx)
	controllerOptions := wopts.ControllerOptions
	if controllerOptions == nil {
		const queueName = "DefaultingWebhook"
		controllerOptions = &controller.ControllerOptions{WorkQueueName: queueName, Logger: logger.Named(queueName)}
	}
	c := controller.NewContext(ctx, wh, *controllerOptions)

	// Reconcile when the named MutatingWebhookConfiguration changes.
	mwhInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithName(name),
		// It doesn't matter what we enqueue because we will always Reconcile
		// the named MWH resource.
		Handler: controller.HandleAll(c.Enqueue),
	})

	// Reconcile when the cert bundle changes.
	secretInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterWithNameAndNamespace(system.Namespace(), wh.secretName),
		// It doesn't matter what we enqueue because we will always Reconcile
		// the named MWH resource.
		Handler: controller.HandleAll(c.Enqueue),
	})

	return c
}
