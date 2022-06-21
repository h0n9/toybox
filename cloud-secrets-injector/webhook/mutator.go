package webhook

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Mutator struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (mutator *Mutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	return admission.Response{}
}

func (mutator *Mutator) InjectDecoder(decoder *admission.Decoder) error {
	mutator.decoder = decoder
	return nil
}
