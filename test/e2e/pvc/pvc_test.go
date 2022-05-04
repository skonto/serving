//go:build e2e
// +build e2e

/*
Copyright 2022 The Knative Authors

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

package pvc

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/ptr"
	pkgTest "knative.dev/pkg/test"
	"knative.dev/pkg/test/spoof"
	"knative.dev/serving/pkg/apis/autoscaling"
	v1 "knative.dev/serving/pkg/apis/serving/v1"
	. "knative.dev/serving/pkg/testing/v1"
	"knative.dev/serving/test"
	v1test "knative.dev/serving/test/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"knative.dev/serving/pkg/resources"
)

const (
	unprivilegedUserID = 65532
)

// TestPersistentVolumeClaims tests pvc support.
func TestPersistentVolumeClaims(t *testing.T) {
	if !test.ServingFlags.EnableAlphaFeatures {
		t.Skip("Alpha features not enabled")
	}
	t.Parallel()
	clients := test.Setup(t)

	names := test.ResourceNames{
		Service: test.ObjectNameForTest(t),
		Image:   test.Volumes,
	}

	test.EnsureTearDown(t, clients, &names)

	t.Log("Creating a new Service")

	withVolume := WithVolume("data", "/data", corev1.VolumeSource{
		PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
			ClaimName: "knative-pv-claim",
			ReadOnly:  false,
		},
	})

	// make sure default user can access the written file
	withPodSecurityContext := WithPodSecurityContext(corev1.PodSecurityContext{
		FSGroup: ptr.Int64(unprivilegedUserID),
	})

	resources, err := v1test.CreateServiceReady(t, clients, &names, withVolume, withPodSecurityContext, withMinScale(3))
	if err != nil {
		t.Fatalf("Failed to create initial Service: %v: %v", names.Service, err)
	}

	t.Log("Holding service at minScale after becoming ready")
	if lr, ok := ensureDesiredScale(clients, t, names.Service, gte(3)); !ok {
		t.Fatalf("The service %q observed scale %d < %d after becoming ready", names.Service, lr, 3)
	}

	url := resources.Route.Status.URL.URL()
	if _, err := pkgTest.CheckEndpointState(
		context.Background(),
		clients.KubeClient,
		t.Logf,
		url,
		spoof.MatchesAllOf(spoof.IsStatusOK, spoof.MatchesBody(test.EmptyDirText)),
		"PVCText",
		test.ServingFlags.ResolvableDomain,
		test.AddRootCAtoTransport(context.Background(), t.Logf, clients, test.ServingFlags.HTTPS),
	); err != nil {
		t.Fatalf("The endpoint %s for Route %s didn't serve the expected text %q: %v", url, names.Route, test.EmptyDirText, err)
	}
}

func withMinScale(minScale int) func(cfg *v1.Service) {
	return func(svc *v1.Service) {
		if svc.Spec.Template.Annotations == nil {
			svc.Spec.Template.Annotations = make(map[string]string, 1)
		}
		svc.Spec.Template.Annotations[autoscaling.MinScaleAnnotationKey] = strconv.Itoa(minScale)
	}
}


func gte(m int) func(int) bool {
	return func(n int) bool {
		return n >= m
	}
}

func ensureDesiredScale(clients *test.Clients, t *testing.T, serviceName string, cond func(int) bool) (latestReady int, observed bool) {
	endpoints := clients.KubeClient.CoreV1().Endpoints(test.ServingFlags.TestNamespace)

	err := wait.PollImmediate(250*time.Millisecond, 300*time.Second, func() (bool, error) {
		endpoint, err := endpoints.Get(context.Background(), serviceName, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		if latestReady = resources.ReadyAddressCount(endpoint); !cond(latestReady) {
			return false, fmt.Errorf("scale %d didn't meet condition", latestReady)
		}

		return false, nil
	})
	if !errors.Is(err, wait.ErrWaitTimeout) {
		t.Log("PollError =", err)
	}

	return latestReady, errors.Is(err, wait.ErrWaitTimeout)
}