package main

import (
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
)

func (c *Client) GetDeployments(namespace string) ([]appsv1.Deployment, error) {
	dps, err := c.clientSet.AppsV1().Deployments(namespace).List(c.ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return dps.Items, nil
}

func (c *Client) RestartDeployment(dp *appsv1.Deployment) (*appsv1.Deployment, error) {
	diff, err := GenerateDiff(dp)
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

func (c *Client) WaitDeployment(dp *appsv1.Deployment) error {
	w, err := c.clientSet.AppsV1().Deployments(dp.Namespace).Watch(c.ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	stop := false
	for !stop {
		result := <-w.ResultChan()
		rdp := result.Object.(*appsv1.Deployment)
		if rdp == nil {
			continue
		}
		if !(rdp.Name == dp.Name && rdp.Namespace == dp.Namespace) {
			continue
		}
		if result.Type == watch.Modified {
			stop = true
		}
	}
	w.Stop()
	return nil
}
