package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

func (c *Client) GetDeployments(namespace string) ([]appsv1.Deployment, error) {
	dps, err := c.clientSet.AppsV1().Deployments(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return dps.Items, nil
}

func (c *Client) RestartDeployment(dp *appsv1.Deployment) (*appsv1.Deployment, error) {
	before, err := runtime.Encode(scheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), dp)
	if err != nil {
		return nil, err
	}
	after, err := polymorphichelpers.ObjectRestarterFn(dp)
	if err != nil {
		return nil, err
	}
	diff, err := strategicpatch.CreateTwoWayMergePatch(before, after, dp)
	if err != nil {
		return nil, err
	}
	return c.clientSet.AppsV1().Deployments(dp.Namespace).Patch(
		c.ctx,
		dp.Name,
		types.StrategicMergePatchType,
		diff,
		metav1.PatchOptions{},
	)
}
