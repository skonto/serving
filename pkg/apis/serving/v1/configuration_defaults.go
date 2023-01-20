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
	"fmt"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/equality"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/kmp"
	"knative.dev/pkg/logging"
	"knative.dev/serving/pkg/apis/serving"
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
	var existing *Configuration
	if prev, ok := apis.GetBaseline(ctx).(*Configuration); ok && prev != nil {
		prevSpec = &prev.Spec
		existing = prev
		ctx = WithPreviousConfigurationSpec(ctx, prevSpec)
	}

	c.Spec.SetDefaults(apis.WithinSpec(ctx))

	if c.GetOwnerReferences() == nil {
		serving.SetUserInfo(ctx, prevSpec, &c.Spec, c)
	}

	if existing != nil {
		configSemanticEquals(ctx, c, existing)
	}
}
func configSemanticEquals(ctx context.Context, desiredConfig, config *Configuration) (bool, error) {
	logger := logging.FromContext(ctx)
	specDiff, err := kmp.SafeDiff(desiredConfig.Spec, config.Spec)
	if err != nil {
		logger.Warnw("Error diffing config spec", zap.Error(err))
		return false, fmt.Errorf("failed to diff Configuration: %w", err)
	} else if specDiff != "" {
		logger.Info("Reconciling configuration difff (-desired, +observed):\n", specDiff)
	}
	return equality.Semantic.DeepEqual(desiredConfig.Spec, config.Spec) &&
		equality.Semantic.DeepEqual(desiredConfig.Labels, config.Labels) &&
		equality.Semantic.DeepEqual(desiredConfig.Annotations, config.Annotations) &&
		specDiff == "", nil
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
	cs.Template.SetDefaults(ctx)
}
