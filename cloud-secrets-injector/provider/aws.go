package provider

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"

	"github.com/h0n9/toybox/cloud-secrets-injector/util"
)

type AWS struct {
	ctx    context.Context
	cfg    config.Config
	client *secretsmanager.Client
}

func NewAWS(ctx context.Context) (*AWS, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	client := secretsmanager.NewFromConfig(cfg)
	return &AWS{
		ctx:    ctx,
		cfg:    cfg,
		client: client,
	}, nil
}

func (provider *AWS) GetSecretValue(secretId string) (string, error) {
	secret, err := provider.client.GetSecretValue(provider.ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretId,
	})
	if err != nil {
		return "", err
	}
	return *secret.SecretString, nil
}

func (provider *AWS) GetAndSaveSecretValueToFile(secretId, path string) (string, error) {
	secretString, err := provider.GetSecretValue(secretId)
	if err != nil {
		return "", err
	}
	err = util.SaveStringToFile(path, secretString)
	if err != nil {
		return "", err
	}
	return secretString, nil
}
