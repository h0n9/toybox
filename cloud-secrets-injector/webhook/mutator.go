package webhook

import (
	"context"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Mutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (mutator *Mutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := mutator.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	annotations := pod.Annotations
	if len(annotations) == 0 {
		return admission.Allowed("found no cloud-secrets-injector related annotations")
	}

	return admission.Response{}
}

func (mutator *Mutator) InjectDecoder(decoder *admission.Decoder) error {
	mutator.decoder = decoder
	return nil
}
