package handler

import (
	"github.com/h0n9/toybox/cloud-secrets-injector/provider"
	"github.com/h0n9/toybox/cloud-secrets-injector/util"
)

type SecretHandlerFunc func(string) (string, error)

type SecretHandler struct {
	provider provider.Provider
}

func NewSecretHandler(provider provider.Provider) *SecretHandler {
	return &SecretHandler{provider: provider}
}

func (handler *SecretHandler) Get(secretId string) (string, error) {
	return handler.provider.GetSecretValue(secretId)
}

func (handler *SecretHandler) Save(secretId, path string) error {
	secretValue, err := handler.Get(secretId)
	if err != nil {
		return err
	}
	return util.SaveStringToFile(secretValue, path)
}
