package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (c *Client) GetStatefulSet(namespace string) ([]appsv1.StatefulSet, error) {
	sts, err := c.clientSet.AppsV1().StatefulSets(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return sts.Items, err
}

func (c *Client) RestartStatefulSet(ss *appsv1.StatefulSet) (*appsv1.StatefulSet, error) {
	diff, err := GenerateDiff(ss)
	if err != nil {
		return nil, err
	}
	return c.clientSet.AppsV1().StatefulSets(ss.Namespace).Patch(
		c.ctx,
		ss.Name,
		types.StrategicMergePatchType,
		diff,
		metav1.PatchOptions{},
	)
}
