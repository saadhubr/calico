// Copyright (c) 2025 Tigera, Inc. All rights reserved.

// Code generated by client-gen. DO NOT EDIT.

package v3

import (
	"context"

	v3 "github.com/projectcalico/api/pkg/apis/projectcalico/v3"
	scheme "github.com/projectcalico/api/pkg/client/clientset_generated/clientset/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	gentype "k8s.io/client-go/gentype"
)

// GlobalNetworkPoliciesGetter has a method to return a GlobalNetworkPolicyInterface.
// A group's client should implement this interface.
type GlobalNetworkPoliciesGetter interface {
	GlobalNetworkPolicies() GlobalNetworkPolicyInterface
}

// GlobalNetworkPolicyInterface has methods to work with GlobalNetworkPolicy resources.
type GlobalNetworkPolicyInterface interface {
	Create(ctx context.Context, globalNetworkPolicy *v3.GlobalNetworkPolicy, opts v1.CreateOptions) (*v3.GlobalNetworkPolicy, error)
	Update(ctx context.Context, globalNetworkPolicy *v3.GlobalNetworkPolicy, opts v1.UpdateOptions) (*v3.GlobalNetworkPolicy, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v3.GlobalNetworkPolicy, error)
	List(ctx context.Context, opts v1.ListOptions) (*v3.GlobalNetworkPolicyList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v3.GlobalNetworkPolicy, err error)
	GlobalNetworkPolicyExpansion
}

// globalNetworkPolicies implements GlobalNetworkPolicyInterface
type globalNetworkPolicies struct {
	*gentype.ClientWithList[*v3.GlobalNetworkPolicy, *v3.GlobalNetworkPolicyList]
}

// newGlobalNetworkPolicies returns a GlobalNetworkPolicies
func newGlobalNetworkPolicies(c *ProjectcalicoV3Client) *globalNetworkPolicies {
	return &globalNetworkPolicies{
		gentype.NewClientWithList[*v3.GlobalNetworkPolicy, *v3.GlobalNetworkPolicyList](
			"globalnetworkpolicies",
			c.RESTClient(),
			scheme.ParameterCodec,
			"",
			func() *v3.GlobalNetworkPolicy { return &v3.GlobalNetworkPolicy{} },
			func() *v3.GlobalNetworkPolicyList { return &v3.GlobalNetworkPolicyList{} }),
	}
}
