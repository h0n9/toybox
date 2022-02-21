package main

import (
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
)

func GenerateDiff(obj runtime.Object) ([]byte, error) {
	before, err := runtime.Encode(scheme.Codecs.LegacyCodec(appsv1.SchemeGroupVersion), obj)
	if err != nil {
		return nil, err
	}
	after, err := polymorphichelpers.ObjectRestarterFn(obj)
	if err != nil {
		return nil, err
	}
	return strategicpatch.CreateTwoWayMergePatch(before, after, obj)
}
