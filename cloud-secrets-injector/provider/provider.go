package provider

type SecretHandler func(string) (string, error)

type Provider interface {
	GetSecretValue(string) (string, error)
	GetAndSaveSecretValueToFile(string, string) error
	GetAndHandleSecretValue(string, SecretHandler) error
}
