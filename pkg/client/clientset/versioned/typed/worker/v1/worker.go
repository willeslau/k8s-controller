/*
Copyright The Kubernetes Authors.

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

	v1 "github.com/willeslau/k8s-controller/pkg/apis/worker/v1"
	scheme "github.com/willeslau/k8s-controller/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// WorkersGetter has a method to return a WorkerInterface.
// A group's client should implement this interface.
type WorkersGetter interface {
	Workers(namespace string) WorkerInterface
}

// WorkerInterface has methods to work with Worker resources.
type WorkerInterface interface {
	Create(ctx context.Context, worker *v1.Worker, opts metav1.CreateOptions) (*v1.Worker, error)
	Update(ctx context.Context, worker *v1.Worker, opts metav1.UpdateOptions) (*v1.Worker, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.Worker, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.WorkerList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Worker, err error)
	WorkerExpansion
}

// workers implements WorkerInterface
type workers struct {
	client rest.Interface
	ns     string
}

// newWorkers returns a Workers
func newWorkers(c *WillesxmV1Client, namespace string) *workers {
	return &workers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the worker, and returns the corresponding worker object, and an error if there is any.
func (c *workers) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.Worker, err error) {
	result = &v1.Worker{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("workers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Workers that match those selectors.
func (c *workers) List(ctx context.Context, opts metav1.ListOptions) (result *v1.WorkerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.WorkerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("workers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested workers.
func (c *workers) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("workers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a worker and creates it.  Returns the server's representation of the worker, and an error, if there is any.
func (c *workers) Create(ctx context.Context, worker *v1.Worker, opts metav1.CreateOptions) (result *v1.Worker, err error) {
	result = &v1.Worker{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("workers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(worker).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a worker and updates it. Returns the server's representation of the worker, and an error, if there is any.
func (c *workers) Update(ctx context.Context, worker *v1.Worker, opts metav1.UpdateOptions) (result *v1.Worker, err error) {
	result = &v1.Worker{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("workers").
		Name(worker.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(worker).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the worker and deletes it. Returns an error if one occurs.
func (c *workers) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("workers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *workers) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("workers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched worker.
func (c *workers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.Worker, err error) {
	result = &v1.Worker{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("workers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}