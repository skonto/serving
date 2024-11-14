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

package v1

import (
	"context"
	"log"

	"knative.dev/pkg/apis"
	// "knative.dev/serving/pkg/apis/config"
	"knative.dev/serving/pkg/apis/serving"
	// cconfig "knative.dev/serving/pkg/reconciler/configuration/config"
)

type configSpecKey struct{}

// WithPreviousConfigurationSpec stores the pre-update ConfigurationSpec in the
// context, to allow ConfigurationSpec.SetDefaults to determine whether the
// update would create a new Revision.
func WithPreviousConfigurationSpec(ctx context.Context, spec *ConfigurationSpec) context.Context {
	return context.WithValue(ctx, configSpecKey{}, spec)
}

func previousConfigSpec(ctx context.Context) *ConfigurationSpec {
	if spec, ok := ctx.Value(configSpecKey{}).(*ConfigurationSpec); ok {
		return spec
	}
	return nil
}

// SetDefaults implements apis.Defaultable
func (c *Configuration) SetDefaults(ctx context.Context) {
	ctx = apis.WithinParent(ctx, c.ObjectMeta)

	var prevSpec *ConfigurationSpec
	if prev, ok := apis.GetBaseline(ctx).(*Configuration); ok && prev != nil {
		prevSpec = &prev.Spec
		ctx = WithPreviousConfigurationSpec(ctx, prevSpec)
	}
	if c.Spec.Template.Spec.ResponseStartTimeoutSeconds != nil {
		log.Printf("CONFIG_DEFAULTS1: %v\n", *c.Spec.Template.Spec.ResponseStartTimeoutSeconds)
	}
	log.Printf("CONFIG_DEFAULTS2:\n")
	c.Spec.SetDefaults(apis.WithinSpec(ctx))
	if c.Spec.Template.Spec.ResponseStartTimeoutSeconds != nil {
		log.Printf("CONFIG_DEFAULTS3: %v\n", *c.Spec.Template.Spec.ResponseStartTimeoutSeconds)
	}
	log.Printf("CONFIG_DEFAULTS4:\n")

	if c.GetOwnerReferences() == nil {
		serving.SetUserInfo(ctx, prevSpec, &c.Spec, c)
	}
}

// SetDefaults implements apis.Defaultable
func (cs *ConfigurationSpec) SetDefaults(ctx context.Context) {
	if prev := previousConfigSpec(ctx); prev != nil {
		newName := cs.Template.ObjectMeta.Name
		oldName := prev.Template.ObjectMeta.Name
		if newName != "" && newName == oldName {
			// Skip defaulting, to avoid suggesting changes that would conflict with
			// "BYO RevisionName".
			return
		}
	}

	//cfg := cconfig.FromContext(ctx)
	//cf := config.Config{}
	//if cfg != nil && cfg.Defaults != nil {
	//	cf.Defaults = cfg.Defaults.DeepCopy()
	//}
	//if cfg != nil && cfg.Features != nil {
	//	cf.Features = cfg.Features.DeepCopy()
	//}
	//if cfg != nil {
	//	ctx = config.ToContext(ctx, &cf)
	//}

	cs.Template.SetDefaults(ctx)
}
