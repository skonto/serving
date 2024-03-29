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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	scheme "knative.dev/serving/pkg/netcertmanager/client/certmanager/clientset/versioned/scheme"
)

// CertificateRequestsGetter has a method to return a CertificateRequestInterface.
// A group's client should implement this interface.
type CertificateRequestsGetter interface {
	CertificateRequests(namespace string) CertificateRequestInterface
}

// CertificateRequestInterface has methods to work with CertificateRequest resources.
type CertificateRequestInterface interface {
	Create(ctx context.Context, certificateRequest *v1.CertificateRequest, opts metav1.CreateOptions) (*v1.CertificateRequest, error)
	Update(ctx context.Context, certificateRequest *v1.CertificateRequest, opts metav1.UpdateOptions) (*v1.CertificateRequest, error)
	UpdateStatus(ctx context.Context, certificateRequest *v1.CertificateRequest, opts metav1.UpdateOptions) (*v1.CertificateRequest, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.CertificateRequest, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.CertificateRequestList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.CertificateRequest, err error)
	CertificateRequestExpansion
}

// certificateRequests implements CertificateRequestInterface
type certificateRequests struct {
	client rest.Interface
	ns     string
}

// newCertificateRequests returns a CertificateRequests
func newCertificateRequests(c *CertmanagerV1Client, namespace string) *certificateRequests {
	return &certificateRequests{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the certificateRequest, and returns the corresponding certificateRequest object, and an error if there is any.
func (c *certificateRequests) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.CertificateRequest, err error) {
	result = &v1.CertificateRequest{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("certificaterequests").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CertificateRequests that match those selectors.
func (c *certificateRequests) List(ctx context.Context, opts metav1.ListOptions) (result *v1.CertificateRequestList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.CertificateRequestList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("certificaterequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested certificateRequests.
func (c *certificateRequests) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("certificaterequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a certificateRequest and creates it.  Returns the server's representation of the certificateRequest, and an error, if there is any.
func (c *certificateRequests) Create(ctx context.Context, certificateRequest *v1.CertificateRequest, opts metav1.CreateOptions) (result *v1.CertificateRequest, err error) {
	result = &v1.CertificateRequest{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("certificaterequests").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(certificateRequest).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a certificateRequest and updates it. Returns the server's representation of the certificateRequest, and an error, if there is any.
func (c *certificateRequests) Update(ctx context.Context, certificateRequest *v1.CertificateRequest, opts metav1.UpdateOptions) (result *v1.CertificateRequest, err error) {
	result = &v1.CertificateRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("certificaterequests").
		Name(certificateRequest.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(certificateRequest).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *certificateRequests) UpdateStatus(ctx context.Context, certificateRequest *v1.CertificateRequest, opts metav1.UpdateOptions) (result *v1.CertificateRequest, err error) {
	result = &v1.CertificateRequest{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("certificaterequests").
		Name(certificateRequest.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(certificateRequest).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the certificateRequest and deletes it. Returns an error if one occurs.
func (c *certificateRequests) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("certificaterequests").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *certificateRequests) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("certificaterequests").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched certificateRequest.
func (c *certificateRequests) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.CertificateRequest, err error) {
	result = &v1.CertificateRequest{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("certificaterequests").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
